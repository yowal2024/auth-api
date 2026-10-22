// Reconstructed SQL is extracted from the original executable.
package repo

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
)

func NewBlockRepo(db *sqlx.DB) *BlockRepo { return &BlockRepo{db: db} }

func (r *BlockRepo) GetByID(ctx context.Context, appid, id int64) (*Block, error) {
	var v Block
	e := r.db.GetContext(ctx, &v, SQL("SELECT `id`,`appid`,`type`,`value`,`reason`,`create_time` FROM {blocks} WHERE appid = ? AND id = ? LIMIT 1"), appid, id)
	if errors.Is(e, sql.ErrNoRows) {
		return nil, nil
	}
	return &v, e
}

func (r *BlockRepo) ListByApp(ctx context.Context, appid int64) ([]Block, error) {
	v := []Block{}
	e := r.db.SelectContext(ctx, &v, SQL("SELECT `id`,`appid`,`type`,`value`,`reason`,`create_time` FROM {blocks} WHERE appid = ? ORDER BY id DESC"), appid)
	return v, e
}

func (r *BlockRepo) Create(ctx context.Context, v *Block) (int64, error) {
	res, e := r.db.ExecContext(ctx, SQL("INSERT INTO {blocks} (appid, type, value, reason, create_time)\n\t\t VALUES (?,?,?,?,?)"), v.AppID, v.Type, v.Value, v.Reason, v.CreateTime)
	if e != nil {
		return 0, e
	}
	return res.LastInsertId()
}

func (r *BlockRepo) Delete(ctx context.Context, appid, id int64) error {
	_, e := r.db.ExecContext(ctx, SQL("DELETE FROM {blocks} WHERE appid = ? AND id = ?"), appid, id)
	return e
}
