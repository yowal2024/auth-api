package server

import (
	"auth-api/internal/repo"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
)

func testLogScope(t *testing.T, d *sqlx.DB, appID *int64, foreignID int64, token string, request func(string, string, any, string, bool) map[string]any) {
	t.Helper()
	original := *appID
	defer func() { *appID = original }()
	execID := func(q string, args ...any) int64 {
		t.Helper()
		res, err := d.Exec(repo.SQL(q), args...)
		if err != nil {
			t.Fatal(err)
		}
		id, err := res.LastInsertId()
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	u := execID("INSERT INTO {users} (appid,username,ip_list) VALUES (?, 'scope-probe','[]')", original)
	f := execID("INSERT INTO {users} (appid,username,ip_list) VALUES (?, 'scope-foreign','[]')", foreignID)
	k := execID("INSERT INTO {kamis} (appid,card,ip_list) VALUES (?, 'scope-card','[]')", original)
	defer d.Exec(repo.SQL("DELETE FROM {users} WHERE id IN (?,?)"), u, f)
	defer d.Exec(repo.SQL("DELETE FROM {kamis} WHERE id=?"), k)
	defer d.Exec(repo.SQL("DELETE FROM {user_logs} WHERE uid IN (?,?,?)"), u, f, -k)
	first := execID("INSERT INTO {user_logs} (uid,action,create_time) VALUES (?, 'scope-test', 100)", u)
	execID("INSERT INTO {user_logs} (uid,action,create_time) VALUES (?, 'scope-test', 200)", u)
	card := execID("INSERT INTO {user_logs} (uid,action,create_time) VALUES (?, 'scope-test', 200)", -k)
	foreign := execID("INSERT INTO {user_logs} (uid,action,create_time) VALUES (?, 'foreign-only', 200)", f)
	ok := func(method, path string, body any) map[string]any {
		t.Helper()
		v := request(method, "/api/admin"+path, body, token, false)
		if v["code"] != float64(0) {
			t.Fatalf("log API failed: %v", v)
		}
		return v
	}
	t.Run("AuditOutcomeAndAccountValidation", func(t *testing.T) {
		for _, tc := range []struct {
			path string
			body map[string]any
		}{
			{fmt.Sprintf("/user/%d", u), map[string]any{"bindLimit": -1}},
			{fmt.Sprintf("/user/%d", u), map[string]any{"ipList": "not-json"}},
			{fmt.Sprintf("/kami/%d", k), map[string]any{"regionLimit": -1}},
		} {
			if got := request("PUT", "/api/admin"+tc.path, tc.body, token, false); got["code"] != float64(400) {
				t.Fatalf("invalid account settings accepted: %v", got)
			}
		}
		var row struct{ Action, Data string }
		if err := d.Get(&row, repo.SQL("SELECT action,data FROM {logs} WHERE type='admin' ORDER BY id DESC LIMIT 1")); err != nil {
			t.Fatal(err)
		}
		var outcome struct {
			Code       int
			Success    bool
			HTTPStatus int
		}
		if json.Unmarshal([]byte(row.Data), &outcome) != nil || outcome.Code != 400 || outcome.Success || outcome.HTTPStatus != 200 || row.Action != "修改了卡密" {
			t.Fatalf("wrong failure audit: %+v %+v", row, outcome)
		}
		probe := execID("INSERT INTO {logs} (type,action,operator,username,create_time) VALUES ('admin','保存了公告','audit-operator-search','audit-operator-search',1)")
		defer d.Exec(repo.SQL("DELETE FROM {logs} WHERE id=?"), probe)
		if got := ok("GET", "/log/list?category=admin&keyword=audit-operator-search", nil)["data"].(map[string]any)["total"]; got != float64(1) {
			t.Fatal("operator search failed", got)
		}
	})
	for query, want := range map[string]float64{
		"category=user&action=scope-test&start=200&end=200": 2,
		"category=user&keyword=scope-probe":                 2,
		"category=user&keyword=scope-card":                  1,
		"category=user&keyword=scope-foreign":               0,
	} {
		if got := ok("GET", "/log/list?"+query, nil)["data"].(map[string]any)["total"]; got != want {
			t.Fatalf("%s total=%v want=%v", query, got, want)
		}
	}
	for _, kind := range []string{"system", "admin"} {
		id := execID("INSERT INTO {logs} (type,action,create_time) VALUES (?, 'scope-log-type', 100)", kind)
		defer d.Exec(repo.SQL("DELETE FROM {logs} WHERE id=?"), id)
	}
	for _, kind := range []string{"system", "admin"} {
		result := ok("GET", "/log/list?type="+kind+"&action=scope-log-type", nil)["data"].(map[string]any)
		if result["total"] != float64(1) || result["list"].([]any)[0].(map[string]any)["type"] != kind {
			t.Fatal("log type filter ignored", kind, result)
		}
	}
	for _, action := range ok("GET", "/log/actions?category=user", nil)["data"].([]any) {
		if action == "foreign-only" {
			t.Fatal("cross-app log action exposed")
		}
	}
	for _, query := range []string{"start=bad", "start=201&end=200", "end=-1"} {
		if v := request("GET", "/api/admin/log/list?category=user&"+query, nil, token, false); v["code"] == float64(0) {
			t.Fatal("invalid date filter accepted", query)
		}
	}
	*appID = 0
	if v := request("POST", "/api/admin/log/clear", map[string]any{"category": "user"}, token, false); v["code"] == float64(0) {
		t.Fatal("unscoped log deletion accepted")
	}
	*appID = original
	v := ok("POST", "/log/clear", map[string]any{"category": "user", "ids": []int64{first, card, foreign}})
	if v["data"].(map[string]any)["affected"] != float64(2) {
		t.Fatal("selected deletion scope", v)
	}
	ok("POST", "/log/clear", map[string]any{"category": "user", "before": 201})
	var count int
	if err := d.Get(&count, repo.SQL("SELECT COUNT(*) FROM {user_logs} WHERE id=?"), foreign); err != nil || count != 1 {
		t.Fatal("foreign log removed", count, err)
	}
	if err := d.Get(&count, repo.SQL("SELECT COUNT(*) FROM {user_logs} WHERE uid IN (?,?)"), u, -k); err != nil || count != 0 {
		t.Fatal("own log cleanup failed", count, err)
	}
	*appID = foreignID
	if got := ok("GET", "/log/list?category=user&keyword=scope-foreign", nil)["data"].(map[string]any)["total"]; got != float64(1) {
		t.Fatal("foreign app lost its log", got)
	}
	t.Run("DashboardTrendRanges", func(t *testing.T) {
		*appID = original
		now := time.Now()
		midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		// Records exactly on the first included day, and one second before it.
		for _, offset := range []int{-29, -6} {
			stamp := midnight.AddDate(0, 0, offset).Unix()
			for _, uid := range []int64{u, -k, f} {
				id := execID("INSERT INTO {user_logs} (uid,action,create_time) VALUES (?,'login',?)", uid, stamp)
				defer d.Exec(repo.SQL("DELETE FROM {user_logs} WHERE id=?"), id)
			}
			id := execID("INSERT INTO {user_logs} (uid,action,create_time) VALUES (?,'login',?)", u, stamp-1)
			defer d.Exec(repo.SQL("DELETE FROM {user_logs} WHERE id=?"), id)
			for _, app := range []int64{original, foreignID} {
				id := execID("INSERT INTO {users} (appid,username,create_time,ip_list) VALUES (?,?,?,'[]')", app, fmt.Sprintf("trend-%d-%d", app, offset), stamp)
				defer d.Exec(repo.SQL("DELETE FROM {users} WHERE id=?"), id)
			}
		}
		for query, days := range map[string]int{"": 7, "?range=7": 7, "?range=30": 30} {
			trend := ok("GET", "/dashboard"+query, nil)["data"].(map[string]any)["trend"].(map[string]any)
			dates, logins, regs := trend["dates"].([]any), trend["logins"].([]any), trend["regs"].([]any)
			if len(dates) != days || len(logins) != days || len(regs) != days {
				t.Fatalf("wrong range %s: %v", query, trend)
			}
			for i, date := range dates {
				if date != midnight.AddDate(0, 0, i-days+1).Format("01-02") {
					t.Fatal("date sequence", dates)
				}
			}
			if logins[0] != float64(2) || regs[0] != float64(1) {
				t.Fatal("date boundary or app scope", trend)
			}
			if days == 30 && (logins[23] != float64(2) || regs[23] != float64(1)) {
				t.Fatal("seven-day records absent in thirty-day series", trend)
			}
		}
		for _, value := range []string{"0", "-1", "8", "365", "bad", "30.5"} {
			if v := request("GET", "/api/admin/dashboard?range="+value, nil, token, false); v["code"] != float64(400) {
				t.Fatal("invalid range accepted", value, v)
			}
		}
	})
	t.Run("DashboardRecentMergesAdminAndScopedUsers", func(t *testing.T) {
		base := time.Now().Add(time.Hour).Unix()
		for i := 0; i < 10; i++ {
			operator := "dashboard-admin"
			if i == 0 {
				operator = ""
			}
			id := execID("INSERT INTO {logs} (type,action,operator,create_time) VALUES ('admin',?,?,?)", fmt.Sprintf("dashboard-action-%d", i), operator, base-int64(i))
			defer d.Exec(repo.SQL("DELETE FROM {logs} WHERE id=?"), id)
		}
		system := execID("INSERT INTO {logs} (type,action,create_time) VALUES ('system','dashboard-system-secret',?)", base+10)
		defer d.Exec(repo.SQL("DELETE FROM {logs} WHERE id=?"), system)
		userLog := execID("INSERT INTO {user_logs} (uid,action,create_time) VALUES (?,'dashboard-user',?)", u, base+3)
		cardLog := execID("INSERT INTO {user_logs} (uid,action,create_time) VALUES (?,'dashboard-card',?)", -k, base+2)
		foreignLog := execID("INSERT INTO {user_logs} (uid,action,create_time) VALUES (?,'dashboard-foreign',?)", f, base+4)
		defer d.Exec(repo.SQL("DELETE FROM {user_logs} WHERE id IN (?,?,?)"), userLog, cardLog, foreignLog)
		getRecent := func() []any { return ok("GET", "/dashboard", nil)["data"].(map[string]any)["recent"].([]any) }
		// No selected app has no user events, but admin activity must still appear.
		*appID = 0
		rows := getRecent()
		if len(rows) != 8 || rows[0].(map[string]any)["text"] != "管理员 · dashboard-action-0" {
			t.Fatalf("admin-only recent missing or not newest first: %v", rows)
		}
		for i, row := range rows {
			v := row.(map[string]any)
			if v["type"] != "success" || !strings.HasSuffix(v["text"].(string), fmt.Sprintf("dashboard-action-%d", i)) {
				t.Fatalf("wrong admin ordering or timeline type: %v", rows)
			}
		}
		*appID = original
		rows = getRecent()
		if len(rows) != 8 || rows[0].(map[string]any)["text"] != "scope-probe · dashboard-user" || rows[1].(map[string]any)["text"] != "scope-card · dashboard-card" || rows[2].(map[string]any)["text"] != "管理员 · dashboard-action-0" {
			t.Fatalf("merged recent order/limit: %v", rows)
		}
		if rows[0].(map[string]any)["type"] != "primary" || rows[1].(map[string]any)["type"] != "primary" {
			t.Fatal("invalid user timeline type", rows)
		}
		for _, row := range rows {
			text := row.(map[string]any)["text"].(string)
			if strings.Contains(text, "dashboard-foreign") || strings.Contains(text, "dashboard-system") {
				t.Fatal("unrelated log exposed", rows)
			}
		}
	})
}
