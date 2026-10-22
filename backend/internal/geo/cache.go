package geo

import (
	"auth-api/internal/repo"
	"context"
	"database/sql"
	"errors"
	"time"
)

// Resolver caches by provider and IP using the recovered ip_info schema.
type CacheDB interface {
	GetContext(context.Context, any, string, ...any) error
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func Resolve(ctx context.Context, d CacheDB, provider, key, ip string) (Point, error) {
	var cache int
	if err := d.GetContext(ctx, &cache, repo.SQL("SELECT cache_enabled FROM {settings} WHERE id=1")); err != nil {
		return Point{}, err
	}
	if cache == 0 {
		return Lookup(ctx, provider, key, ip)
	}
	if provider == "" {
		provider = "tencent"
	}
	var row struct {
		Lat      float64 `db:"lat"`
		Lng      float64 `db:"lng"`
		Nation   string  `db:"nation"`
		Province string  `db:"province"`
		City     string  `db:"city"`
		District string  `db:"district"`
		Updated  int64   `db:"update_time"`
	}
	e := d.GetContext(ctx, &row, repo.SQL("SELECT lat,lng,nation,province,city,district,update_time FROM {ip_info} WHERE ip=? AND provider=?"), ip, provider)
	if e == nil && time.Now().Unix()-row.Updated < 86400 {
		return Point{row.Lat, row.Lng, row.Nation, row.Province, row.City, row.District}, nil
	}
	if e != nil && !errors.Is(e, sql.ErrNoRows) {
		return Point{}, e
	}
	p, e := Lookup(ctx, provider, key, ip)
	if e != nil {
		return p, e
	}
	_, e = d.ExecContext(ctx, repo.SQL("INSERT INTO {ip_info} (ip,provider,lat,lng,nation,province,city,district,update_time) VALUES (?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE lat=VALUES(lat),lng=VALUES(lng),nation=VALUES(nation),province=VALUES(province),city=VALUES(city),district=VALUES(district),update_time=VALUES(update_time)"), ip, provider, p.Lat, p.Lng, p.Nation, p.Province, p.City, p.District, time.Now().Unix())
	return p, e
}
