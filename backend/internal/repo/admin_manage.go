package repo

import (
	"context"
	"errors"
)

// Lock all administrators in ID order to preserve the last-super invariant.
func (r *AdminRepo) Manage(ctx context.Context, id int64, role, password string, remove bool) error {
	tx, e := r.db.BeginTxx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	rows := []Admin{}
	if e = tx.SelectContext(ctx, &rows, SQL(adminSelect+" ORDER BY id FOR UPDATE")); e != nil {
		return e
	}
	var target *Admin
	supers := 0
	for i := range rows {
		if rows[i].Role == "super" {
			supers++
		}
		if rows[i].ID == id {
			target = &rows[i]
		}
	}
	if target == nil {
		return errors.New("管理员不存在")
	}
	if target.Role == "super" && (remove || role != "super") && supers < 2 {
		return errors.New("必须保留至少一位超级管理员")
	}
	if remove {
		_, e = tx.ExecContext(ctx, SQL("DELETE FROM {admins} WHERE id=?"), id)
	} else {
		_, e = tx.ExecContext(ctx, SQL("UPDATE {admins} SET role=? WHERE id=?"), role, id)
		if e == nil && password != "" {
			_, e = tx.ExecContext(ctx, SQL("UPDATE {admins} SET password=? WHERE id=?"), password, id)
		}
	}
	if e != nil {
		return e
	}
	return tx.Commit()
}
