// Reconstructed SQL is extracted from the original executable.
package repo

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
)

func NewKamiGroupRepo(db *sqlx.DB) *KamiGroupRepo { return &KamiGroupRepo{db: db} }

func (r *KamiGroupRepo) GetByID(ctx context.Context, appid, id int64) (*KamiGroup, error) {
	var v KamiGroup
	e := r.db.GetContext(ctx, &v, SQL("SELECT "+AppRecordIDColumn("kami_groups", "{kami_groups}")+"id, appid, name, type, value, remark FROM {kami_groups} WHERE appid = ? AND id = ? LIMIT 1"), appid, id)
	if errors.Is(e, sql.ErrNoRows) {
		return nil, nil
	}
	return &v, e
}

func (r *KamiGroupRepo) ListByApp(ctx context.Context, appid int64) ([]KamiGroup, error) {
	v := []KamiGroup{}
	e := r.db.SelectContext(ctx, &v, SQL("SELECT "+AppRecordIDColumn("kami_groups", "{kami_groups}")+"id, appid, name, type, value, remark FROM {kami_groups} WHERE appid = ? ORDER BY id DESC"), appid)
	return v, e
}

func (r *KamiGroupRepo) Create(ctx context.Context, v *KamiGroup) (int64, error) {
	res, e := ExecAppInsert(ctx, r.db, v.AppID, "kami_groups", SQL("INSERT INTO {kami_groups} (appid, name, type, value, remark)\n\t\t VALUES (?,?,?,?,?)"), v.AppID, v.Name, v.Type, v.Value, v.Remark)
	if e != nil {
		return 0, e
	}
	return res.LastInsertId()
}

func (r *KamiGroupRepo) Update(ctx context.Context, v *KamiGroup) error {
	_, e := r.db.ExecContext(ctx, SQL("UPDATE {kami_groups} SET name=?, type=?, value=?, remark=? WHERE appid=? AND id=?"), v.Name, v.Type, v.Value, v.Remark, v.AppID, v.ID)
	return e
}
