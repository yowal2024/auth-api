// Reconstructed declarations from Go runtime metadata; method bodies are not recovered.
package install

import (
	"auth-api/internal/config"
	"io/fs"
	"sync"
)

// Recovered from runtime type 0x134a580; original size 64 bytes.
type Handler struct {
	configPath     string
	cfg            *config.Config
	migrations     fs.FS
	mu             sync.Mutex
	OnFinalize     func() error
	AdminEntryFunc func() string
}

// Recovered from runtime type 0x130b400; original size 48 bytes.
type createAdminReq struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	AdminEntry string `json:"adminEntry"`
}

// Recovered from runtime type 0x130b580; original size 136 bytes.
type saveConfigReq struct {
	MySQL  testMySQLReq `json:"mysql"`
	Redis  testRedisReq `json:"redis"`
	Server *struct {
		Addr string "json:\"addr\""
		Mode string "json:\"mode\""
	} `json:"server,omitempty"`
}

// Recovered from runtime type 0x134a480; original size 88 bytes.
type testMySQLReq struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	User        string `json:"user"`
	Password    string `json:"password"`
	Database    string `json:"database"`
	TablePrefix string `json:"tablePrefix"`
}

// Recovered from runtime type 0x130b4c0; original size 40 bytes.
type testRedisReq struct {
	Addr     string `json:"addr"`
	DB       int    `json:"db"`
	Password string `json:"password"`
}
