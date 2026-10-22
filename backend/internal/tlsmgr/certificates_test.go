package tlsmgr

import (
	"crypto/tls"
	"crypto/x509"
	"testing"
)

func TestClientCertificateChain(t *testing.T) {
	ca, key, e := Generate("test CA", true)
	if e != nil {
		t.Fatal(e)
	}
	cert, priv, e := IssueClient(ca, key, 1000)
	if e != nil {
		t.Fatal(e)
	}
	pair, e := tls.X509KeyPair([]byte(cert), []byte(priv))
	if e != nil {
		t.Fatal(e)
	}
	leaf, e := x509.ParseCertificate(pair.Certificate[0])
	if e != nil {
		t.Fatal(e)
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM([]byte(ca))
	if _, e = leaf.Verify(x509.VerifyOptions{Roots: pool, KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}}); e != nil {
		t.Fatal(e)
	}
	if leaf.Subject.CommonName != "app-1000" || Fingerprint(cert) == "" {
		t.Fatal("certificate identity missing")
	}
}
