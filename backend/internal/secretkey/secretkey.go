// Reconstructed AES-GCM format: standard base64(nonce || ciphertext || tag).
package secretkey

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
)

func Load(path string) ([]byte, error) {
	key, err := os.ReadFile(path)
	if err == nil {
		if len(key) != 32 {
			return nil, fmt.Errorf("master key %s 长度异常: 期望 %d 实际 %d", path, 32, len(key))
		}
		return key, nil
	}
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read master key: %w", err)
	}
	key = make([]byte, 32)
	if _, err = rand.Read(key); err != nil {
		return nil, fmt.Errorf("gen master key: %w", err)
	}
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, fmt.Errorf("mkdir for master key: %w", err)
	}
	// Publish a complete key without replacing one created by another request/process.
	f, err := os.CreateTemp(filepath.Dir(path), ".master-key-*")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(key); err != nil {
		f.Close()
		return nil, err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return nil, err
	}
	if err = f.Close(); err != nil {
		return nil, err
	}
	if err = os.Link(f.Name(), path); err != nil && !os.IsExist(err) {
		return nil, fmt.Errorf("publish master key: %w", err)
	}
	// Every caller reads the winner, including when another process won the race.
	key, err = os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("invalid master key length: %d", len(key))
	}
	return key, nil
}
func newGCM(key []byte) (cipher.AEAD, error) {
	b, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}
	return cipher.NewGCM(b)
}
func Encrypt(key, plain []byte) (string, error) {
	g, err := newGCM(key)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, g.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return "", fmt.Errorf("gcm nonce: %w", err)
	}
	return base64.StdEncoding.EncodeToString(g.Seal(nonce, nonce, plain, nil)), nil
}
func Decrypt(key []byte, encoded string) ([]byte, error) {
	g, err := newGCM(key)
	if err != nil {
		return nil, err
	}
	b, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("b64 decode: %w", err)
	}
	if len(b) < g.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}
	plain, err := g.Open(nil, b[:g.NonceSize()], b[g.NonceSize():], nil)
	if err != nil {
		return nil, fmt.Errorf("gcm open: %w", err)
	}
	return plain, nil
}
