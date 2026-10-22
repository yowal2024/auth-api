package server

import (
	"auth-api/internal/repo"
	"bytes"
	"context"
	"encoding/json"
	"github.com/jmoiron/sqlx"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Rewrite before Gin selects a route, including default routes disabled by overrides.
type RewriteHandler struct {
	Next    http.Handler
	DB      *sqlx.DB
	mu      sync.Mutex
	expires time.Time
	routes  []repo.APIRoute
	codes   []repo.APICode
	entry   string
	Err     error
}

func (h *RewriteHandler) Refresh(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	var entry string
	routes := []repo.APIRoute{}
	codes := []repo.APICode{}
	if e := h.DB.GetContext(ctx, &entry, repo.SQL("SELECT admin_entry FROM {settings} WHERE id=1")); e != nil {
		return e
	}
	if e := h.DB.SelectContext(ctx, &routes, repo.SQL("SELECT id,api_key,name,override_path,rewritten,sort_order FROM {api_routes}")); e != nil {
		return e
	}
	if e := h.DB.SelectContext(ctx, &codes, repo.SQL("SELECT id,code,override_code,message,remark,rewritten FROM {api_codes}")); e != nil {
		return e
	}
	h.entry = entry
	h.routes = routes
	h.codes = codes
	h.expires = time.Now().Add(time.Second)
	return nil
}
func (h *RewriteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	stale := time.Now().After(h.expires)
	h.mu.Unlock()
	if stale {
		if e := h.Refresh(r.Context()); e != nil {
			http.Error(w, "configuration unavailable", 503)
			return
		}
	}
	h.mu.Lock()
	routes := append([]repo.APIRoute{}, h.routes...)
	codes := append([]repo.APICode{}, h.codes...)
	entry := h.entry
	h.mu.Unlock()
	p := r.URL.Path
	if strings.HasPrefix(p, "/api/") && entry != "" && entry != "admin" {
		if p == "/api/admin" || strings.HasPrefix(p, "/api/admin/") {
			http.NotFound(w, r)
			return
		}
		prefix := "/api/" + entry
		if p == prefix || strings.HasPrefix(p, prefix+"/") {
			r = r.Clone(r.Context())
			u := *r.URL
			r.URL = &u
			r.URL.Path = "/api/admin" + strings.TrimPrefix(p, prefix)
		}
	}
	if strings.HasPrefix(p, "/api/sdk/") {
		for _, v := range routes {
			if v.Override != "" && !strings.HasPrefix(v.Override, "/") {
				v.Override = "/api/sdk/" + v.Override
			}
			if v.Rewritten != 1 || v.Override == "" {
				continue
			}
			original := "/api/sdk/" + v.APIKey
			if p == original && p != v.Override {
				http.NotFound(w, r)
				return
			}
			if p == v.Override {
				r = r.Clone(r.Context())
				u := *r.URL
				r.URL = &u
				r.URL.Path = original
				break
			}
		}
		bw := &codeWriter{ResponseWriter: w}
		h.Next.ServeHTTP(bw, r)
		if bw.streaming {
			return
		}
		if bw.Header().Get("Content-Disposition") == "" && strings.Contains(bw.Header().Get("Content-Type"), "application/json") {
			var body map[string]json.RawMessage
			if json.Unmarshal(bw.body.Bytes(), &body) == nil {
				var code int
				if json.Unmarshal(body["code"], &code) == nil {
					for _, v := range codes {
						if v.Code == code && v.Rewritten == 1 {
							body["code"], _ = json.Marshal(v.OverrideCode)
							break
						}
					}
					b, e := json.Marshal(body)
					if e == nil {
						bw.body.Reset()
						bw.body.Write(b)
					}
				}
			}
		}
		bw.Header().Del("Content-Length")
		if bw.status == 0 {
			bw.status = 200
		}
		w.WriteHeader(bw.status)
		w.Write(bw.body.Bytes())
		return
	}
	h.Next.ServeHTTP(w, r)
}

type codeWriter struct {
	http.ResponseWriter
	body      bytes.Buffer
	status    int
	streaming bool
}

func (w *codeWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	if !w.streaming && w.body.Len()+len(b) > 4<<20 {
		w.flushBody()
	}
	if w.streaming {
		return w.ResponseWriter.Write(b)
	}
	return w.body.Write(b)
}
func (w *codeWriter) WriteHeader(n int) {
	if w.status != 0 {
		return
	}
	w.status = n
	if w.Header().Get("Content-Disposition") != "" || !strings.Contains(w.Header().Get("Content-Type"), "application/json") {
		w.streaming = true
		w.ResponseWriter.WriteHeader(n)
	}
}
func (w *codeWriter) flushBody() {
	if w.streaming {
		return
	}
	if w.status == 0 {
		w.status = 200
	}
	w.streaming = true
	w.ResponseWriter.WriteHeader(w.status)
	_, _ = w.ResponseWriter.Write(w.body.Bytes())
	w.body.Reset()
}
func (w *codeWriter) Flush() {
	w.flushBody()
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
