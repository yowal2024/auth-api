// Reconstructed JavaScript API. Goja replaces the original cgo QuickJS engine.
package sandbox

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dop251/goja"
	"io"
	"net/http"
	"strings"
	"time"
)

func Run(ctx context.Context, o RunOptions) (result Result) {
	start := time.Now()
	result.Logs = []string{}
	defer func() {
		result.Duration = time.Since(start).Milliseconds()
		if p := recover(); p != nil {
			result.OK = false
			result.Error = fmt.Sprint(p)
		}
	}()
	if o.Timeout <= 0 {
		o.Timeout = 3 * time.Second
	}
	if o.Timeout > 30*time.Second {
		o.Timeout = 30 * time.Second
	}
	if len(o.Code) > 1<<20 {
		result.Error = "代码过大"
		return
	}
	ctx, cancel := context.WithTimeout(ctx, o.Timeout)
	defer cancel()
	vm := goja.New()
	vm.SetMaxCallStackSize(1024)
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			vm.Interrupt("执行超时")
		case <-done:
		}
	}()
	console := map[string]any{}
	for _, level := range []string{"log", "info", "warn", "error"} {
		level := level
		console[level] = func(call goja.FunctionCall) goja.Value {
			if len(result.Logs) < 200 {
				parts := []string{}
				for _, v := range call.Arguments {
					parts = append(parts, v.String())
				}
				line := "[" + level + "] " + strings.Join(parts, " ")
				if len(line) > 4096 {
					line = line[:4096]
				}
				result.Logs = append(result.Logs, line)
			}
			return goja.Undefined()
		}
	}
	vm.Set("console", console)
	vm.Set("crypto", map[string]any{"md5": func(s string) string { x := md5.Sum([]byte(s)); return hex.EncodeToString(x[:]) }, "sha1": func(s string) string { x := sha1.Sum([]byte(s)); return hex.EncodeToString(x[:]) }, "sha256": func(s string) string { x := sha256.Sum256([]byte(s)); return hex.EncodeToString(x[:]) }, "base64Encode": func(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }, "base64Decode": func(s string) (string, error) { b, e := base64.StdEncoding.DecodeString(s); return string(b), e }})
	request := func(method, url string, body any) map[string]any {
		var reader io.Reader
		if body != nil {
			b, e := json.Marshal(body)
			if e != nil {
				return map[string]any{"ok": false, "error": e.Error()}
			}
			reader = strings.NewReader(string(b))
		}
		req, e := http.NewRequestWithContext(ctx, method, url, reader)
		if e != nil {
			return map[string]any{"ok": false, "error": e.Error()}
		}
		if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
			return map[string]any{"ok": false, "error": "unsupported URL scheme"}
		}
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		client := &http.Client{Timeout: o.Timeout, Transport: scriptTransport}
		resp, e := client.Do(req)
		if e != nil {
			return map[string]any{"ok": false, "error": e.Error()}
		}
		defer resp.Body.Close()
		b, e := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
		if e != nil {
			return map[string]any{"ok": false, "error": e.Error()}
		}
		return map[string]any{"ok": resp.StatusCode < 400, "status": resp.StatusCode, "body": string(b)}
	}
	vm.Set("http", map[string]any{"get": func(url string) map[string]any { return request("GET", url, nil) }, "post": func(url string, body any) map[string]any { return request("POST", url, body) }})
	vm.Set("db", map[string]any{"query": func(query string, args ...any) ([]map[string]any, error) {
		if o.DB == nil {
			return nil, fmt.Errorf("database unavailable")
		}
		trim := strings.TrimSpace(query)
		if !strings.HasPrefix(strings.ToUpper(trim), "SELECT ") || strings.Contains(trim, ";") {
			return nil, fmt.Errorf("only one SELECT query is allowed")
		}
		tx, e := o.DB.BeginTxx(ctx, &sql.TxOptions{ReadOnly: true})
		if e != nil {
			return nil, e
		}
		defer tx.Rollback()
		rows, e := tx.QueryxContext(ctx, trim, args...)
		if e != nil {
			return nil, e
		}
		defer rows.Close()
		out := []map[string]any{}
		for rows.Next() {
			if len(out) >= 1000 {
				return nil, fmt.Errorf("query result exceeds 1000 rows")
			}
			v := map[string]any{}
			if e = rows.MapScan(v); e != nil {
				return nil, e
			}
			for k, x := range v {
				if b, ok := x.([]byte); ok {
					v[k] = string(b)
				}
			}
			out = append(out, v)
		}
		return out, rows.Err()
	}})
	app := map[string]any{}
	if o.App != nil {
		app = map[string]any{"id": o.App.ID, "name": o.App.Name, "pkg": o.App.Pkg, "mode": o.App.Mode, "type": o.App.Type}
	}
	var user any
	if o.User != nil {
		b, e := json.Marshal(o.User)
		if e != nil {
			result.Error = e.Error()
			return
		}
		if e = json.Unmarshal(b, &user); e != nil {
			result.Error = e.Error()
			return
		}
	}
	env := map[string]any{"app": app, "user": user, "operator": o.Operator, "payload": o.Payload, "trigger": o.Trigger, "ip": o.IP}
	for k, v := range env {
		vm.Set(k, v)
	}
	_, e := vm.RunString(o.Code)
	if e != nil {
		result.Error = e.Error()
		return
	}
	fn, ok := goja.AssertFunction(vm.Get("handler"))
	if !ok {
		result.Error = "缺少 handler 函数"
		return
	}
	v, e := fn(goja.Undefined(), vm.ToValue(env))
	if e != nil {
		result.Error = e.Error()
		var exception *goja.Exception
		if errors.As(e, &exception) {
			value := exception.Value()
			message := value.String()
			if object, ok := value.(*goja.Object); ok {
				if msg := object.Get("message"); msg != nil && !goja.IsUndefined(msg) {
					message = msg.String()
				}
			}
			if strings.HasPrefix(message, "业务错误:") {
				result.Error = message
			}
		}
		return
	}
	result.Result = v.Export()
	if promise, ok := result.Result.(*goja.Promise); ok {
		if promise.State() == goja.PromiseStateFulfilled {
			result.Result = promise.Result().Export()
		} else if promise.State() == goja.PromiseStateRejected {
			result.Error = promise.Result().String()
			return
		} else {
			result.Error = "异步任务未完成"
			return
		}
	}
	result.OK = true
	return
}
