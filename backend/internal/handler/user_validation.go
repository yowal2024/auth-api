package handler

import (
	"auth-api/internal/geo"
	"encoding/json"
	"fmt"
	"net"
	"strings"
)

func validateAccountLimits(bindLimit, regionLimit *int) error {
	if (bindLimit != nil && (*bindLimit < 0 || *bindLimit > 2147483646)) || (regionLimit != nil && (*regionLimit < 0 || *regionLimit > 2147483647)) {
		return fmt.Errorf("设备和地域名额必须为有效的非负整数")
	}
	return nil
}

func validateIPList(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var entries []geo.IPEntry
	if json.Unmarshal([]byte(raw), &entries) != nil {
		return fmt.Errorf("地域记录必须为 JSON 数组")
	}
	for _, entry := range entries {
		if net.ParseIP(entry.IP) == nil || entry.Lat < -90 || entry.Lat > 90 || entry.Lng < -180 || entry.Lng > 180 || entry.Time < 0 {
			return fmt.Errorf("地域记录的 IP、经纬度或时间无效")
		}
	}
	return nil
}

func spreadsheetSafe(s string) string {
	trimmed := strings.TrimLeft(s, " \t\r\n")
	if strings.ContainsAny(s, "\t\r\n") || (len(trimmed) > 0 && strings.ContainsRune("=+-@", rune(trimmed[0]))) {
		return "'" + s
	}
	return s
}
