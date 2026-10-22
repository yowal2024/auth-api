// Reconstructed from assembly. Compatibility with production clients is not yet verified.
package sdkcrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rc4"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
)

var ErrCiphertext = errors.New("invalid ciphertext")
var ErrPadding = errors.New("invalid PKCS7 padding")

func SignBytes(key, first, second string, body []byte) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(first))
	h.Write([]byte{'\n'})
	h.Write([]byte(second))
	h.Write([]byte{'\n'})
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}
func VerifySign(expected, actual string) bool { return hmac.Equal([]byte(expected), []byte(actual)) }

func DecodeBlob(s, encoding string) ([]byte, error) {
	if encoding == "hex" {
		return hex.DecodeString(s)
	}
	var err error
	for _, enc := range []*base64.Encoding{base64.StdEncoding, base64.RawStdEncoding, base64.URLEncoding, base64.RawURLEncoding} {
		var b []byte
		b, err = enc.DecodeString(s)
		if err == nil {
			return b, nil
		}
	}
	return nil, err
}
func EncodeBlob(b []byte, encoding string) string {
	if encoding == "hex" {
		return hex.EncodeToString(b)
	}
	return base64.StdEncoding.EncodeToString(b)
}
func decodeKeyN(s string, lengths ...int) ([]byte, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, false
	}
	valid := func(b []byte) bool {
		for _, n := range lengths {
			if len(b) == n {
				return true
			}
		}
		return false
	}
	if b, err := hex.DecodeString(s); err == nil && valid(b) {
		return b, true
	}
	for _, enc := range []*base64.Encoding{base64.StdEncoding, base64.RawStdEncoding, base64.URLEncoding, base64.RawURLEncoding} {
		if b, err := enc.DecodeString(s); err == nil && valid(b) {
			return b, true
		}
	}
	if b := []byte(s); valid(b) {
		return b, true
	}
	return nil, false
}
func Supported(algorithm string) bool {
	switch strings.ToLower(strings.TrimSpace(algorithm)) {
	case "aes", "des", "rc4", "rsa", "session":
		return true
	}
	return false
}
func EncryptAESCBC(key, plain []byte) ([]byte, error) {
	b, e := aes.NewCipher(key)
	if e != nil {
		return nil, e
	}
	return encryptCBC(b, plain)
}
func DecryptAESCBC(key, data []byte) ([]byte, error) {
	b, e := aes.NewCipher(key)
	if e != nil {
		return nil, e
	}
	return decryptCBC(b, data)
}
func EncryptDESCBC(key, plain []byte) ([]byte, error) {
	b, e := des.NewCipher(key)
	if e != nil {
		return nil, e
	}
	return encryptCBC(b, plain)
}
func DecryptDESCBC(key, data []byte) ([]byte, error) {
	b, e := des.NewCipher(key)
	if e != nil {
		return nil, e
	}
	return decryptCBC(b, data)
}
func encryptCBC(block cipher.Block, plain []byte) ([]byte, error) {
	bs := block.BlockSize()
	pad := bs - len(plain)%bs
	padded := make([]byte, len(plain)+pad)
	copy(padded, plain)
	for i := len(plain); i < len(padded); i++ {
		padded[i] = byte(pad)
	}
	result := make([]byte, bs+len(padded))
	if _, err := rand.Read(result[:bs]); err != nil {
		return nil, err
	}
	cipher.NewCBCEncrypter(block, result[:bs]).CryptBlocks(result[bs:], padded)
	return result, nil
}
func decryptCBC(block cipher.Block, data []byte) ([]byte, error) {
	bs := block.BlockSize()
	if len(data) < 2*bs || len(data)%bs != 0 {
		return nil, ErrCiphertext
	}
	plain := make([]byte, len(data)-bs)
	cipher.NewCBCDecrypter(block, data[:bs]).CryptBlocks(plain, data[bs:])
	n := int(plain[len(plain)-1])
	if n < 1 || n > bs || n > len(plain) {
		return nil, ErrPadding
	}
	for _, v := range plain[len(plain)-n:] {
		if int(v) != n {
			return nil, ErrPadding
		}
	}
	return plain[:len(plain)-n], nil
}
func RC4Crypt(key, data []byte) ([]byte, error) {
	c, err := rc4.NewCipher(key)
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(data))
	c.XORKeyStream(out, data)
	return out, nil
}
