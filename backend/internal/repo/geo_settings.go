package repo

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type GeoSettings struct {
	Provider string `db:"geo_provider" json:"provider"`
	Key      string `db:"geo_key" json:"key"`
}

// An empty provider only occurs when an upgrade finds conflicting application
// settings. Keep those applications working until a global service is saved.
func (s GeoSettings) ForApp(app *App) GeoSettings {
	if s.Provider != "" {
		return s
	}
	provider := app.GeoProvider
	if provider == "" {
		provider = "tencent"
	}
	return GeoSettings{Provider: provider, Key: app.GeoKey}
}

func LoadGeoSettings(ctx context.Context, q interface {
	GetContext(context.Context, any, string, ...any) error
}) (GeoSettings, error) {
	var settings GeoSettings
	err := q.GetContext(ctx, &settings, SQL("SELECT geo_provider,geo_key FROM {settings} WHERE id=1"))
	return settings, err
}

func SaveGeoSettings(ctx context.Context, db *sqlx.DB, settings GeoSettings) error {
	_, err := db.ExecContext(ctx, SQL("UPDATE {settings} SET geo_provider=?,geo_key=? WHERE id=1"), settings.Provider, settings.Key)
	return err
}

func (r *AppRepo) GeoSettings(ctx context.Context, app *App) (GeoSettings, error) {
	settings, err := LoadGeoSettings(ctx, r.db)
	return settings.ForApp(app), err
}

type LegacyGeoSetting struct {
	ID       int64  `db:"id" json:"id"`
	Name     string `db:"name" json:"name"`
	Provider string `db:"geo_provider" json:"provider"`
	Key      string `db:"geo_key" json:"key"`
}

func LegacyGeoSettings(ctx context.Context, db *sqlx.DB) ([]LegacyGeoSetting, error) {
	rows := []LegacyGeoSetting{}
	err := db.SelectContext(ctx, &rows, SQL("SELECT id,name,CASE WHEN geo_provider='' THEN 'tencent' ELSE geo_provider END AS geo_provider,geo_key FROM {apps} WHERE bind_region=1 OR (geo_provider IN ('tencent','amap','') AND TRIM(geo_key)<>'') ORDER BY id"))
	return rows, err
}
