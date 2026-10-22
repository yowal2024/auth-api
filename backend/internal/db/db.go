package db

import (
	"auth-api/internal/config"
	"auth-api/internal/repo"
	"context"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"time"
)

func Open(ctx context.Context, c config.MySQLConfig) (*sqlx.DB, error) {
	if e := repo.InitTablePrefix(c.TablePrefix); e != nil {
		return nil, e
	}
	d, e := sqlx.ConnectContext(ctx, "mysql", c.DSN)
	if e != nil {
		return nil, e
	}
	d.SetMaxOpenConns(c.MaxOpen)
	d.SetMaxIdleConns(c.MaxIdle)
	d.SetConnMaxLifetime(time.Duration(c.ConnLifetime) * time.Second)
	return d, nil
}
