package handler

import (
	"auth-api/internal/buildpool"
	"auth-api/internal/middleware"
	"auth-api/internal/repo"
	"auth-api/internal/response"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"path/filepath"
)

func (s *SDKService) latestVersion(c *gin.Context) (any, bool) {
	var req sdkIniReq
	if c.Request.ContentLength != 0 {
		if c.ShouldBindJSON(&req) != nil {
			response.BadRequest(c, "请求参数错误")
			return nil, false
		}
	}
	app := middleware.CurrentApp(c)
	var latest repo.Version
	q := versionSelect + " WHERE appid=? AND status=1"
	args := []any{app.ID}
	e := s.db.GetContext(c.Request.Context(), &latest, repo.SQL(q+" ORDER BY create_time DESC,id DESC LIMIT 1"), args...)
	if errors.Is(e, sql.ErrNoRows) {
		return nil, true
	}
	if dbError(c, e) {
		return nil, false
	}
	hasUpdate := latest.VersionNo != req.VersionNo || latest.VersionKey != req.VersionKey
	v := gin.H{"versionName": latest.VersionName, "versionKey": latest.VersionKey, "versionNo": latest.VersionNo, "apkUrl": latest.ApkURL, "url": latest.URL, "changelog": latest.Changelog, "hasUpdate": hasUpdate}
	if app.DiffUpdate && latest.CloneDiffEnable == 1 && latest.IsCanonical == 1 && req.VersionNo != "" && hasUpdate {
		var old repo.Version
		oldKey := req.VersionKey
		if oldKey == "" {
			oldKey = latest.VersionKey
		}
		e = s.db.GetContext(c.Request.Context(), &old, repo.SQL(versionSelect+" WHERE appid=? AND version_key=? AND version_no=? AND is_canonical=1 ORDER BY id DESC LIMIT 1"), app.ID, oldKey, req.VersionNo)
		if e == nil {
			cfg := s.cloneConfig
			if cfg.OutputDir == "" {
				cfg.OutputDir = "uploads/diffs"
			}
			cfg.OutputDir = filepath.Join(cfg.OutputDir, fmt.Sprint(app.ID))
			files := NewFileService(s.db, s.uploadConfig, s.cloneConfig)
			oldPath, err := files.CanonicalPath(old.CanonicalPath, app.ID)
			if err != nil {
				_ = c.Error(err)
				return v, true
			}
			newPath, err := files.CanonicalPath(latest.CanonicalPath, app.ID)
			if err != nil {
				_ = c.Error(err)
				return v, true
			}
			store, err := buildpool.NewDiffStore(cfg)
			if err != nil {
				_ = c.Error(err)
				return v, true
			}
			meta, err := store.GenerateUpgradeDiff(oldPath, newPath)
			if err != nil {
				_ = c.Error(err)
				return v, true
			}
			v["diffUrl"] = "/api/sdk/diff/" + meta.DiffName
			v["diffOldSha"] = meta.OriginalSha256
			v["diffNewSha"] = meta.CloneSha256
		} else if !errors.Is(e, sql.ErrNoRows) {
			dbError(c, e)
			return nil, false
		}
	}
	return v, true
}
