// Reconstructed SQL is extracted from the original executable.
package repo

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
)

func NewAppRepo(db *sqlx.DB) *AppRepo { return &AppRepo{db: db} }

func (r *AppRepo) HasEncryption(ctx context.Context, appID, id int64) (bool, error) {
	var n int
	err := r.db.GetContext(ctx, &n, SQL("SELECT COUNT(*) FROM {encrypts} WHERE appid=? AND id=?"), appID, id)
	return n == 1, err
}

func (r *AppRepo) GetByID(ctx context.Context, id int64) (*App, error) {
	var v App
	e := r.db.GetContext(ctx, &v, SQL("SELECT id, name, type, mode, COALESCE(logo, '') AS logo, pkg, token_key,\n  bind_device, bind_region, diff_update, random_package,\n  encryption, user_token_ttl, geo_provider, geo_key, ip_distance_max,\n  brute_force_enabled, mtls_enabled, status, create_time FROM {apps} WHERE id = ? LIMIT 1"), id)
	if errors.Is(e, sql.ErrNoRows) {
		return nil, nil
	}
	return &v, e
}

func (r *AppRepo) ListByApp(ctx context.Context) ([]App, error) {
	v := []App{}
	e := r.db.SelectContext(ctx, &v, SQL("SELECT id, name, type, mode, COALESCE(logo, '') AS logo, pkg, token_key,\n  bind_device, bind_region, diff_update, random_package,\n  encryption, user_token_ttl, geo_provider, geo_key, ip_distance_max,\n  brute_force_enabled, mtls_enabled, status, create_time FROM {apps} ORDER BY id ASC"))
	return v, e
}

func (r *AppRepo) Create(ctx context.Context, v *App) (int64, error) {
	return createApp(ctx, r.db, v)
}
func createApp(ctx context.Context, d sqlx.ExecerContext, v *App) (int64, error) {
	r := struct{ db sqlx.ExecerContext }{d}
	res, e := r.db.ExecContext(ctx, SQL("INSERT INTO {apps} (name, type, mode, logo, pkg, token_key,\n\t\t  bind_device, bind_region, diff_update, random_package,\n\t\t  encryption, user_token_ttl, geo_provider, geo_key, ip_distance_max,\n\t\t  brute_force_enabled, mtls_enabled, status, create_time)\n\t\t VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"), v.Name, v.Type, v.Mode, v.Logo, v.Pkg, v.TokenKey, v.BindDevice, v.BindRegion, v.DiffUpdate, v.RandomPackage, v.Encryption, v.UserTokenTTL, v.GeoProvider, v.GeoKey, v.IPDistanceMax, v.BruteForce, v.MtlsEnabled, v.Status, v.CreateTime)
	if e != nil {
		return 0, e
	}
	return res.LastInsertId()
}

func (r *AppRepo) Update(ctx context.Context, v *App) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var appID int64
	if err = tx.GetContext(ctx, &appID, SQL("SELECT id FROM {apps} WHERE id=? FOR UPDATE"), v.ID); err != nil {
		return err
	}
	if v.Encryption > 0 {
		var policyID int64
		err = tx.GetContext(ctx, &policyID, SQL("SELECT id FROM {encrypts} WHERE appid=? AND id=? FOR UPDATE"), v.ID, v.Encryption)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrEncryptionUnavailable
		}
		if err != nil {
			return err
		}
	}
	_, e := tx.ExecContext(ctx, SQL("UPDATE {apps} SET name=?, type=?, mode=?, logo=?, pkg=?, token_key=?,\n\t\t  bind_device=?, bind_region=?, diff_update=?, random_package=?,\n\t\t  encryption=?, user_token_ttl=?, geo_provider=?, geo_key=?, ip_distance_max=?,\n\t\t  brute_force_enabled=?, mtls_enabled=?, status=?\n\t\t WHERE id = ?"), v.Name, v.Type, v.Mode, v.Logo, v.Pkg, v.TokenKey, v.BindDevice, v.BindRegion, v.DiffUpdate, v.RandomPackage, v.Encryption, v.UserTokenTTL, v.GeoProvider, v.GeoKey, v.IPDistanceMax, v.BruteForce, v.MtlsEnabled, v.Status, v.ID)
	if e != nil {
		return e
	}
	return tx.Commit()
}

// CreateInherited copies the selected encryption policy into the new app scope.
func (r *AppRepo) CreateInherited(ctx context.Context, v *App, source int64) (int64, error) {
	if source <= 0 || v.Encryption <= 0 {
		return r.Create(ctx, v)
	}
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	id, err := createApp(ctx, tx, v)
	if err != nil {
		return 0, err
	}
	res, err := tx.ExecContext(ctx, SQL("INSERT INTO {encrypts} (appid,name,algorithm,encoding,sym_key,time_skew,sign_mode,client_private_key,client_public_key,server_private_key,server_public_key,create_time) SELECT ?,name,algorithm,encoding,sym_key,time_skew,sign_mode,client_private_key,client_public_key,server_private_key,server_public_key,? FROM {encrypts} WHERE appid=? AND id=?"), id, v.CreateTime, source, v.Encryption)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if n != 1 {
		return 0, errors.New("inherited encryption policy missing")
	}
	encryption, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err = AssignAppRecordIDs(ctx, tx, id, "encrypts", []int64{encryption}); err != nil {
		return 0, err
	}
	if _, err = tx.ExecContext(ctx, SQL("UPDATE {apps} SET encryption=? WHERE id=?"), encryption, id); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	v.Encryption = encryption
	return id, nil
}
