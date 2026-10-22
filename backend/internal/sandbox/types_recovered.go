// Reconstructed declarations from Go runtime metadata; method bodies are not recovered.
package sandbox

import (
	"auth-api/internal/repo"
	"github.com/jmoiron/sqlx"
	"time"
)

// Recovered from runtime type 0x133f160; original size 72 bytes.
type Result struct {
	OK       bool        `json:"ok"`
	Result   interface{} `json:"result,omitempty"`
	Logs     []string    `json:"logs,omitempty"`
	Error    string      `json:"error,omitempty"`
	Duration int64       `json:"duration"`
}

// Recovered from runtime type 0x1367820; original size 120 bytes.
type RunOptions struct {
	Code     string
	Timeout  time.Duration
	Payload  interface{}
	App      *repo.App
	Operator string
	DB       *sqlx.DB
	User     interface{}
	Trigger  string
	IP       string
}
