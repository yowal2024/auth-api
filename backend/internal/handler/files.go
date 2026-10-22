package handler

import (
	"auth-api/internal/config"
	"auth-api/internal/middleware"
	"auth-api/internal/repo"
	"auth-api/internal/response"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type FileService struct {
	DB            *sqlx.DB
	Root, URL     string
	CanonicalRoot string
}

func NewFileService(d *sqlx.DB, c config.UploadConfig, clone ...config.CloneConfig) *FileService {
	root := c.Root
	if root == "" {
		root = "./uploads"
	}
	url := c.URLPrefix
	if url == "" {
		url = "/uploads"
	}
	canonical := filepath.Join(root, "canonical")
	if len(clone) > 0 && clone[0].CanonicalDir != "" {
		canonical = clone[0].CanonicalDir
	}
	return &FileService{DB: d, Root: root, URL: strings.TrimRight(url, "/"), CanonicalRoot: canonical}
}
func (f *FileService) SafePath(path string) (string, error) {
	return safeFilePath(f.Root, path)
}
func safeFilePath(root, path string) (string, error) {
	base, e := filepath.Abs(root)
	if e != nil {
		return "", e
	}
	p, e := filepath.Abs(path)
	if e != nil {
		return "", e
	}
	base, e = filepath.EvalSymlinks(base)
	if e != nil {
		return "", e
	}
	p, e = filepath.EvalSymlinks(p)
	if e != nil {
		return "", e
	}
	rel, e := filepath.Rel(base, p)
	if e != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("file outside upload directory")
	}
	return p, nil
}
func (f *FileService) CanonicalPath(path string, appID int64) (string, error) {
	p, e := safeFilePath(f.CanonicalRoot, path)
	if e != nil {
		return "", e
	}
	root, e := filepath.Abs(filepath.Join(f.CanonicalRoot, fmt.Sprint(appID)))
	if e != nil {
		return "", e
	}
	root, e = filepath.EvalSymlinks(root)
	if e != nil {
		return "", e
	}
	rel, e := filepath.Rel(root, p)
	if e != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("APK belongs to a different application")
	}
	return p, nil
}
func (f *FileService) Save(c *gin.Context, dir string, max int64, imageOnly bool) (path, url, name, mime string, size int64, err error) {
	// Bound multipart parsing itself, including ignored extra parts.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, max+(1<<20))
	if c.Request.ContentLength > max+(1<<20) {
		err = fmt.Errorf("文件超过大小限制")
		return
	}
	header, e := c.FormFile("file")
	if e != nil {
		err = e
		return
	}
	if header.Size > max {
		err = fmt.Errorf("文件超过大小限制")
		return
	}
	in, e := header.Open()
	if e != nil {
		err = e
		return
	}
	defer in.Close()
	head := make([]byte, 512)
	n, e := in.Read(head)
	if e != nil && e != io.EOF {
		err = e
		return
	}
	mime = http.DetectContentType(head[:n])
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if imageOnly {
		allowed := map[string]string{"image/png": ".png", "image/jpeg": ".jpg", "image/gif": ".gif", "image/webp": ".webp"}
		ext = allowed[mime]
		if ext == "" {
			err = fmt.Errorf("仅支持 PNG/JPG/GIF/WEBP 图片")
			return
		}
	}
	random := make([]byte, 16)
	if _, err = rand.Read(random); err != nil {
		return
	}
	name = filepath.Base(header.Filename)
	if len(name) > 255 || strings.ContainsAny(name, "\r\n\x00") {
		err = fmt.Errorf("文件名无效或过长")
		return
	}
	filename := hex.EncodeToString(random) + ext
	path = filepath.Join(f.Root, dir, filename)
	if strings.HasPrefix(dir, "canonical/") {
		path = filepath.Join(f.CanonicalRoot, strings.TrimPrefix(dir, "canonical/"), filename)
	}
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return
	}
	out, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		err = e
		return
	}
	size, e = io.Copy(out, io.LimitReader(io.MultiReader(strings.NewReader(string(head[:n])), in), max+1))
	closeErr := out.Close()
	if e != nil || size > max || closeErr != nil {
		os.Remove(path)
		err = fmt.Errorf("文件写入失败或超过限制")
		return
	}
	url = f.URL + "/" + dir + "/" + filename
	return
}

const fileSelect = "SELECT id,appid,name,filename,filepath,filesize,filetype,category,remark,version,download_url,download_count,created_at,updated_at FROM {extend_files}"

// Open once and reject directories; never sniff HTML/JSON as an API response.
func serveDownload(c *gin.Context, path, name string) {
	in, err := os.Open(path)
	if err != nil {
		c.Status(404)
		return
	}
	defer in.Close()
	info, err := in.Stat()
	if err != nil || !info.Mode().IsRegular() {
		c.Status(404)
		return
	}
	c.Header("Content-Type", "application/octet-stream")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": name}))
	c.Header("Cache-Control", "no-store")
	http.ServeContent(c.Writer, c.Request, name, info.ModTime(), in)
}

