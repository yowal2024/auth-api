// Reconstructed declarations from Go runtime metadata; method bodies are not recovered.
package buildpool

import (
	"auth-api/internal/config"
	"sync"
)

// Recovered from runtime type 0x12e8300; original size 48 bytes.
type DiffStore struct {
	cfg config.CloneConfig
	mu  sync.Mutex
}

// Recovered from runtime type 0x1328fe0; original size 64 bytes.
type Meta struct {
	NewPackage     string `json:"newPackage"`
	OriginalSha256 string `json:"originalSha256"`
	CloneSha256    string `json:"cloneSha256"`
	DiffName       string `json:"diffName"`
}
