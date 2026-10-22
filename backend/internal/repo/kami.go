package repo

import "context"

func (r *KamiRepo) UpdateBinding(ctx context.Context, a, id int64, v map[string]any) error {
	return updatePartial(ctx, r.db, "kamis", a, id, v, []string{"nickname", "vip", "ban_expire", "bind_limit", "region_limit", "mi", "remark"})
}

// Lock against activation/topup so thawing never makes a consumed card reusable.
func (r *KamiRepo) UpdateAccount(ctx context.Context, a, id int64, v map[string]any, frozen *bool) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var k struct {
		Status  int    `db:"status"`
		UseUser string `db:"use_user"`
		UseTime int64  `db:"use_time"`
	}
	if err = tx.GetContext(ctx, &k, SQL("SELECT status,use_user,use_time FROM {kamis} WHERE appid=? AND id=? FOR UPDATE"), a, id); err != nil {
		return err
	}
	if frozen != nil {
		status := k.Status
		if *frozen {
			status = 2
		} else if status == 2 {
			status = 0
			if k.UseTime != 0 || k.UseUser != "" {
				status = 1
			}
		}
		v["status"] = status
	}
	if err = updatePartial(ctx, tx, "kamis", a, id, v, []string{"nickname", "vip", "ban_expire", "bind_limit", "region_limit", "mi", "remark", "status"}); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *KamiRepo) Delete(ctx context.Context, appid, id int64, usedOnly bool) (int64, error) {
	ids, err := r.DeleteReturningIDs(ctx, appid, id, usedOnly)
	return int64(len(ids)), err
}

func (r *KamiRepo) DeleteReturningIDs(ctx context.Context, appid, id int64, usedOnly bool) ([]int64, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	where := "appid=? AND id=?"
	args := []any{appid, id}
	if usedOnly {
		where = "appid=? AND (status=1 OR (status=2 AND (use_time>0 OR use_user<>'')))"
		args = []any{appid}
	}
	var ids []int64
	if err = tx.SelectContext(ctx, &ids, SQL("SELECT id FROM {kamis} WHERE "+where+" FOR UPDATE"), args...); err != nil {
		return nil, err
	}
	for _, table := range []string{"user_devices", "user_extends", "user_logs"} {
		if _, err = tx.ExecContext(ctx, SQL("DELETE FROM {"+table+"} WHERE uid IN (SELECT -id FROM {kamis} WHERE "+where+")"), args...); err != nil {
			return nil, err
		}
	}
	_, err = tx.ExecContext(ctx, SQL("DELETE FROM {kamis} WHERE "+where), args...)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return ids, nil
}
