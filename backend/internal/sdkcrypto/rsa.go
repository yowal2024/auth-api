// Reconstructed RSA PKCS#1 v1.5 block processing; not yet tested against the original executable.
package sdkcrypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

func parseRSAPub(s string) (*rsa.PublicKey, error) {
	b, _ := pem.Decode([]byte(s))
	if b == nil {
		return nil, errors.New("invalid public PEM")
	}
	if value, err := x509.ParsePKIXPublicKey(b.Bytes); err == nil {
		if key, ok := value.(*rsa.PublicKey); ok {
			return key, nil
		}
	}
	return x509.ParsePKCS1PublicKey(b.Bytes)
}
func parseRSAPriv(s string) (*rsa.PrivateKey, error) {
	b, _ := pem.Decode([]byte(s))
	if b == nil {
		return nil, errors.New("invalid private PEM")
	}
	if key, err := x509.ParsePKCS1PrivateKey(b.Bytes); err == nil {
		return key, nil
	}
	value, err := x509.ParsePKCS8PrivateKey(b.Bytes)
	if err != nil {
		return nil, err
	}
	key, ok := value.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an RSA private key")
	}
	return key, nil
}
func RSAEncrypt(publicPEM string, plain []byte) ([]byte, error) {
	key, err := parseRSAPub(publicPEM)
	if err != nil {
		return nil, err
	}
	step := key.Size() - 11
	if step <= 0 {
		return nil, errors.New("RSA key too small")
	}
	out := make([]byte, 0)
	for len(plain) > 0 {
		n := min(step, len(plain))
		block, err := rsa.EncryptPKCS1v15(rand.Reader, key, plain[:n])
		if err != nil {
			return nil, err
		}
		out = append(out, block...)
		plain = plain[n:]
	}
	return out, nil
}
func RSADecrypt(privatePEM string, data []byte) ([]byte, error) {
	key, err := parseRSAPriv(privatePEM)
	if err != nil {
		return nil, err
	}
	step := key.Size()
	if len(data) == 0 || len(data)%step != 0 {
		return nil, ErrCiphertext
	}
	out := make([]byte, 0)
	for len(data) > 0 {
		block, err := rsa.DecryptPKCS1v15(rand.Reader, key, data[:step])
		if err != nil {
			return nil, err
		}
		out = append(out, block...)
		data = data[step:]
	}
	return out, nil
}
