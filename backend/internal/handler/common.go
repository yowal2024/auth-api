package handler

import (
	"auth-api/internal/response"
	"github.com/gin-gonic/gin"
	"log"
	"strconv"
)

func requireAppID(c *gin.Context) (int64, bool) {
	id := c.GetInt64("appid")
	if id <= 0 {
		response.BadRequest(c, "缺少 X-App-Id 头")
		return 0, false
	}
	return id, true
}
func parseIDParam(c *gin.Context, name string) (int64, bool) {
	v, e := strconv.ParseInt(c.Param(name), 10, 64)
	if e != nil || v <= 0 {
		response.BadRequest(c, "无效的 ID")
		return 0, false
	}
	return v, true
}
func dbError(c *gin.Context, e error) bool {
	if e == nil {
		return false
	}
	log.Printf("%s %s: %v", c.Request.Method, c.FullPath(), e)
	response.ServerError(c, "数据库操作失败")
	return true
}
func pageParams(c *gin.Context) (int, int) {
	p, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	s, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if p < 1 {
		p = 1
	}
	if s < 1 {
		s = 20
	}
	if s > 200 {
		s = 200
	}
	return p, s
}
