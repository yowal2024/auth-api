package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"sort"
	"strings"
	"time"
)

var userSelect = "SELECT " + AppRecordIDColumn("users", "u") + `(COALESCE(OCTET_LENGTH(u.sign_pk8),0)>0 AND COALESCE(OCTET_LENGTH(u.sign_x509),0)>0) AS has_sign_key,u.id,u.appid,u.group_id,u.username,u.password,u.nickname,u.avatar,u.vip,u.mi,u.region,u.register_ip,u.last_ip,u.last_region,u.last_login_time,u.last_active_time,u.bind_limit,u.region_limit,COALESCE(u.ip_list,'') AS ip_list,u.ban_expire,u.create_time,COALESCE(g.name,'') AS group_name,COALESCE(g.level,0) AS group_level FROM {users} u LEFT JOIN {user_groups} g ON g.id=u.group_id`

func NewUserRepo(d *sqlx.DB) *UserRepo { return &UserRepo{d} }

var ErrUsernameExists = errors.New("该应用下账号已存在")

func (r *UserRepo) GetByID(ctx context.Context, a, id int64) (*User, error) {
	var v User
	e := r.db.GetContext(ctx, &v, SQL(userSelect+" WHERE u.appid=? AND u.id=? LIMIT 1"), a, id)
	if errors.Is(e, sql.ErrNoRows) {
		return nil, nil
	}
	return &v, e
}

type UserFilter struct {
	Page, Size                       int
	GroupID                          int64
	Field, Keyword, Status, UserType string
}

