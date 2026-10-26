// Reconstructed SQL is extracted from the original executable.
package repo

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
)

func NewNoticeRepo(db *sqlx.DB) *NoticeRepo { return &NoticeRepo{db: db} }

func (r *NoticeRepo) GetByID(ctx context.Context, appid, id int64) (*Notice, error) {
	var v Notice
	e := r.db.GetContext(ctx, &v, SQL("SELECT "+AppRecordIDColumn("notices", "{notices}")+"id, appid, title, content, enabled, admin_id, author, views, update_time FROM {notices} WHERE appid = ? AND id = ? LIMIT 1"), appid, id)
	if errors.Is(e, sql.ErrNoRows) {
		return nil, nil
	}
	return &v, e
}

func (r *NoticeRepo) ListByApp(ctx context.Context, appid int64) ([]Notice, error) {
	v := []Notice{}
	e := r.db.SelectContext(ctx, &v, SQL("SELECT "+AppRecordIDColumn("notices", "n")+"n.id, n.appid, n.title, n.content, n.enabled, n.admin_id,\n\t\t        COALESCE(NULLIF(a.nickname, ''), a.username, n.author) AS author,\n\t\t        n.views, n.update_time\n\t\t FROM {notices} n\n\t\t LEFT JOIN {admins} a ON a.id = n.admin_id\n\t\t WHERE n.appid = ?\n\t\t ORDER BY n.update_time DESC, n.id DESC"), appid)
	return v, e
}

func (r *NoticeRepo) Create(ctx context.Context, v *Notice) (int64, error) {
	res, e := ExecAppInsert(ctx, r.db, v.AppID, "notices", SQL("INSERT INTO {notices} (appid, title, content, enabled, admin_id, author, views, update_time)\n\t\t VALUES (?,?,?,?,?,?,?,?)"), v.AppID, v.Title, v.Content, v.Enabled, v.AdminID, v.Author, v.Views, v.UpdateTime)
	if e != nil {
		return 0, e
	}
	return res.LastInsertId()
}

func (r *NoticeRepo) Update(ctx context.Context, v *Notice) error {
	_, e := r.db.ExecContext(ctx, SQL("UPDATE {notices} SET title=?, content=?, enabled=?, update_time=?\n\t\t WHERE appid=? AND id=?"), v.Title, v.Content, v.Enabled, v.UpdateTime, v.AppID, v.ID)
	return e
}

func (r *NoticeRepo) Delete(ctx context.Context, appid, id int64) error {
	_, e := r.db.ExecContext(ctx, SQL("DELETE FROM {notices} WHERE appid = ? AND id = ?"), appid, id)
	return e
}