// The recovered executable publishes canonical APKs and extension files by URL.
// Serve individual files only: no directory listing or symlink escape.
func (f *FileService) MountDownloads(g *gin.Engine) {
	serve := func(c *gin.Context) {
		name := c.Param("file")
		if name == "" || name == "." || strings.ContainsAny(name, "/\\") {
			c.Status(404)
			return
		}
		var path string
		var err error
		var allowed int
		if c.Param("appid") != "" {
			appID, e := strconv.ParseInt(c.Param("appid"), 10, 64)
			if e != nil || appID <= 0 {
				c.Status(404)
				return
			}
			path, err = f.CanonicalPath(filepath.Join(f.CanonicalRoot, strconv.FormatInt(appID, 10), name), appID)
			if err == nil {
				err = f.DB.GetContext(c.Request.Context(), &allowed, repo.SQL("SELECT COUNT(*) FROM {versions} v JOIN {apps} a ON a.id=v.appid WHERE a.id=? AND a.status=1 AND v.status=1 AND v.is_canonical=1 AND v.canonical_path=?"), appID, filepath.Join(f.CanonicalRoot, strconv.FormatInt(appID, 10), name))
			}
		} else {
			path, err = safeFilePath(filepath.Join(f.Root, "extend_files"), filepath.Join(f.Root, "extend_files", name))
			if err == nil {
				err = f.DB.GetContext(c.Request.Context(), &allowed, repo.SQL("SELECT COUNT(*) FROM {extend_files} f LEFT JOIN {apps} a ON a.id=f.appid WHERE f.filepath=? AND (f.appid=0 OR a.status=1)"), filepath.Join(f.Root, "extend_files", name))
			}
		}
		if err != nil || allowed == 0 {
			c.Status(404)
			return
		}
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() {
			c.Status(404)
			return
		}
		serveDownload(c, path, name)
	}
	for _, dir := range []string{"avatars", "app_logos"} {
		dir := dir
		imageHandler := func(c *gin.Context) {
			name := c.Param("file")
			if name == "" || strings.ContainsAny(name, "/\\") {
				c.Status(404)
				return
			}
			path, err := safeFilePath(filepath.Join(f.Root, dir), filepath.Join(f.Root, dir, name))
			if err != nil {
				c.Status(404)
				return
			}
			in, err := os.Open(path)
			if err != nil {
				c.Status(404)
				return
			}
			defer in.Close()
			info, err := in.Stat()
			if err != nil || !info.Mode().IsRegular() {
				c.Status(404)
				return
			}
			head := make([]byte, 512)
			n, _ := in.Read(head)
			kind := http.DetectContentType(head[:n])
			if kind != "image/png" && kind != "image/jpeg" && kind != "image/gif" && kind != "image/webp" {
				c.Status(404)
				return
			}
			if _, err = in.Seek(0, 0); err != nil {
				c.Status(404)
				return
			}
			c.Header("Content-Type", kind)
			c.Header("X-Content-Type-Options", "nosniff")
			http.ServeContent(c.Writer, c.Request, name, info.ModTime(), in)
		}
		g.GET(f.URL+"/"+dir+"/:file", imageHandler)
		g.HEAD(f.URL+"/"+dir+"/:file", imageHandler)
	}
	for _, route := range []string{f.URL + "/canonical/:appid/:file", f.URL + "/extend_files/:file"} {
		g.GET(route, serve)
		g.HEAD(route, serve)
	}
}

