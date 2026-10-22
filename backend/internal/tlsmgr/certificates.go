package tlsmgr

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"strings"
	"time"
)

func Fingerprint(certPEM string) string {
	b, _ := pem.Decode([]byte(certPEM))
	if b == nil {
		return ""
	}
	sum := sha256.Sum256(b.Bytes)
	return hex.EncodeToString(sum[:])
}
func Generate(host string, ca bool) (string, string, error) {
	key, e := rsa.GenerateKey(rand.Reader, 2048)
	if e != nil {
		return "", "", e
	}
	serial, e := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if e != nil {
		return "", "", e
	}
	cert := &x509.Certificate{SerialNumber: serial, Subject: pkix.Name{CommonName: host}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().AddDate(10, 0, 0), BasicConstraintsValid: true, IsCA: ca, KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment}
	if ca {
		cert.KeyUsage |= x509.KeyUsageCertSign | x509.KeyUsageCRLSign
	} else {
		cert.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
		for _, h := range strings.Split(host, ",") {
			h = strings.TrimSpace(h)
			if ip := net.ParseIP(h); ip != nil {
				cert.IPAddresses = append(cert.IPAddresses, ip)
			} else {
				cert.DNSNames = append(cert.DNSNames, h)
			}
		}
	}
	b, e := x509.CreateCertificate(rand.Reader, cert, cert, &key.PublicKey, key)
	if e != nil {
		return "", "", e
	}
	priv, e := x509.MarshalPKCS8PrivateKey(key)
	if e != nil {
		return "", "", e
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: b})), string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: priv})), nil
}
func IssueClient(caCert, caKey string, appid int64) (string, string, error) {
	pair, e := tls.X509KeyPair([]byte(caCert), []byte(caKey))
	if e != nil {
		return "", "", e
	}
	ca, e := x509.ParseCertificate(pair.Certificate[0])
	if e != nil || !ca.IsCA {
		return "", "", fmt.Errorf("invalid CA")
	}
	key, e := rsa.GenerateKey(rand.Reader, 2048)
	if e != nil {
		return "", "", e
	}
	serial, e := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if e != nil {
		return "", "", e
	}
	cert := &x509.Certificate{SerialNumber: serial, Subject: pkix.Name{CommonName: fmt.Sprintf("app-%d", appid)}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().AddDate(1, 0, 0), KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}}
	b, e := x509.CreateCertificate(rand.Reader, cert, ca, &key.PublicKey, pair.PrivateKey)
	if e != nil {
		return "", "", e
	}
	priv, e := x509.MarshalPKCS8PrivateKey(key)
	if e != nil {
		return "", "", e
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: b})), string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: priv})), nil
}
