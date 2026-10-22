package sdkcrypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"testing"
)

func unhex(s string) []byte {
	b, e := hex.DecodeString(s)
	if e != nil {
		panic(e)
	}
	return b
}
func TestKnownAESCBCVectorAndPadding(t *testing.T) {
	// NIST SP 800-38A F.2.1 first plaintext/ciphertext block.
	key := unhex("2b7e151628aed2a6abf7158809cf4f3c")
	iv := unhex("000102030405060708090a0b0c0d0e0f")
	plain := unhex("6bc1bee22e409f96e93d7e117393172a")
	first := unhex("7649abac8119b246cee98e9b12e9197d")
	b, _ := aes.NewCipher(key)
	pad := bytes.Repeat([]byte{16}, 16)
	last := make([]byte, 16)
	cipher.NewCBCEncrypter(b, first).CryptBlocks(last, pad)
	input := append(append(append([]byte{}, iv...), first...), last...)
	got, err := DecryptAESCBC(key, input)
	if err != nil || !bytes.Equal(got, plain) {
		t.Fatalf("vector failed: %x %v", got, err)
	}
	// Flip the previous block to make the final padding byte inconsistent.
	input[31] ^= 1
	if _, err := DecryptAESCBC(key, input); err == nil {
		t.Fatal("accepted invalid padding")
	}
	for _, bad := range [][]byte{nil, make([]byte, 16), make([]byte, 33)} {
		if _, err := DecryptAESCBC(key, bad); err == nil {
			t.Fatal("accepted invalid length")
		}
	}
}
func TestSignVector(t *testing.T) {
	// Independently calculated using Python hmac/sha256 over timestamp + LF + nonce + LF + body.
	want := "0a910c6435cd1ecfd57ccd5c06a3afb66723f42eae7f9dcedd409e5b16acd72a"
	got := SignBytes("test-key", "1700000000", "abc123", []byte(`{"user":"alice"}`))
	if got != want {
		t.Fatalf("signature: %s", got)
	}
	if !VerifySign(got, want) || VerifySign(got, want+"0") {
		t.Fatal("signature comparison")
	}
}
func TestCodecsAndKeyFormats(t *testing.T) {
	data := []byte{0xfb, 0xff, 0xfa, 0x01}
	for _, enc := range []*base64.Encoding{base64.StdEncoding, base64.RawStdEncoding, base64.URLEncoding, base64.RawURLEncoding} {
		got, err := DecodeBlob(enc.EncodeToString(data), "base64")
		if err != nil || !bytes.Equal(got, data) {
			t.Fatal(err)
		}
	}
	key := []byte("sixteen-byte-key")
	for _, s := range []string{string(key), hex.EncodeToString(key), base64.StdEncoding.EncodeToString(key)} {
		got, ok := decodeKeyN(s, 16, 24, 32)
		if !ok || !bytes.Equal(got, key) {
			t.Fatalf("key format: %q", s)
		}
	}
	if _, ok := decodeKeyN("bad", 16); ok {
		t.Fatal("accepted bad key")
	}
}
func TestCBCAndRSABoundaries(t *testing.T) {
	key := []byte("sixteen-byte-key")
	for _, n := range []int{0, 1, 15, 16, 17, 31, 32, 1000} {
		p := bytes.Repeat([]byte{'x'}, n)
		enc, err := EncryptAESCBC(key, p)
		if err != nil {
			t.Fatal(err)
		}
		dec, err := DecryptAESCBC(key, enc)
		if err != nil || !bytes.Equal(p, dec) {
			t.Fatalf("AES n=%d: %v", n, err)
		}
		enc, err = EncryptDESCBC(key[:8], p)
		if err != nil {
			t.Fatal(err)
		}
		dec, err = DecryptDESCBC(key[:8], enc)
		if err != nil || !bytes.Equal(p, dec) {
			t.Fatalf("DES n=%d: %v", n, err)
		}
	}
	rk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	priv := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(rk)}))
	pub := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(&rk.PublicKey)}))
	p := bytes.Repeat([]byte("boundary"), 100)
	enc, err := RSAEncrypt(pub, p)
	if err != nil {
		t.Fatal(err)
	}
	dec, err := RSADecrypt(priv, enc)
	if err != nil || !bytes.Equal(p, dec) {
		t.Fatal("RSA multi-block", err)
	}
	if _, err := RSADecrypt(priv, enc[:len(enc)-1]); err == nil {
		t.Fatal("accepted truncated RSA")
	}
	if _, err := RSADecrypt(priv, nil); err == nil {
		t.Fatal("accepted empty RSA")
	}
}
