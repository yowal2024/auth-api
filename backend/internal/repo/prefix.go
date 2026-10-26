package repo

import (
	"fmt"
	"regexp"
	"strings"
	"sync/atomic"
)

var prefix atomic.Pointer[strings.Replacer]
var tableNames = []string{"admins", "apps", "users", "user_extends", "user_devices", "user_logs", "notices", "extends", "versions", "blocks", "kami_groups", "kamis", "functions", "encrypts", "logs", "settings", "api_routes", "api_codes", "user_groups", "ip_info", "extend_files", "security_events", "backup_settings", "backup_runs", "app_record_ids", "app_record_counters"}

func ValidatePrefix(p string) error {
	if !regexp.MustCompile(`^[a-zA-Z0-9_]{0,32}$`).MatchString(p) {
		return fmt.Errorf("invalid table prefix")
	}
	return nil
}
func InitTablePrefix(p string) error {
	if e := ValidatePrefix(p); e != nil {
		return e
	}
	pairs := []string{}
	for _, n := range tableNames {
		pairs = append(pairs, "{"+n+"}", p+n)
	}
	prefix.Store(strings.NewReplacer(pairs...))
	return nil
}
func SQL(q string) string {
	if p := prefix.Load(); p != nil {
		return p.Replace(q)
	}
	return strings.NewReplacer("{", "", "}", "").Replace(q)
}
