// Reconstructed SQL is extracted from the original executable.
package repo

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
)

func NewFunctionRepo(db *sqlx.DB) *FunctionRepo { return &FunctionRepo{db: db} }

func (r *FunctionRepo) GetByID(ctx context.Context, appid, id int64) (*Function, error) {
	var v Function
	e := r.db.GetContext(ctx, &v, SQL("SELECT "+AppRecordIDColumn("functions", "{functions}")+"id, appid, name, `trigger`, code, remark, timeout, enabled, last_run FROM {functions} WHERE appid = ? AND id = ? LIMIT 1"), appid, id)
	if errors.Is(e, sql.ErrNoRows) {
		return nil, nil
	}
	return &v, e
}

func (r *FunctionRepo) ListByApp(ctx context.Context, appid int64) ([]Function, error) {
	v := []Function{}
	e := r.db.SelectContext(ctx, &v, SQL("SELECT "+AppRecordIDColumn("functions", "{functions}")+"id, appid, name, `trigger`, code, remark, timeout, enabled, last_run FROM {functions} WHERE appid = ? ORDER BY id DESC"), appid)
	return v, e
}

func (r *FunctionRepo) Create(ctx context.Context, v *Function) (int64, error) {
	res, e := ExecAppInsert(ctx, r.db, v.AppID, "functions", SQL("INSERT INTO {functions} (appid, name, `trigger`, code, remark, timeout, enabled, last_run) VALUES (?,?,?,?,?,?,?,?)"), v.AppID, v.Name, v.Trigger, v.Code, v.Remark, v.Timeout, v.Enabled, v.LastRun)
	if e != nil {
		return 0, e
	}
	return res.LastInsertId()
}

func (r *FunctionRepo) Update(ctx context.Context, v *Function) error {
	_, e := r.db.ExecContext(ctx, SQL("UPDATE {functions} SET name=?, `trigger`=?, code=?, remark=?, timeout=?, enabled=? WHERE appid=? AND id=?"), v.Name, v.Trigger, v.Code, v.Remark, v.Timeout, v.Enabled, v.AppID, v.ID)
	return e
}

func (r *FunctionRepo) Delete(ctx context.Context, appid, id int64) error {
	_, e := r.db.ExecContext(ctx, SQL("DELETE FROM {functions} WHERE appid = ? AND id = ?"), appid, id)
	return e
}
