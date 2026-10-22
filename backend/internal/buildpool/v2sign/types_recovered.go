// Reconstructed declarations from Go runtime metadata; method bodies are not recovered.
package v2sign

import (
	"crypto/rsa"
)

// Recovered from runtime type 0x1311400; original size 56 bytes.
type KeyAndCert struct {
	PrivateKey   *rsa.PrivateKey
	CertDER      []uint8
	PublicKeyDER []uint8
}
