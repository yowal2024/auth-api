// Reconstructed SQL is extracted from the original executable.
package repo

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
)

func NewUserGroupRepo(db *sqlx.DB) *UserGroupRepo { return &UserGroupRepo{db: db} }

func (r *UserGroupRepo) GetByID(ctx context.Context, appid, id int64) (*UserGroup, error) {
	var v UserGroup
	e := r.db.GetContext(ctx, &v, SQL("SELECT id, appid, name, level, remark, create_time FROM {user_groups} WHERE appid = ? AND id = ? LIMIT 1"), appid, id)
	if errors.Is(e, sql.ErrNoRows) {
		return nil, nil
	}
	return &v, e
}

func (r *UserGroupRepo) ListByApp(ctx context.Context, appid int64) ([]UserGroup, error) {
	v := []UserGroup{}
	e := r.db.SelectContext(ctx, &v, SQL("SELECT id, appid, name, level, remark, create_time FROM {user_groups} WHERE appid = ? ORDER BY level DESC, id DESC"), appid)
	return v, e
}

func (r *UserGroupRepo) Create(ctx context.Context, v *UserGroup) (int64, error) {
	res, e := r.db.ExecContext(ctx, SQL("INSERT INTO {user_groups} (appid, name, level, remark, create_time) VALUES (?,?,?,?,?)"), v.AppID, v.Name, v.Level, v.Remark, v.CreateTime)
	if e != nil {
		return 0, e
	}
	return res.LastInsertId()
}

func (r *UserGroupRepo) Update(ctx context.Context, v *UserGroup) error {
	_, e := r.db.ExecContext(ctx, SQL("UPDATE {user_groups} SET name=?, level=?, remark=? WHERE appid=? AND id=?"), v.Name, v.Level, v.Remark, v.AppID, v.ID)
	return e
}

func (r *UserGroupRepo) Delete(ctx context.Context, appid, id int64) error {
	_, e := r.db.ExecContext(ctx, SQL("DELETE FROM {user_groups} WHERE appid = ? AND id = ?"), appid, id)
	return e
}
