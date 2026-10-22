package response

import "github.com/gin-gonic/gin"

// Keep the business result available to audit middleware without capturing
// response bodies (which may contain passwords, keys or uploaded content).
const CodeContextKey = "response.businessCode"

func OK(c *gin.Context, data any) {
	c.Set(CodeContextKey, 0)
	c.JSON(200, Envelope{Code: 0, Msg: "ok", Data: data})
}
func Fail(c *gin.Context, code int, msg string) {
	c.Set(CodeContextKey, code)
	c.JSON(200, Envelope{Code: code, Msg: msg})
}
func FailData(c *gin.Context, code int, msg string, data any) {
	c.Set(CodeContextKey, code)
	c.JSON(200, Envelope{Code: code, Msg: msg, Data: data})
}
func Unauthorized(c *gin.Context)            { Fail(c, 401, "未登录或登录已过期") }
func BadRequest(c *gin.Context, msg string)  { Fail(c, 400, msg) }
func ServerError(c *gin.Context, msg string) { Fail(c, 500, msg) }