func (r *UserRepo) List(ctx context.Context, a int64, f UserFilter) ([]User, int, error) {
	where := "u.appid=?"
	args := []any{a}
	if f.GroupID > 0 {
		where += " AND u.group_id=?"
		args = append(args, f.GroupID)
	}
	if f.Keyword != "" {
		allowed := map[string]string{"username": "u.username", "nickname": "u.nickname", "mi": "u.mi", "id": "u.id", "ip": "u.last_ip", "lastIp": "u.last_ip"}
		allowed["appRecordId"] = AppRecordIDExpr("users", "u")
		col := allowed[f.Field]
		if col == "" {
			col = "u.username"
		}
		where += " AND " + col + " LIKE ?"
		args = append(args, "%"+f.Keyword+"%")
	}
	now := time.Now().Unix()
	switch f.Status {
	case "banned", "0":
		where += " AND (u.ban_expire=-1 OR u.ban_expire>?)"
		args = append(args, now)
	case "normal", "1":
		where += " AND u.ban_expire>=0 AND u.ban_expire<=?"
		args = append(args, now)
	}
	switch f.UserType {
	case "vip":
		where += " AND u.vip>?"
		args = append(args, now)
	case "normal", "free":
		where += " AND u.vip<=?"
		args = append(args, now)
	}
	var total int
	e := r.db.GetContext(ctx, &total, SQL("SELECT COUNT(*) FROM {users} u WHERE "+where), args...)
	if e != nil {
		return nil, 0, e
	}
	v := []User{}
	args = append(args, f.Size, (f.Page-1)*f.Size)
	e = r.db.SelectContext(ctx, &v, SQL(userSelect+" WHERE "+where+" ORDER BY u.id DESC LIMIT ? OFFSET ?"), args...)
	return v, total, e
}
func (r *UserRepo) Create(ctx context.Context, v *User) (int64, error) {
	return r.CreateWith(ctx, v, nil)
}
func (r *UserRepo) CreateWith(ctx context.Context, v *User, initialize func(*sqlx.Tx) error) (int64, error) {
	// Lock the parent app to serialize username checks without inventing a schema index.
	tx, e := r.db.BeginTxx(ctx, nil)
	if e != nil {
		return 0, e
	}
	defer tx.Rollback()
	var app int64
	if e = tx.GetContext(ctx, &app, SQL("SELECT id FROM {apps} WHERE id=? FOR UPDATE"), v.AppID); e != nil {
		return 0, e
	}
	var n int
	if e = tx.GetContext(ctx, &n, SQL("SELECT COUNT(*) FROM {users} WHERE appid=? AND username=?"), v.AppID, v.Username); e != nil {
		return 0, e
	}
	if n > 0 {
		return 0, ErrUsernameExists
	}
	result, e := tx.ExecContext(ctx, SQL("INSERT INTO {users} (appid,group_id,username,password,nickname,avatar,vip,mi,region,register_ip,last_ip,last_region,last_login_time,bind_limit,region_limit,ip_list,ban_expire,create_time) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"), v.AppID, v.GroupID, v.Username, v.Password, v.Nickname, v.Avatar, v.Vip, v.Mi, v.Region, v.RegisterIP, v.LastIP, v.LastRegion, v.LastLoginTime, v.BindLimit, v.RegionLimit, v.IPList, v.BanExpire, v.CreateTime)
	if e != nil {
		return 0, e
	}
	id, e := result.LastInsertId()
	if e != nil {
		return 0, e
	}
	if e = AssignAppRecordIDs(ctx, tx, v.AppID, "users", []int64{id}); e != nil {
		return 0, e
	}
	v.ID = id
	if initialize != nil {
		if e = initialize(tx); e != nil {
			return 0, e
		}
	}
	return id, tx.Commit()
}
func (r *UserRepo) UpdatePartial(ctx context.Context, a, id int64, values map[string]any) error {
	return updatePartial(ctx, r.db, "users", a, id, values, []string{"password", "nickname", "avatar", "vip", "mi", "region", "bind_limit", "region_limit", "ip_list", "ban_expire", "group_id"})
}
func updatePartial(ctx context.Context, d sqlx.ExecerContext, table string, a, id int64, v map[string]any, allowed []string) error {
	keys := []string{}
	for k := range v {
		ok := false
		for _, f := range allowed {
			if k == f {
				ok = true
				break
			}
		}
		if !ok {
			return fmt.Errorf("unsupported field %s", k)
		}
		keys = append(keys, k)
	}
	if len(keys) == 0 {
		return nil
	}
	sort.Strings(keys)
	sets := []string{}
	args := []any{}
	for _, k := range keys {
		sets = append(sets, "`"+k+"`=?")
		args = append(args, v[k])
	}
	args = append(args, a, id)
	_, e := d.ExecContext(ctx, SQL("UPDATE {"+table+"} SET "+strings.Join(sets, ",")+" WHERE appid=? AND id=?"), args...)
	return e
}
func (r *UserRepo) BatchDelete(ctx context.Context, a int64, ids []int64) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	tx, e := r.db.BeginTxx(ctx, nil)
	if e != nil {
		return 0, e
	}
	defer tx.Rollback()
	query, args, e := sqlx.In(SQL("SELECT id FROM {users} WHERE appid=? AND id IN (?) FOR UPDATE"), a, ids)
	if e != nil {
		return 0, e
	}
	owned := []int64{}
	if e = tx.SelectContext(ctx, &owned, query, args...); e != nil {
		return 0, e
	}
	if len(owned) == 0 {
		return 0, nil
	}
	for _, table := range []string{"user_extends", "user_devices", "user_logs"} {
		q, v, err := sqlx.In(SQL("DELETE FROM {"+table+"} WHERE uid IN (?)"), owned)
		if err != nil {
			return 0, err
		}
		if _, e = tx.ExecContext(ctx, q, v...); e != nil {
			return 0, e
		}
	}
	q, v, e := sqlx.In(SQL("DELETE FROM {users} WHERE appid=? AND id IN (?)"), a, owned)
	if e != nil {
		return 0, e
	}
	res, e := tx.ExecContext(ctx, q, v...)
	if e != nil {
		return 0, e
	}
	n, e := res.RowsAffected()
	if e != nil {
		return 0, e
	}
	return n, tx.Commit()
}
func (r *UserRepo) Delete(ctx context.Context, a, id int64) error {
	_, e := r.BatchDelete(ctx, a, []int64{id})
	return e
}
func (r *UserRepo) Reward(ctx context.Context, a, seconds int64, target string, ids []int64) (int64, error) {
	now := time.Now().Unix()
	q := "UPDATE {users} SET vip=GREATEST(vip,?)+? WHERE appid=?"
	args := []any{now, seconds, a}
	switch target {
	case "all":
	case "vip":
		q += " AND vip>?"
		args = append(args, now)
	case "selected":
		if len(ids) == 0 {
			return 0, errors.New("请选择用户")
		}
		q += " AND id IN (?)"
		args = append(args, ids)
	default:
		return 0, errors.New("无效奖励范围")
	}
	q, args, e := sqlx.In(SQL(q), args...)
	if e != nil {
		return 0, e
	}
	res, e := r.db.ExecContext(ctx, q, args...)
	if e != nil {
		return 0, e
	}
	return res.RowsAffected()
}
