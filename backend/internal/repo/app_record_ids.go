package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

func appRecordEntity(entity string) bool {
	switch entity {
	case "users", "user_groups", "kamis", "kami_groups", "versions", "notices",
		"extends", "extend_files", "blocks", "encrypts", "functions":
		return true
	}
	return false
}

// AppRecordIDExpr uses only source-controlled table names and aliases. The
// indexed lookup preserves all existing primary keys, joins and SDK identifiers.
func AppRecordIDExpr(entity, alias string) string {
	if !appRecordEntity(entity) {
		panic("unsupported application record entity: " + entity)
	}
	return "COALESCE((SELECT rid.local_id FROM {app_record_ids} rid WHERE rid.appid=" + alias + ".appid AND rid.entity='" + entity + "' AND rid.record_id=" + alias + ".id),0)"
}

func AppRecordIDColumn(entity, alias string) string {
	return AppRecordIDExpr(entity, alias) + " AS app_record_id,"
}

// AssignAppRecordIDs runs in the same transaction as record creation. The
// counter row serializes concurrent inserts per application and entity. Keep
// mappings after record deletion so a previously visible number is never reused.
func AssignAppRecordIDs(ctx context.Context, tx *sqlx.Tx, appID int64, entity string, ids []int64) error {
	if appID <= 0 || !appRecordEntity(entity) {
		return fmt.Errorf("invalid application record scope")
	}
	if len(ids) == 0 {
		return nil
	}
	for _, id := range ids {
		if id <= 0 {
			return fmt.Errorf("invalid application record ID")
		}
	}
	count := int64(len(ids))
	_, err := tx.ExecContext(ctx, SQL("INSERT INTO {app_record_counters} (appid,entity,last_id) VALUES (?,?,?) ON DUPLICATE KEY UPDATE last_id=last_id+?"), appID, entity, count, count)
	if err != nil {
		return err
	}
	var last int64
	if err = tx.GetContext(ctx, &last, SQL("SELECT last_id FROM {app_record_counters} WHERE appid=? AND entity=?"), appID, entity); err != nil {
		return err
	}
	marks := make([]string, len(ids))
	args := make([]any, 0, len(ids)*4)
	for i, id := range ids {
		marks[i] = "(?,?,?,?)"
		args = append(args, appID, entity, id, last-count+1+int64(i))
	}
	_, err = tx.ExecContext(ctx, SQL("INSERT INTO {app_record_ids} (appid,entity,record_id,local_id) VALUES "+strings.Join(marks, ",")), args...)
	return err
}

// ExecAppInsert commits a new record and its permanent local number together.
func ExecAppInsert(ctx context.Context, db *sqlx.DB, appID int64, entity, query string, args ...any) (sql.Result, error) {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, SQL(query), args...)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	if err = AssignAppRecordIDs(ctx, tx, appID, entity, []int64{id}); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return result, nil
}
