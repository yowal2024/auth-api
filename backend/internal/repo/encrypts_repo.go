// Reconstructed SQL is extracted from the original executable.
package repo

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
)

func NewEncryptRepo(db *sqlx.DB) *EncryptRepo { return &EncryptRepo{db: db} }

func (r *EncryptRepo) GetByID(ctx context.Context, appid, id int64) (*Encrypt, error) {
	var v Encrypt
	e := r.db.GetContext(ctx, &v, SQL("SELECT "+AppRecordIDColumn("encrypts", "{encrypts}")+"id, appid, name, algorithm, encoding, sym_key, time_skew, sign_mode,\n\tCOALESCE(client_private_key, '') AS client_private_key,\n\tCOALESCE(client_public_key, '')  AS client_public_key,\n\tCOALESCE(server_private_key, '') AS server_private_key,\n\tCOALESCE(server_public_key, '')  AS server_public_key,\n\tcreate_time FROM {encrypts} WHERE appid = ? AND id = ? LIMIT 1"), appid, id)
	if errors.Is(e, sql.ErrNoRows) {
		return nil, nil
	}
	return &v, e
}

func (r *EncryptRepo) ListByApp(ctx context.Context, appid int64) ([]Encrypt, error) {
	v := []Encrypt{}
	e := r.db.SelectContext(ctx, &v, SQL("SELECT "+AppRecordIDColumn("encrypts", "{encrypts}")+"id, appid, name, algorithm, encoding, sym_key, time_skew, sign_mode,\n\tCOALESCE(client_private_key, '') AS client_private_key,\n\tCOALESCE(client_public_key, '')  AS client_public_key,\n\tCOALESCE(server_private_key, '') AS server_private_key,\n\tCOALESCE(server_public_key, '')  AS server_public_key,\n\tcreate_time FROM {encrypts} WHERE appid = ? ORDER BY id DESC"), appid)
	return v, e
}

func (r *EncryptRepo) Create(ctx context.Context, v *Encrypt) (int64, error) {
	res, e := ExecAppInsert(ctx, r.db, v.AppID, "encrypts", SQL("INSERT INTO {encrypts}\n\t\t (appid, name, algorithm, encoding, sym_key, time_skew, sign_mode,\n\t\t  client_private_key, client_public_key, server_private_key, server_public_key, create_time)\n\t\t VALUES (?,?,?,?,?,?,?,?,?,?,?,?)"), v.AppID, v.Name, v.Algorithm, v.Encoding, v.SymKey, v.TimeSkew, v.SignMode, v.ClientPrivateKey, v.ClientPublicKey, v.ServerPrivateKey, v.ServerPublicKey, v.CreateTime)
	if e != nil {
		return 0, e
	}
	return res.LastInsertId()
}

func (r *EncryptRepo) Update(ctx context.Context, v *Encrypt) error {
	_, e := r.db.ExecContext(ctx, SQL("UPDATE {encrypts} SET name=?, algorithm=?, encoding=?, sym_key=?, time_skew=?, sign_mode=?,\n\t\t client_private_key=?, client_public_key=?, server_private_key=?, server_public_key=?\n\t\t WHERE appid=? AND id=?"), v.Name, v.Algorithm, v.Encoding, v.SymKey, v.TimeSkew, v.SignMode, v.ClientPrivateKey, v.ClientPublicKey, v.ServerPrivateKey, v.ServerPublicKey, v.AppID, v.ID)
	return e
}

var ErrEncryptionInUse = errors.New("加密方案正在使用，请先在应用设置中切换方案")
var ErrEncryptionUnavailable = errors.New("加密方案不存在或不属于当前应用")

func (r *EncryptRepo) Delete(ctx context.Context, appid, id int64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// Serialize policy deletion with app updates before checking the selection.
	var selected int64
	if err = tx.GetContext(ctx, &selected, SQL("SELECT encryption FROM {apps} WHERE id=? FOR UPDATE"), appid); err != nil {
		return err
	}
	if selected == id {
		return ErrEncryptionInUse
	}
	res, err := tx.ExecContext(ctx, SQL("DELETE FROM {encrypts} WHERE appid = ? AND id = ?"), appid, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrEncryptionUnavailable
	}
	return tx.Commit()
}
