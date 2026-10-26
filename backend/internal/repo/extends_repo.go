// Reconstructed SQL is extracted from the original executable.
package repo

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
)

func NewExtendRepo(db *sqlx.DB) *ExtendRepo { return &ExtendRepo{db: db} }

func (r *ExtendRepo) GetByID(ctx context.Context, appid, id int64) (*Extend, error) {
	var v Extend
	e := r.db.GetContext(ctx, &v, SQL("SELECT "+AppRecordIDColumn("extends", "{extends}")+"id, appid, e_key, e_value, remark, update_time FROM {extends} WHERE appid = ? AND id = ? LIMIT 1"), appid, id)
	if errors.Is(e, sql.ErrNoRows) {
		return nil, nil
	}
	return &v, e
}

func (r *ExtendRepo) ListByApp(ctx context.Context, appid int64) ([]Extend, error) {
	v := []Extend{}
	e := r.db.SelectContext(ctx, &v, SQL("SELECT "+AppRecordIDColumn("extends", "{extends}")+"id, appid, e_key, e_value, remark, update_time FROM {extends} WHERE appid = ? ORDER BY update_time DESC,id DESC"), appid)
	return v, e
}

func (r *ExtendRepo) Create(ctx context.Context, v *Extend) (int64, error) {
	res, e := ExecAppInsert(ctx, r.db, v.AppID, "extends", SQL("INSERT INTO {extends} (appid, e_key, e_value, remark, update_time)\n\t\t VALUES (?,?,?,?,?)"), v.AppID, v.Key, v.Value, v.Remark, v.UpdateTime)
	if e != nil {
		return 0, e
	}
	return res.LastInsertId()
}

func (r *ExtendRepo) Update(ctx context.Context, v *Extend) error {
	_, e := r.db.ExecContext(ctx, SQL("UPDATE {extends} SET e_key=?, e_value=?, remark=?, update_time=?\n\t\t WHERE appid=? AND id=?"), v.Key, v.Value, v.Remark, v.UpdateTime, v.AppID, v.ID)
	return e
}

func (r *ExtendRepo) Delete(ctx context.Context, appid, id int64) error {
	_, e := r.db.ExecContext(ctx, SQL("DELETE FROM {extends} WHERE appid = ? AND id = ?"), appid, id)
	return e
}
