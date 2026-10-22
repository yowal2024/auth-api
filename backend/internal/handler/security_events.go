package handler

import (
	"auth-api/internal/repo"
	"auth-api/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func MountSecurityEvents(g *gin.RouterGroup, d *sqlx.DB) {
	g.GET("/system/security/events", func(c *gin.Context) {
		page, size := pageParams(c)
		where := "1=1"
		args := []any{}
		if kind := c.Query("kind"); kind != "" {
			where += " AND kind=?"
			args = append(args, kind)
		}
		if ip := c.Query("ip"); ip != "" {
			where += " AND ip=?"
			args = append(args, ip)
		}
		if app := c.Query("appid"); app != "" {
			where += " AND appid=?"
			args = append(args, app)
		}
		var total int
		if dbError(c, d.GetContext(c.Request.Context(), &total, repo.SQL("SELECT COUNT(*) FROM {security_events} WHERE "+where), args...)) {
			return
		}
		rows := []struct {
			ID         int64  `db:"id" json:"id"`
			AppID      int64  `db:"appid" json:"appid"`
			Kind       string `db:"kind" json:"kind"`
			Summary    string `db:"summary" json:"summary"`
			IP         string `db:"ip" json:"ip"`
			Route      string `db:"route" json:"route"`
			CreateTime int64  `db:"create_time" json:"createTime"`
		}{}
		args = append(args, size, (page-1)*size)
		if !dbError(c, d.SelectContext(c.Request.Context(), &rows, repo.SQL("SELECT id,appid,kind,summary,ip,route,create_time FROM {security_events} WHERE "+where+" ORDER BY id DESC LIMIT ? OFFSET ?"), args...)) {
			response.OK(c, gin.H{"list": rows, "total": total})
		}
	})
}
