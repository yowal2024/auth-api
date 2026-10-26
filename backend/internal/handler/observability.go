package handler

import (
	"auth-api/internal/middleware"
	"auth-api/internal/repo"
	"auth-api/internal/response"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"sort"
	"strconv"
	"strings"
	"time"
)

func MountObservability(g *gin.RouterGroup, d *sqlx.DB, r *redis.Client) {
	g.GET("/log/list", func(c *gin.Context) {
		page, size := pageParams(c)
		category := c.DefaultQuery("category", "admin")
		if category != "admin" && category != "user" {
			response.BadRequest(c, "日志分类无效")
			return
		}
		if category == "user" && c.GetInt64("appid") <= 0 {
			response.BadRequest(c, "请选择应用")
			return
		}
		args := []any{}
		where := "1=1"
		if category == "admin" && c.Query("type") != "system" {
			where = "l.type='admin' AND l.appid=?"
			args = append(args, max(c.GetInt64("appid"), 0))
		}
		if v := c.Query("type"); v != "" {
			where += " AND l.type=?"
			args = append(args, v)
		}
		var start, end int64
		for _, filter := range []struct {
			name string
			op   string
			dest *int64
		}{{"start", ">=", &start}, {"end", "<=", &end}} {
			if raw := c.Query(filter.name); raw != "" {
				value, err := strconv.ParseInt(raw, 10, 64)
				if err != nil || value < 0 {
					response.BadRequest(c, "日志时间范围无效")
					return
				}
				*filter.dest = value
				where += " AND l.create_time" + filter.op + "?"
				args = append(args, value)
			}
		}
		if c.Query("end") != "" && start > end {
			response.BadRequest(c, "日志时间范围无效")
			return
		}
		if v := c.Query("action"); v != "" {
			where += " AND l.action=?"
			args = append(args, v)
		}
		if v := c.Query("keyword"); v != "" {
			if category == "user" {
				where += " AND (l.action LIKE ? OR l.data LIKE ? OR l.uid IN (SELECT id FROM {users} WHERE username LIKE ? OR nickname LIKE ?) OR l.uid IN (SELECT -id FROM {kamis} WHERE card LIKE ? OR nickname LIKE ?))"
				for i := 0; i < 6; i++ {
					args = append(args, "%"+v+"%")
				}
			} else {
				where += " AND (l.action LIKE ? OR l.data LIKE ? OR l.operator LIKE ? OR l.username LIKE ?)"
				args = append(args, "%"+v+"%", "%"+v+"%", "%"+v+"%", "%"+v+"%")
			}
		}
		table := "logs"
		if category == "user" {
			table = "user_logs"
			if a := c.GetInt64("appid"); a > 0 {
				where += " AND (l.uid IN (SELECT id FROM {users} WHERE appid=?) OR l.uid IN (SELECT -id FROM {kamis} WHERE appid=?))"
				args = append(args, a, a)
			}
		}
		var total int
		if dbError(c, d.GetContext(c.Request.Context(), &total, repo.SQL("SELECT COUNT(*) FROM {"+table+"} l WHERE "+where), args...)) {
			return
		}
		args = append(args, size, (page-1)*size)
		if category == "user" {
			v := []repo.UserLogListItem{}
			q := "SELECT l.id,COALESCE(u.username,k.card,'') AS username,COALESCE(u.nickname,k.nickname,'') AS nickname,COALESCE(u.avatar,'') AS avatar,l.action,COALESCE(l.data,'') AS data,l.ip,l.region,l.create_time FROM {user_logs} l LEFT JOIN {users} u ON u.id=l.uid LEFT JOIN {kamis} k ON -k.id=l.uid WHERE " + where + " ORDER BY l.id DESC LIMIT ? OFFSET ?"
			if !dbError(c, d.SelectContext(c.Request.Context(), &v, repo.SQL(q), args...)) {
				response.OK(c, gin.H{"list": v, "total": total})
			}
		} else {
			v := []repo.Log{}
			q := "SELECT l.id,l.appid,l.type,l.action,COALESCE(l.data,'') AS data,l.target,l.operator,l.uid,l.ip,l.region,l.username,l.nickname,l.avatar,l.create_time FROM {logs} l WHERE " + where + " ORDER BY l.id DESC LIMIT ? OFFSET ?"
			if !dbError(c, d.SelectContext(c.Request.Context(), &v, repo.SQL(q), args...)) {
				for i := range v {
					if v[i].Type == "admin" || v[i].Type == "system" {
						actor := v[i].Operator
						if actor == "" {
							actor = "管理员"
						}
						v[i].Summary = recentActivity(actor, v[i].Action, v[i].Data, "success")
					}
				}
				response.OK(c, gin.H{"list": v, "total": total})
			}
		}
	})
	g.GET("/log/actions", func(c *gin.Context) {
		if category := c.DefaultQuery("category", "admin"); category != "admin" && category != "user" {
			response.BadRequest(c, "日志分类无效")
			return
		}
		table := "logs"
		where := "1=1"
		args := []any{}
		if c.Query("category") == "user" {
			if c.GetInt64("appid") <= 0 {
				response.BadRequest(c, "请选择应用")
				return
			}
			table = "user_logs"
			where += " AND (uid IN (SELECT id FROM {users} WHERE appid=?) OR uid IN (SELECT -id FROM {kamis} WHERE appid=?))"
			args = append(args, c.GetInt64("appid"), c.GetInt64("appid"))
		} else if c.Query("type") == "system" {
			where = "type='system'"
		} else {
			where = "type='admin' AND appid=?"
			args = append(args, max(c.GetInt64("appid"), 0))
		}
		v := []string{}
		if !dbError(c, d.SelectContext(c.Request.Context(), &v, repo.SQL("SELECT DISTINCT action FROM {"+table+"} WHERE "+where+" ORDER BY action"), args...)) {
			response.OK(c, v)
		}
	})
	g.POST("/log/clear", func(c *gin.Context) {
		if middleware.CurrentAdmin(c).Role != "super" {
			response.Fail(c, 403, "需要超级管理员权限")
			return
		}
		var v logClearReq
		if c.ShouldBindJSON(&v) != nil {
			response.BadRequest(c, "请求体格式错误")
			return
		}
		table := "logs"
		if v.Category != "" && v.Category != "admin" && v.Category != "user" {
			response.BadRequest(c, "日志分类无效")
			return
		}
		if v.Category == "user" {
			table = "user_logs"
		}
		q := "DELETE FROM {" + table + "} WHERE 1=1"
		args := []any{}
		if v.Category == "user" {
			if c.GetInt64("appid") <= 0 {
				response.BadRequest(c, "请选择应用")
				return
			}
			q += " AND (uid IN (SELECT id FROM {users} WHERE appid=?) OR uid IN (SELECT -id FROM {kamis} WHERE appid=?))"
			args = append(args, c.GetInt64("appid"), c.GetInt64("appid"))
		} else if v.Type != "system" {
			q += " AND type='admin' AND appid=?"
			args = append(args, max(c.GetInt64("appid"), 0))
		}
		if len(v.IDs) > 0 {
			q += " AND id IN (?)"
			args = append(args, v.IDs)
		}
		if v.Before > 0 {
			q += " AND create_time<?"
			args = append(args, v.Before)
		}
		if v.Type != "" {
			q += " AND type=?"
			args = append(args, v.Type)
		}
		q, args, e := sqlx.In(repo.SQL(q), args...)
		if dbError(c, e) {
			return
		}
		res, e := d.ExecContext(c.Request.Context(), q, args...)
		if dbError(c, e) {
			return
		}
		n, e := res.RowsAffected()
		if !dbError(c, e) {
			response.OK(c, gin.H{"affected": n})
		}
	})
	g.GET("/dashboard", func(c *gin.Context) {
		days := 7
		switch c.DefaultQuery("range", "7") {
		case "7":
		case "30":
			days = 30
		default:
			response.BadRequest(c, "统计范围仅支持 7 天或 30 天")
			return
		}
		a := c.GetInt64("appid")
		now := time.Now()
		midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
		v := dashboardResp{Trend: dashboardTrend{Dates: []string{}, Logins: []int{}, Regs: []int{}}, Recent: []dashboardRecent{}}
		scope := ""
		args := []any{}
		if a > 0 {
			scope = " AND appid=?"
			args = append(args, a)
		}
		for _, item := range []struct {
			q    string
			dest *int
			args []any
		}{{"SELECT COUNT(*) FROM {users} WHERE 1=1" + scope, &v.Users, args}, {"SELECT COUNT(*) FROM {users} WHERE create_time>=?" + scope, &v.UsersToday, append([]any{midnight}, args...)}, {"SELECT COUNT(*) FROM {kamis} WHERE status=0" + scope, &v.KamiAvail, args}, {"SELECT COUNT(*) FROM {kamis} WHERE status=1 AND use_time>=?" + scope, &v.KamiUsedToday, append([]any{midnight}, args...)}} {
			if dbError(c, d.GetContext(c.Request.Context(), item.dest, repo.SQL(item.q), item.args...)) {
				return
			}
		}
		if r != nil {
			presence := fmt.Sprintf("presence:%d", a)
			if err := r.ZRemRangeByScore(c.Request.Context(), presence, "-inf", strconv.FormatInt(now.Unix(), 10)).Err(); dbError(c, err) {
				return
			}
			online, err := r.ZCard(c.Request.Context(), presence).Result()
			if dbError(c, err) {
				return
			}
			v.OnlineNow = int(online)
			v.APITotal, _ = r.Get(c.Request.Context(), fmt.Sprintf("api:total:%d", a)).Int64()
			v.QueryToday, _ = r.Get(c.Request.Context(), fmt.Sprintf("api:day:%d:%s", a, now.Format("2006-01-02"))).Int64()
			v.OnlinePeakToday, _ = r.Get(c.Request.Context(), fmt.Sprintf("online:peak:%d:%s", a, now.Format("2006-01-02"))).Int()
		}
		for day := days - 1; day >= 0; day-- {
			date := now.AddDate(0, 0, -day)
			start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
			var regs, logins int
			q := "SELECT COUNT(*) FROM {users} WHERE create_time>=? AND create_time<?" + scope
			if dbError(c, d.GetContext(c.Request.Context(), &regs, repo.SQL(q), append([]any{start.Unix(), start.AddDate(0, 0, 1).Unix()}, args...)...)) {
				return
			}
			q = "SELECT COUNT(*) FROM {user_logs} WHERE action='login' AND create_time>=? AND create_time<?"
			la := []any{start.Unix(), start.AddDate(0, 0, 1).Unix()}
			if a > 0 {
				q += " AND (uid IN (SELECT id FROM {users} WHERE appid=?) OR uid IN (SELECT -id FROM {kamis} WHERE appid=?))"
				la = append(la, a, a)
			}
			if dbError(c, d.GetContext(c.Request.Context(), &logins, repo.SQL(q), la...)) {
				return
			}
			v.Trend.Dates = append(v.Trend.Dates, start.Format("01-02"))
			v.Trend.Regs = append(v.Trend.Regs, regs)
			v.Trend.Logins = append(v.Trend.Logins, logins)
		}
		recent := []struct {
			Username string `db:"username"`
			Action   string `db:"action"`
			Time     int64  `db:"create_time"`
			Type     string `db:"type"`
			Data     string `db:"data"`
		}{}
		// Both kinds of activity belong to the selected application. Unassigned
		// historical and global administrator events stay in the global scope.
		if a > 0 {
			q := "SELECT COALESCE(u.username,k.card,'') AS username,l.action,COALESCE(l.data,'') AS data,l.create_time,'primary' AS type FROM {user_logs} l LEFT JOIN {users} u ON u.id=l.uid LEFT JOIN {kamis} k ON -k.id=l.uid WHERE (u.appid=? OR k.appid=?) ORDER BY l.create_time DESC,l.id DESC LIMIT 8"
			if dbError(c, d.SelectContext(c.Request.Context(), &recent, repo.SQL(q), a, a)) {
				return
			}
		}
		adminRecent := recent[:0:0]
		q := "SELECT COALESCE(NULLIF(operator,''),'管理员') AS username,action,COALESCE(data,'') AS data,create_time,'success' AS type FROM {logs} WHERE type='admin' AND appid=? ORDER BY create_time DESC,id DESC LIMIT 8"
		if actor := middleware.CurrentAdmin(c); actor != nil && actor.Role == "super" {
			if dbError(c, d.SelectContext(c.Request.Context(), &adminRecent, repo.SQL(q), max(a, 0))) {
				return
			}
		}
		recent = append(recent, adminRecent...)
		sort.SliceStable(recent, func(i, j int) bool { return recent[i].Time > recent[j].Time })
		if len(recent) > 8 {
			recent = recent[:8]
		}
		for _, item := range recent {
			v.Recent = append(v.Recent, dashboardRecent{Text: recentActivity(item.Username, item.Action, item.Data, item.Type), Time: time.Unix(item.Time, 0).Format("2006-01-02 15:04:05"), Type: item.Type})
		}
		response.OK(c, v)
	})
	g.GET("/user/export", func(c *gin.Context) {
		a, ok := requireAppID(c)
		if !ok {
			return
		}
		q := "SELECT " + repo.AppRecordIDColumn("users", "{users}") + "id,username,nickname,vip,ban_expire,create_time FROM {users} WHERE appid=?"
		args := []any{a}
		switch c.DefaultQuery("scope", "all") {
		case "all":
		case "group":
			q += " AND group_id=?"
			args = append(args, c.Query("groupId"))
		case "vip":
			q += " AND vip>?"
			args = append(args, time.Now().Unix())
		case "banned":
			q += " AND (ban_expire<0 OR ban_expire>?)"
			args = append(args, time.Now().Unix())
		default:
			response.BadRequest(c, "不支持的导出范围")
			return
		}
		v := []repo.User{}
		if dbError(c, d.SelectContext(c.Request.Context(), &v, repo.SQL(q+" ORDER BY id"), args...)) {
			return
		}
		format := c.DefaultQuery("format", "csv")
		if format != "csv" && format != "txt" {
			response.BadRequest(c, "不支持的导出格式")
			return
		}
		c.Header("Content-Disposition", `attachment; filename="users_`+time.Now().Format("20060102_150405")+"."+format+`"`)
		c.Header("Content-Type", "text/plain; charset=utf-8")
		writer := csv.NewWriter(c.Writer)
		if format == "txt" {
			writer.Comma = '\t'
		}
		writer.Write([]string{"应用内编号", "系统 ID", "账号", "昵称", "VIP到期", "封禁到期", "注册时间"})

		for _, u := range v {
			writer.Write([]string{strconv.FormatInt(u.AppRecordID, 10), strconv.FormatInt(u.ID, 10), spreadsheetSafe(u.Username), spreadsheetSafe(u.Nickname), strconv.FormatInt(u.Vip, 10), strconv.FormatInt(u.BanExpire, 10), strconv.FormatInt(u.CreateTime, 10)})
		}
		writer.Flush()
	})
}
func AdminAudit(d *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" {
			return
		}
		a := middleware.CurrentAdmin(c)
		if a == nil {
			return
		}
		var enabled int
		if d.GetContext(c.Request.Context(), &enabled, repo.SQL("SELECT admin_log_enabled FROM {settings} WHERE id=1")) != nil {
			return
		}
		system := isSystemOperation(c.FullPath())
		if enabled == 0 && !system {
			return
		}
		appID := auditAppID(c)
		details := gin.H{"httpStatus": c.Writer.Status(), "method": c.Request.Method, "route": c.FullPath(), "appId": appID}
		if code, exists := c.Get(response.CodeContextKey); exists {
			details["code"] = code
			details["success"] = code == 0 && c.Writer.Status() < 400 && !c.GetBool("audit.executionFailed") && !c.GetBool("audit.hotSwapFailed")
		}
		action := auditAction(c.Request.Method, c.FullPath())
		if action == "" {
			action = "执行了后台操作"
		}
		data, _ := json.Marshal(details)
		write := func(kind string) {
			_, e := d.ExecContext(c.Request.Context(), repo.SQL("INSERT INTO {logs} (type,appid,action,data,target,operator,uid,ip,region,username,nickname,avatar,create_time) VALUES ('"+kind+"',?,?,?,?,?,?,?,'',?,?,?,?)"), appID, action, string(data), c.Param("id"), a.Username, a.AdminID, c.ClientIP(), a.Username, a.Nickname, a.Avatar, time.Now().Unix())
			if e != nil {
				c.Error(e)
			}
		}
		if enabled != 0 && !system {
			write("admin")
		}
		if system {
			var level string
			if e := d.GetContext(c.Request.Context(), &level, repo.SQL("SELECT log_level FROM {settings} WHERE id=1")); e != nil {
				c.Error(e)
				return
			}
			if systemLogAllowed(level, c.GetInt(response.CodeContextKey), c.Writer.Status(), c.GetBool("audit.executionFailed") || c.GetBool("audit.hotSwapFailed")) {
				write("system")
			}
		}
	}
}

func isSystemOperation(route string) bool {
	parts := strings.SplitN(strings.TrimPrefix(route, "/"), "/", 3)
	if len(parts) != 3 || parts[0] != "api" {
		return false
	}
	return strings.HasPrefix(parts[2], "system/") || parts[2] == "geo/local/update"
}
func systemLogAllowed(level string, code, status int, failed bool) bool {
	switch level {
	case "warning":
		return code != 0 || status >= 400 || failed
	case "error":
		return (code >= 500 && code < 600) || status >= 500 || failed
	default:
		return true
	}
}
