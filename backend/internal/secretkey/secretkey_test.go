package secretkey

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func TestGCMFormatAndTamper(t *testing.T) {
	key := bytes.Repeat([]byte{7}, 32)
	nonce := bytes.Repeat([]byte{2}, 12)
	plain := []byte("saved secret")
	b, _ := aes.NewCipher(key)
	g, _ := cipher.NewGCM(b)
	wire := g.Seal(append([]byte{}, nonce...), nonce, plain, nil)
	got, err := Decrypt(key, base64.StdEncoding.EncodeToString(wire))
	if err != nil || !bytes.Equal(got, plain) {
		t.Fatal(err)
	}
	wire[len(wire)-1] ^= 1
	if _, err := Decrypt(key, base64.StdEncoding.EncodeToString(wire)); err == nil {
		t.Fatal("tampering accepted")
	}
	encoded, err := Encrypt(key, plain)
	if err != nil {
		t.Fatal(err)
	}
	wire, err = base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatal(err)
	}
	got, err = g.Open(nil, wire[:12], wire[12:], nil)
	if err != nil || !bytes.Equal(got, plain) {
		t.Fatal(err)
	}
}
func TestLoadPreservesExistingKey(t *testing.T) {
	p := filepath.Join(t.TempDir(), "data", "master.key")
	first, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Load(p)
	if err != nil || !bytes.Equal(first, second) {
		t.Fatal("key was replaced", err)
	}
	if err := os.WriteFile(p, []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(p); err == nil {
		t.Fatal("accepted corrupt key")
	}
	after, _ := os.ReadFile(p)
	if string(after) != "broken" {
		t.Fatal("overwrote corrupt key")
	}
}
