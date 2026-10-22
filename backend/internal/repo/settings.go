package repo

import (
	"context"
	"github.com/jmoiron/sqlx"
)

func NewSettingsRepo(d *sqlx.DB) *SettingsRepo { return &SettingsRepo{db: d} }
func (r *SettingsRepo) Get(ctx context.Context) (*Settings, error) {
	var v Settings
	e := r.db.GetContext(ctx, &v, SQL("SELECT id,web_url,upload_limit,admin_token_expire,admin_entry,mysql_max_conn,mysql_max_idle,cache_enabled,user_log_enabled,admin_log_enabled,log_level,https_enabled,https_addr,COALESCE(https_cert,'') AS https_cert FROM {settings} WHERE id=1 LIMIT 1"))
	if v.LogLevel == "" || v.LogLevel == "info" {
		v.LogLevel = "all"
	}
	return &v, e
}
func (r *SettingsRepo) Refresh(ctx context.Context) error {
	v, e := r.Get(ctx)
	if e == nil {
		r.cache.Store(v)
	}
	return e
}
func (r *SettingsRepo) Snapshot() *Settings {
	v := r.cache.Load()
	if v == nil {
		return nil
	}
	copy := *v
	return &copy
}
func (r *SettingsRepo) Save(ctx context.Context, v *Settings) error {
	_, e := r.db.ExecContext(ctx, SQL("UPDATE {settings} SET web_url=?,upload_limit=?,admin_token_expire=?,admin_entry=?,mysql_max_conn=?,mysql_max_idle=?,cache_enabled=?,user_log_enabled=?,admin_log_enabled=?,log_level=? WHERE id=1"), v.WebURL, v.UploadLimit, v.AdminTokenExpire, v.AdminEntry, v.MysqlMaxConn, v.MysqlMaxIdle, v.CacheEnabled, v.UserLogEnabled, v.AdminLogEnabled, v.LogLevel)
	if e != nil {
		return e
	}
	return r.Refresh(ctx)
}
