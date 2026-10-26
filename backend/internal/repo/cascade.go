package repo

import (
	"context"
	"github.com/jmoiron/sqlx"
)

func (r *AppRepo) Delete(ctx context.Context, id int64) error {
	tx, e := r.db.BeginTxx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	for _, table := range []string{"user_extends", "user_devices", "user_logs"} {
		if _, e = tx.ExecContext(ctx, SQL("DELETE FROM {"+table+"} WHERE uid IN (SELECT id FROM {users} WHERE appid=?) OR uid IN (SELECT -id FROM {kamis} WHERE appid=?)"), id, id); e != nil {
			return e
		}
	}
	for _, table := range []string{"users", "user_groups", "kami_groups", "kamis", "notices", "extends", "extend_files", "versions", "blocks", "functions", "encrypts", "app_record_ids", "app_record_counters"} {
		if _, e = tx.ExecContext(ctx, SQL("DELETE FROM {"+table+"} WHERE appid=?"), id); e != nil {
			return e
		}
	}
	if _, e = tx.ExecContext(ctx, SQL("DELETE FROM {apps} WHERE id=?"), id); e != nil {
		return e
	}
	return tx.Commit()
}
func (r *KamiGroupRepo) Delete(ctx context.Context, appid, id int64) error {
	_, err := r.DeleteReturningIDs(ctx, appid, id)
	return err
}

func (r *KamiGroupRepo) DeleteReturningIDs(ctx context.Context, appid, id int64) ([]int64, error) {
	tx, e := r.db.BeginTxx(ctx, nil)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback()
	var groupID int64
	if e = tx.GetContext(ctx, &groupID, SQL("SELECT id FROM {kami_groups} WHERE appid=? AND id=? FOR UPDATE"), appid, id); e != nil {
		return nil, e
	}
	var ids []int64
	if e = tx.SelectContext(ctx, &ids, SQL("SELECT -id FROM {kamis} WHERE appid=? AND group_id=? FOR UPDATE"), appid, id); e != nil {
		return nil, e
	}
	for _, table := range []string{"user_devices", "user_extends", "user_logs"} {
		if _, e = tx.ExecContext(ctx, SQL("DELETE FROM {"+table+"} WHERE uid IN (SELECT -id FROM {kamis} WHERE appid=? AND group_id=?)"), appid, id); e != nil {
			return nil, e
		}
	}
	if _, e = tx.ExecContext(ctx, SQL("DELETE FROM {kamis} WHERE appid=? AND group_id=?"), appid, id); e != nil {
		return nil, e
	}
	if _, e = tx.ExecContext(ctx, SQL("DELETE FROM {kami_groups} WHERE appid=? AND id=?"), appid, id); e != nil {
		return nil, e
	}
	return ids, tx.Commit()
}
func (r *UserGroupRepo) CountUsers(ctx context.Context, appid, id int64) (int, error) {
	var n int
	e := r.db.GetContext(ctx, &n, SQL("SELECT COUNT(*) FROM {users} WHERE appid=? AND group_id=?"), appid, id)
	return n, e
}
func NewKamiRepo(d *sqlx.DB) *KamiRepo { return &KamiRepo{d} }
