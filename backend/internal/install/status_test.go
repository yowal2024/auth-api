package install

import (
	"auth-api/internal/config"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestStatusDoesNotDiscloseHiddenEntry(t *testing.T) {
	h := NewHandler("unused", &config.Config{Installed: true}, nil)
	h.AdminEntryFunc = func() string { return "private-secret" }
	g := gin.New()
	g.GET("/status", h.Status)
	for _, candidate := range []string{"", "admin", "private-secret"} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/status", nil)
		req.Header.Set("X-Admin-Entry", candidate)
		g.ServeHTTP(w, req)
		var result struct{ Data map[string]any }
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if result.Data["installed"] != true {
			t.Fatal(result)
		}
		if candidate == "private-secret" {
			if result.Data["adminEntry"] != candidate {
				t.Fatal("known entry not validated")
			}
		} else if strings.Contains(w.Body.String(), "private-secret") || result.Data["adminEntry"] != nil {
			t.Fatal("entry disclosed", w.Body.String())
		}
		if w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("entry validation must not be cached")
		}
	}
}
