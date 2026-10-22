// Reconstructed declarations from Go runtime metadata; method bodies are not recovered.
package tlsmgr

import (
	"auth-api/internal/repo"
	"crypto/x509"
	"sync/atomic"
)

// Recovered from runtime type 0x12e8ee0; original size 16 bytes.
type Manager struct {
	apps   *repo.AppRepo
	caPool *atomicPool
}

// Recovered from runtime type 0x12b9ee0; original size 8 bytes.
type atomicPool struct {
	p atomic.Pointer[x509.CertPool]
}
