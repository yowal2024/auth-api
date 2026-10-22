package repo

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
)

const adminSelect = "SELECT id, username, nickname, avatar, password, role, last_login, last_ip FROM {admins}"

func NewAdminRepo(d *sqlx.DB) *AdminRepo { return &AdminRepo{d} }
func (r *AdminRepo) FindByUsername(ctx context.Context, name string) (*Admin, error) {
	return r.find(ctx, "username", name)
}
func (r *AdminRepo) GetByID(ctx context.Context, id int64) (*Admin, error) {
	return r.find(ctx, "id", id)
}
func (r *AdminRepo) find(ctx context.Context, column string, value any) (*Admin, error) {
	var v Admin
	e := r.db.GetContext(ctx, &v, SQL(adminSelect+" WHERE "+column+" = ? LIMIT 1"), value)
	if errors.Is(e, sql.ErrNoRows) {
		return nil, nil
	}
	return &v, e
}
func (r *AdminRepo) List(ctx context.Context) ([]Admin, error) {
	v := []Admin{}
	e := r.db.SelectContext(ctx, &v, SQL(adminSelect+" ORDER BY id DESC"))
	return v, e
}
func (r *AdminRepo) Create(ctx context.Context, username, password, role string) (int64, error) {
	v, e := r.db.ExecContext(ctx, SQL("INSERT INTO {admins} (username, nickname, avatar, password, role, last_login, last_ip) VALUES (?, '', '', ?, ?, 0, '')"), username, password, role)
	if e != nil {
		return 0, e
	}
	return v.LastInsertId()
}
func (r *AdminRepo) UpdateLastLogin(ctx context.Context, id, ts int64, ip string) error {
	_, e := r.db.ExecContext(ctx, SQL("UPDATE {admins} SET last_login = ?, last_ip = ? WHERE id = ?"), ts, ip, id)
	return e
}
func (r *AdminRepo) UpdatePassword(ctx context.Context, id int64, hash string) error {
	_, e := r.db.ExecContext(ctx, SQL("UPDATE {admins} SET password = ? WHERE id = ?"), hash, id)
	return e
}
func (r *AdminRepo) UpdateRole(ctx context.Context, id int64, role string) error {
	_, e := r.db.ExecContext(ctx, SQL("UPDATE {admins} SET role = ? WHERE id = ?"), role, id)
	return e
}
func (r *AdminRepo) Delete(ctx context.Context, id int64) error {
	_, e := r.db.ExecContext(ctx, SQL("DELETE FROM {admins} WHERE id = ?"), id)
	return e
}
func (r *AdminRepo) UpdateProfile(ctx context.Context, id int64, name, nick, avatar string) error {
	var other int64
	for _, x := range []struct{ field, value string }{{"username", name}, {"nickname", nick}} {
		if x.value == "" {
			continue
		}
		e := r.db.GetContext(ctx, &other, SQL("SELECT id FROM {admins} WHERE "+x.field+" = ? AND id <> ? LIMIT 1"), x.value, id)
		if e == nil {
			return errors.New("账号或昵称已存在")
		}
		if !errors.Is(e, sql.ErrNoRows) {
			return e
		}
	}
	_, e := r.db.ExecContext(ctx, SQL("UPDATE {admins} SET username = ?, nickname = ?, avatar = ? WHERE id = ?"), name, nick, avatar, id)
	return e
}
