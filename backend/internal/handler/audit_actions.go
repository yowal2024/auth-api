package handler

import (
	"encoding/json"
	"strings"
)

// Route patterns, not concrete URLs: supports custom administrator entry names
// and old audit rows without storing request bodies or credentials.
func auditAction(method, route string) string {
	parts := strings.SplitN(strings.TrimPrefix(route, "/"), "/", 3)
	if len(parts) != 3 || parts[0] != "api" {
		return ""
	}
	key := method + " /" + parts[2]
	if label, ok := auditActions[key]; ok {
		return label
	}
	for _, kind := range []string{"routes", "codes"} {
		resource := "接口路径"
		if kind == "codes" {
			resource = "错误码"
		}
		for _, base := range []string{"/system/api-"} {
			switch key {
			case "POST " + base + kind + "/save", "POST " + base + kind + "/:id/override":
				return "修改了" + resource
			case "POST " + base + kind + "/:id/toggle":
				return "调整了" + resource + "重定义开关"
			case "POST " + base + kind + "/randomize-all":
				return "随机重定义了" + resource
			}
		}
	}
	return ""
}

var auditActions = map[string]string{
	"DELETE /user/:id/sessions": "下线了用户全部会话", "DELETE /user/:id/sessions/:sessionId": "下线了用户单个会话",
	"DELETE /kami/:id/sessions": "下线了卡密全部会话", "DELETE /kami/:id/sessions/:sessionId": "下线了卡密单个会话",
	"POST /system/backup/settings": "修改了自动备份设置", "POST /system/backup/run": "创建了备份",

	"POST /geo/local/update": "更新了本地 IP 库",
	"POST /logout":           "退出了后台", "PUT /me/profile": "修改了个人资料", "PUT /me/password": "修改了自己的密码",
	"POST /app": "创建了应用", "PUT /app/:id": "修改了应用设置", "DELETE /app/:id": "删除了应用",
	"POST /user": "创建了用户", "PUT /user/:id": "修改了用户资料", "DELETE /user/:id": "删除了用户", "POST /user/batch-delete": "批量删除了用户",
	"POST /user/reward": "调整了用户会员时长", "POST /user/:id/password": "重置了用户密码", "POST /user/:id/unbind-mi": "解绑了用户设备",
	"POST /user/:id/extends": "保存了用户扩展信息", "DELETE /user/:id/extends/:eid": "删除了用户扩展信息",
	"DELETE /user/:id/devices": "清空了用户绑定设备", "DELETE /user/:id/devices/:deviceId": "移除了用户绑定设备",
	"POST /kami/generate": "生成了卡密", "PUT /kami/:id": "修改了卡密", "DELETE /kami/:id": "删除了卡密", "POST /kami/clean-used": "清理了已用卡密",
	"POST /kami/:id/reward": "调整了卡密会员时长", "POST /kami/:id/unbind-mi": "解绑了卡密设备",
	"DELETE /kami/:id/devices": "清空了卡密绑定设备", "DELETE /kami/:id/devices/:deviceId": "移除了卡密绑定设备",
	"POST /notice": "保存了公告", "DELETE /notice/:id": "删除了公告", "POST /extend": "保存了应用扩展变量", "DELETE /extend/:id": "删除了应用扩展变量",
	"POST /block": "添加了封禁规则", "DELETE /block/:id": "删除了封禁规则",
	"POST /user/group": "保存了用户分组", "DELETE /user/group/:id": "删除了用户分组", "POST /kami/group": "保存了卡密分组", "DELETE /kami/group/:id": "删除了卡密分组",
	"POST /encrypt": "保存了加密方案", "DELETE /encrypt/:id": "删除了加密方案", "POST /function": "保存了云函数", "DELETE /function/:id": "删除了云函数", "POST /function/:id/run": "试运行了云函数",
	"POST /ver": "保存了版本", "DELETE /ver/:id": "删除了版本", "POST /ver/upload-apk": "上传了 APK",
	"POST /upload/avatar": "上传了头像", "POST /upload/app-logo": "上传了应用图标", "POST /extend-file/upload": "上传了扩展文件", "PUT /extend-file/:id": "修改了扩展文件信息", "DELETE /extend-file/:id": "删除了扩展文件",
	"POST /admin": "保存了管理员信息", "DELETE /admin/:id": "删除了管理员", "POST /log/clear": "清理了日志", "POST /system/settings": "修改了系统设置",
	"POST /app/:id/mtls/gen-ca": "生成了应用 CA 证书", "POST /app/:id/mtls/issue-cert": "签发了客户端证书", "POST /system/https/gen-cert": "生成了 HTTPS 证书", "POST /system/https": "修改了 HTTPS 设置",
}

func recentActivity(actor, action, data, kind string) string {
	label := action
	legacy := false
	if method, route, ok := strings.Cut(action, " "); ok && strings.HasPrefix(route, "/api/") {
		label = auditAction(method, route)
		if label == "" {
			label = "执行了后台操作"
		}
		legacy = true
	} else if kind == "primary" {
		switch action {
		case "login":
			label = "登录了客户端"
		case "register":
			label = "注册了账号"
		case "logout":
			label = "退出了客户端"
		case "kamiTopup":
			label = "使用卡密充值了账户"
		}
	}
	var outcome struct {
		Success    *bool `json:"success"`
		HTTPStatus int   `json:"httpStatus"`
	}
	_ = json.Unmarshal([]byte(data), &outcome)
	if (outcome.Success != nil && !*outcome.Success) || outcome.HTTPStatus >= 400 {
		label = strings.Replace(label, "了", "", 1) + "（失败）"
	} else if legacy && outcome.Success == nil {
		// Old HTTP-200 records do not prove business success. Describe the
		// attempted action without claiming it completed.
		label = strings.Replace(label, "了", "", 1)
	}
	if label == action && !strings.ContainsAny(label, "了成") {
		return actor + " · " + label
	}
	return actor + " " + label
}