func (f *FileService) Mount(admin *gin.RouterGroup) {
	for _, route := range []struct{ path, dir string }{{"/upload/avatar", "avatars"}, {"/upload/app-logo", "app_logos"}} {
		route := route
		admin.POST(route.path, func(c *gin.Context) {
			_, url, _, _, _, e := f.Save(c, route.dir, 5<<20, true)
			if e != nil {
				response.BadRequest(c, e.Error())
				return
			}
			response.OK(c, gin.H{"url": url})
		})
	}
	admin.GET("/extend-file/list", f.List)
	admin.GET("/extend-file/categories", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		v := []string{}
		e := f.DB.SelectContext(c.Request.Context(), &v, repo.SQL("SELECT DISTINCT category FROM {extend_files} WHERE (appid=? OR appid=0) AND category<>'' ORDER BY category"), a)
		if !dbError(c, e) {
			response.OK(c, v)
		}
	})
	admin.POST("/extend-file/upload", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		app, err := repo.NewAppRepo(f.DB).GetByID(c.Request.Context(), a)
		if dbError(c, err) {
			return
		}
		if app == nil {
			response.Fail(c, 404, "应用不存在")
			return
		}
		settings, e := repo.NewSettingsRepo(f.DB).Get(c.Request.Context())
		if dbError(c, e) {
			return
		}
		limit := settings.UploadLimit
		if limit < 1 {
			limit = 10
		}
		path, url, name, mime, size, e := f.Save(c, "extend_files", int64(limit)<<20, false)
		if e != nil {
			response.BadRequest(c, e.Error())
			return
		}
		display := c.PostForm("name")
		if display == "" {
			display = name
		}
		now := time.Now().Unix()
		result, e := f.DB.ExecContext(c.Request.Context(), repo.SQL("INSERT INTO {extend_files} (appid,name,filename,filepath,filesize,filetype,category,remark,version,download_url,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)"), a, display, name, path, size, mime, c.PostForm("category"), c.PostForm("remark"), c.PostForm("version"), url, now, now)
		if e != nil {
			os.Remove(path)
			dbError(c, e)
			return
		}
		id, e := result.LastInsertId()
		if !dbError(c, e) {
			response.OK(c, gin.H{"id": id, "name": display, "downloadUrl": url})
		}
	})
	admin.PUT("/extend-file/:id", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		var v struct{ Name, Category, Remark, Version string }
		if c.ShouldBindJSON(&v) != nil || v.Name == "" {
			response.BadRequest(c, "文件名不能为空")
			return
		}
		var exists int
		if dbError(c, f.DB.GetContext(c.Request.Context(), &exists, repo.SQL("SELECT COUNT(*) FROM {extend_files} WHERE appid=? AND id=?"), a, id)) {
			return
		}
		if exists == 0 {
			response.Fail(c, 404, "文件不存在")
			return
		}
		_, e := f.DB.ExecContext(c.Request.Context(), repo.SQL("UPDATE {extend_files} SET name=?,category=?,remark=?,version=?,updated_at=? WHERE appid=? AND id=?"), v.Name, v.Category, v.Remark, v.Version, time.Now().Unix(), a, id)
		if !dbError(c, e) {
			response.OK(c, nil)
		}
	})
	admin.DELETE("/extend-file/:id", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		var v repo.ExtendFile
		err := f.DB.GetContext(c.Request.Context(), &v, repo.SQL(fileSelect+" WHERE appid=? AND id=?"), a, id)
		if errors.Is(err, sql.ErrNoRows) {
			response.Fail(c, 404, "文件不存在")
			return
		}
		if dbError(c, err) {
			return
		}
		path, e := safeFilePath(filepath.Join(f.Root, "extend_files"), v.Filepath)
		if e != nil && !os.IsNotExist(e) && dbError(c, e) {
			return
		}
		if _, e = f.DB.ExecContext(c.Request.Context(), repo.SQL("DELETE FROM {extend_files} WHERE appid=? AND id=?"), a, id); dbError(c, e) {
			return
		}
		if path != "" {
			e = os.Remove(path)
		} else {
			e = nil
		}
		if e != nil && !os.IsNotExist(e) {
			dbError(c, e)
			return
		}
		response.OK(c, nil)
	})
	admin.GET("/extend-file/:id/download", f.Download)
}
func (f *FileService) List(c *gin.Context) {
	a, ok := requireAppID(c)
	if !ok {
		return
	}
	q := fileSelect + " WHERE (appid=? OR appid=0)"
	args := []any{a}
	if category := c.Query("category"); category != "" {
		q += " AND category=?"
		args = append(args, category)
	}
	if keyword := c.Query("keyword"); keyword != "" {
		q += " AND (name LIKE ? OR remark LIKE ? OR version LIKE ?)"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	q += " ORDER BY appid DESC,created_at DESC,id DESC"
	v := []repo.ExtendFile{}
	if !dbError(c, f.DB.SelectContext(c.Request.Context(), &v, repo.SQL(q), args...)) {
		if middleware.CurrentApp(c) != nil {
			for i := range v {
				v[i].Filepath = ""
			}
		}
		response.OK(c, v)
	}
}
func (f *FileService) Download(c *gin.Context) {
	a, ok := requireAppID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var v repo.ExtendFile
	err := f.DB.GetContext(c.Request.Context(), &v, repo.SQL(fileSelect+" WHERE id=? AND (appid=? OR appid=0)"), id, a)
	if errors.Is(err, sql.ErrNoRows) {
		response.Fail(c, 404, "文件不存在")
		return
	}
	if dbError(c, err) {
		return
	}
	path, e := safeFilePath(filepath.Join(f.Root, "extend_files"), v.Filepath)
	if e != nil {
		response.Fail(c, 404, "文件不存在")
		return
	}
	if _, e = os.Stat(path); e != nil {
		response.Fail(c, 404, "文件不存在")
		return
	}
	if _, e = f.DB.ExecContext(c.Request.Context(), repo.SQL("UPDATE {extend_files} SET download_count=download_count+1 WHERE id=?"), id); dbError(c, e) {
		return
	}
	serveDownload(c, path, v.Filename)
}
func (f *FileService) SDKAvatar(c *gin.Context) {
	u := currentUser(c)
	if session := middleware.CurrentSDKUser(c); session != nil && session.IsKami {
		response.BadRequest(c, "卡密版账号不支持头像")
		return
	}
	path, url, _, _, _, e := f.Save(c, "avatars", 5<<20, true)
	if e != nil {
		response.BadRequest(c, e.Error())
		return
	}
	if e = repo.NewUserRepo(f.DB).UpdatePartial(c.Request.Context(), u.AppID, u.ID, map[string]any{"avatar": url}); e != nil {
		os.Remove(path)
		dbError(c, e)
		return
	}
	u.Avatar = url
	response.OK(c, gin.H{"avatar": url, "user": sdkUserProfile(u), "serverTime": time.Now().Unix()})
}
