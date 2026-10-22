package v2sign

import (
	"auth-api/internal/buildpool/edge"
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/binary"
	"fmt"
	"math/big"
	"time"
)

func GenerateKeyAndCert(commonNames ...string) (*KeyAndCert, error) {
	key, e := rsa.GenerateKey(rand.Reader, 2048)
	if e != nil {
		return nil, e
	}
	serial, e := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if e != nil {
		return nil, e
	}
	name := "Auth API"
	if len(commonNames) > 0 && commonNames[0] != "" {
		name = commonNames[0]
	}
	cert := &x509.Certificate{SerialNumber: serial, Subject: pkix.Name{CommonName: name}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().AddDate(30, 0, 0), KeyUsage: x509.KeyUsageDigitalSignature}
	der, e := x509.CreateCertificate(rand.Reader, cert, cert, &key.PublicKey, key)
	if e != nil {
		return nil, e
	}
	pub, e := x509.MarshalPKIXPublicKey(&key.PublicKey)
	return &KeyAndCert{key, der, pub}, e
}
func MarshalKeyAndCert(k *KeyAndCert) ([]byte, []byte, error) {
	if k == nil {
		return nil, nil, fmt.Errorf("nil key")
	}
	priv, e := x509.MarshalPKCS8PrivateKey(k.PrivateKey)
	return priv, k.CertDER, e
}
func LoadKeyAndCert(priv, cert []byte) (*KeyAndCert, error) {
	p, e := x509.ParsePKCS8PrivateKey(priv)
	if e != nil {
		return nil, e
	}
	key, ok := p.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not RSA")
	}
	c, e := x509.ParseCertificate(cert)
	if e != nil {
		return nil, e
	}
	pub, e := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if e != nil {
		return nil, e
	}
	cp, e := x509.MarshalPKIXPublicKey(c.PublicKey)
	if e != nil || !bytes.Equal(pub, cp) {
		return nil, fmt.Errorf("certificate and key mismatch")
	}
	return &KeyAndCert{key, cert, pub}, nil
}
func u32(v uint32) []byte { b := make([]byte, 4); binary.LittleEndian.PutUint32(b, v); return b }
func lp(b []byte) []byte  { return append(u32(uint32(len(b))), b...) }
func ComputeAPKDigest(apk []byte) ([]byte, error) {
	start, cd, eocd, e := edge.SigningBlock(apk)
	if e != nil {
		return nil, e
	}
	end := append([]byte{}, apk[eocd:]...)
	binary.LittleEndian.PutUint32(end[16:], uint32(start))
	chunks := []byte{}
	count := uint32(0)
	for _, section := range [][]byte{apk[:start], apk[cd:eocd], end} {
		for len(section) > 0 {
			n := len(section)
			if n > 1<<20 {
				n = 1 << 20
			}
			h := sha256.New()
			h.Write([]byte{0xa5})
			h.Write(u32(uint32(n)))
			h.Write(section[:n])
			chunks = append(chunks, h.Sum(nil)...)
			count++
			section = section[n:]
		}
	}
	h := sha256.New()
	h.Write([]byte{0x5a})
	h.Write(u32(count))
	h.Write(chunks)
	return h.Sum(nil), nil
}
func SignV2(apk []byte, key *KeyAndCert) ([]byte, error) {
	pairs, e := edge.Pairs(apk)
	if e != nil {
		return nil, e
	}
	clean, e := edge.StripSigningBlock(apk)
	if e != nil {
		return nil, e
	}
	digest, e := ComputeAPKDigest(clean)
	if e != nil {
		return nil, e
	}
	record := append(u32(0x0103), lp(digest)...)
	signed := append(lp(lp(record)), lp(lp(key.CertDER))...)
	signed = append(signed, lp(nil)...)
	hash := sha256.Sum256(signed)
	signature, e := rsa.SignPKCS1v15(rand.Reader, key.PrivateKey, crypto.SHA256, hash[:])
	if e != nil {
		return nil, e
	}
	sigRecord := append(u32(0x0103), lp(signature)...)
	signer := append(lp(signed), lp(lp(sigRecord))...)
	signer = append(signer, lp(key.PublicKeyDER)...)
	value := lp(lp(signer))
	var body bytes.Buffer
	pair := func(id uint32, v []byte) {
		binary.Write(&body, binary.LittleEndian, uint64(len(v)+4))
		body.Write(u32(id))
		body.Write(v)
	}
	pair(0x7109871a, value)
	// Preserve non-signature metadata pairs; drop invalidated v3/v3.1/source-stamp signatures.
	for id, v := range pairs {
		if id != 0x7109871a && id != 0xf05368c0 && id != 0x1b93ad61 && id != 0x6dff800d {
			pair(id, v)
		}
	}
	size := uint64(body.Len() + 24)
	var block bytes.Buffer
	binary.Write(&block, binary.LittleEndian, size)
	block.Write(body.Bytes())
	binary.Write(&block, binary.LittleEndian, size)
	block.Write(edge.Magic)
	return edge.AddBlock(clean, block.Bytes())
}
