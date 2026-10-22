package handler

import (
	"auth-api/internal/buildpool"
	"auth-api/internal/buildpool/edge"
	"auth-api/internal/buildpool/v2sign"
	"auth-api/internal/config"
	"auth-api/internal/middleware"
	"auth-api/internal/repo"
	"auth-api/internal/response"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"os"
	"path/filepath"
	"time"
)

const versionSelect = "SELECT id,appid,version_name,version_key,version_no,status,url,apk_name,apk_size,apk_url,is_canonical,canonical_path,clone_diff_enable,original_package,COALESCE(changelog,'') AS changelog,create_time FROM {versions}"

type VersionService struct {
	Config config.CloneConfig
	DB     *sqlx.DB
	Files  *FileService
	Diffs  *buildpool.DiffStore
}

func (v *VersionService) Mount(g *gin.RouterGroup) {
	g.GET("/ver/list", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		items := []repo.Version{}
		if !dbError(c, v.DB.SelectContext(c.Request.Context(), &items, repo.SQL(versionSelect+" WHERE appid=? ORDER BY id DESC"), a)) {
			response.OK(c, items)
		}
	})
	g.POST("/ver", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		var r versionSaveReq
		if c.ShouldBindJSON(&r) != nil || r.VersionName == "" || r.VersionKey == "" {
			response.BadRequest(c, "版本名称和索引不能为空")
			return
		}
		if r.IsCanonical == 1 {
			path, e := v.Files.CanonicalPath(r.CanonicalPath, a)
			if e != nil {
				response.BadRequest(c, "无效的 APK 路径")
				return
			}
			raw, e := os.ReadFile(path)
			if e != nil {
				response.BadRequest(c, "APK 不存在")
				return
			}
			pkg, e := buildpool.ExtractPackageName(raw)
			if e != nil {
				response.BadRequest(c, "APK 包名解析失败")
				return
			}
			r.OriginalPackage = pkg
		}
		status := 1
		if r.Status != nil {
			status = *r.Status
		}
		args := []any{r.VersionName, r.VersionKey, r.VersionNo, status, r.URL, r.ApkName, r.ApkSize, r.ApkURL, r.IsCanonical, r.CanonicalPath, r.CloneDiffEnable, r.OriginalPackage, r.Changelog}
		if r.ID > 0 {
			args = append(args, a, r.ID)
			_, e := v.DB.ExecContext(c.Request.Context(), repo.SQL("UPDATE {versions} SET version_name=?,version_key=?,version_no=?,status=?,url=?,apk_name=?,apk_size=?,apk_url=?,is_canonical=?,canonical_path=?,clone_diff_enable=?,original_package=?,changelog=? WHERE appid=? AND id=?"), args...)
			if !dbError(c, e) {
				response.OK(c, gin.H{"id": r.ID})
			}
		} else {
			args = append([]any{a}, args...)
			args = append(args, time.Now().Unix())
			res, e := v.DB.ExecContext(c.Request.Context(), repo.SQL("INSERT INTO {versions} (appid,version_name,version_key,version_no,status,url,apk_name,apk_size,apk_url,is_canonical,canonical_path,clone_diff_enable,original_package,changelog,create_time) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"), args...)
			if dbError(c, e) {
				return
			}
			id, e := res.LastInsertId()
			if !dbError(c, e) {
				response.OK(c, gin.H{"id": id})
			}
		}
	})
	g.DELETE("/ver/:id", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		_, e := v.DB.ExecContext(c.Request.Context(), repo.SQL("DELETE FROM {versions} WHERE appid=? AND id=?"), a, id)
		if !dbError(c, e) {
			response.OK(c, nil)
		}
	})
	g.POST("/ver/upload-apk", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		path, url, name, _, size, e := v.Files.Save(c, fmt.Sprintf("canonical/%d", a), 200<<20, false)
		if e != nil {
			response.BadRequest(c, e.Error())
			return
		}
		if filepath.Ext(name) != ".apk" {
			os.Remove(path)
			response.BadRequest(c, "仅支持 .apk 文件")
			return
		}
		raw, e := os.ReadFile(path)
		if dbError(c, e) {
			return
		}
		pkg, e := buildpool.ExtractPackageName(raw)
		if e != nil {
			os.Remove(path)
			response.BadRequest(c, "APK 包名解析失败: "+e.Error())
			return
		}
		hash := sha256.Sum256(raw)
		pairs, _ := edge.Pairs(raw)
		response.OK(c, gin.H{"path": path, "sha256": hex.EncodeToString(hash[:]), "originalPackage": pkg, "size": size, "url": url, "name": name, "hasEdgePair": len(pairs) > 0})
	})
}
func (v *VersionService) signKey(c *gin.Context, regenerate bool) (*v2sign.KeyAndCert, error) {
	u := currentUser(c)
	if u == nil {
		return nil, fmt.Errorf("user missing")
	}
	tx, e := v.DB.BeginTxx(c.Request.Context(), nil)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback()
	var row struct {
		Private []byte `db:"sign_pk8"`
		Cert    []byte `db:"sign_x509"`
	}
	e = tx.GetContext(c.Request.Context(), &row, repo.SQL("SELECT sign_pk8,sign_x509 FROM {users} WHERE appid=? AND id=? FOR UPDATE"), u.AppID, u.ID)
	if e != nil {
		return nil, e
	}
	var key *v2sign.KeyAndCert
	if len(row.Private) > 0 && len(row.Cert) > 0 && !regenerate {
		key, e = v2sign.LoadKeyAndCert(row.Private, row.Cert)
		if e != nil {
			return nil, e
		}
	} else {
		key, e = v2sign.GenerateKeyAndCert(u.Username)
		if e != nil {
			return nil, e
		}
		priv, cert, e := v2sign.MarshalKeyAndCert(key)
		if e != nil {
			return nil, e
		}
		if _, e = tx.ExecContext(c.Request.Context(), repo.SQL("UPDATE {users} SET sign_pk8=?,sign_x509=? WHERE appid=? AND id=?"), priv, cert, u.AppID, u.ID); e != nil {
			return nil, e
		}
	}
	return key, tx.Commit()
}
func (v *VersionService) Clone(c *gin.Context) {
	if middleware.CurrentSDKUser(c).IsKami {
		response.BadRequest(c, "卡密版不支持独立 APK 签名")
		return
	}
	app := middleware.CurrentApp(c)
	if !app.RandomPackage {
		response.Fail(c, 1005, "应用未开启随机包名")
		return
	}
	var version repo.Version
	q := versionSelect + " WHERE appid=? AND status=1 AND is_canonical=1"
	args := []any{app.ID}
	if key := c.Query("versionKey"); key != "" {
		q += " AND version_key=?"
		args = append(args, key)
	}
	q += " ORDER BY create_time DESC,id DESC LIMIT 1"
	e := v.DB.GetContext(c.Request.Context(), &version, repo.SQL(q), args...)
	if errors.Is(e, sql.ErrNoRows) {
		response.Fail(c, 1004, "没有可用的 APK 版本")
		return
	}
	if dbError(c, e) {
		return
	}
	path, e := v.Files.CanonicalPath(version.CanonicalPath, app.ID)
	if dbError(c, e) {
		return
	}
	key, e := v.signKey(c, false)
	if dbError(c, e) {
		return
	}
	store, e := v.appDiffs(app.ID)
	if dbError(c, e) {
		return
	}
	meta, e := store.Generate(path, key)
	if errors.Is(e, buildpool.ErrBusy) {
		response.Fail(c, 1017, "补丁生成繁忙，请稍后重试")
		return
	}
	if !dbError(c, e) {
		response.OK(c, gin.H{"diffUrl": "/api/sdk/diff/" + meta.DiffName, "newPackage": meta.NewPackage, "originalSha256": meta.OriginalSha256, "cloneSha256": meta.CloneSha256, "versionName": version.VersionName, "versionKey": version.VersionKey, "versionNo": version.VersionNo})
	}
}
func (v *VersionService) RegenerateSign(c *gin.Context) {
	if middleware.CurrentSDKUser(c).IsKami {
		response.BadRequest(c, "卡密版不支持独立 APK 签名")
		return
	}
	key, e := v.signKey(c, true)
	if !dbError(c, e) {
		response.OK(c, gin.H{"regenerated": true, "certBytes": len(key.CertDER)})
	}
}
func (v *VersionService) Diff(c *gin.Context) {
	name := c.Param("name")
	store, e := v.appDiffs(middleware.CurrentApp(c).ID)
	if dbError(c, e) {
		return
	}
	path, e := store.DiffPath(name)
	if e != nil {
		c.Status(404)
		return
	}
	serveDownload(c, path, name)
}
func (v *VersionService) appDiffs(id int64) (*buildpool.DiffStore, error) {
	cfg := v.Config
	if cfg.OutputDir == "" {
		cfg.OutputDir = "uploads/diffs"
	}
	cfg.OutputDir = filepath.Join(cfg.OutputDir, fmt.Sprint(id))
	return buildpool.NewDiffStore(cfg)
}
