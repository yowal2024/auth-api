// Reconstructed declarations from Go runtime metadata; method bodies are not recovered.
package response

// Recovered from runtime type 0x1311ac0; original size 40 bytes.
type Envelope struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}
