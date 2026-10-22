// Reconstructed declarations from Go runtime metadata; method bodies are not recovered.
package middleware

import (
	"bytes"
	"github.com/gin-gonic/gin"
)

// Recovered from runtime type 0x135bba0; original size 24 bytes.
type cryptoWriter struct {
	gin.ResponseWriter
	buf *bytes.Buffer
}
