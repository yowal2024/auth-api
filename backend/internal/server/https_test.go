package server

import (
	"auth-api/internal/repo"
	"auth-api/internal/secretkey"
	"auth-api/internal/tlsmgr"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func listenerDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()
	t.Chdir(t.TempDir())
	if err := repo.InitTablePrefix("a_"); err != nil {
		t.Fatal(err)
	}
	raw, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	db := sqlx.NewDb(raw, "sqlmock")
	t.Cleanup(func() { db.Close() })
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})
	return db, mock
}

func listenerCertificate(t *testing.T) (string, string) {
	t.Helper()
	master, err := secretkey.Load("data/master.key")
	if err != nil {
		t.Fatal(err)
	}
	cert, key, err := tlsmgr.Generate("127.0.0.1,localhost", false)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := secretkey.Encrypt(master, []byte(key))
	if err != nil {
		t.Fatal(err)
	}
	return cert, encoded
}

func expectListenerSettings(mock sqlmock.Sqlmock, enabled bool, legacyAddr, cert, key string) {
	flag := 0
	if enabled {
		flag = 1
	}
	mock.ExpectQuery("SELECT id,web_url").WillReturnRows(sqlmock.NewRows([]string{"id", "https_enabled", "https_addr", "https_cert"}).AddRow(1, flag, legacyAddr, cert))
	if enabled {
		mock.ExpectQuery("SELECT https_key").WillReturnRows(sqlmock.NewRows([]string{"https_key"}).AddRow(key))
	}
}

func listenerClient(t *testing.T, cert string) *http.Client {
	t.Helper()
	transport := &http.Transport{DisableKeepAlives: true}
	if cert != "" {
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM([]byte(cert)) {
			t.Fatal("invalid test certificate")
		}
		transport.TLSClientConfig = &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12}
	}
	t.Cleanup(transport.CloseIdleConnections)
	return &http.Client{Transport: transport, Timeout: 3 * time.Second}
}

func requireListenerResponse(t *testing.T, client *http.Client, url, want string) {
	t.Helper()
	response, err := client.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil || response.StatusCode != http.StatusOK || string(body) != want {
		t.Fatalf("GET %s: status=%d body=%q err=%v", url, response.StatusCode, body, err)
	}
}

func TestListenerSwitchesProtocolOnSamePort(t *testing.T) {
	db, mock := listenerDB(t)
	cert, key := listenerCertificate(t)
	legacyListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { legacyListener.Close() })
	legacyAddr := legacyListener.Addr().String()
	var manager *ListenerManager
	manager = NewListenerManager("127.0.0.1:0", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/switch" {
			if err := manager.Reload(db); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			io.WriteString(w, "switched")
			return
		}
		if r.TLS == nil {
			io.WriteString(w, "HTTP")
		} else {
			io.WriteString(w, "HTTPS")
		}
	}))
	t.Cleanup(manager.Close)
	expectListenerSettings(mock, false, legacyAddr, cert, key)
	if err = manager.Reload(db); err != nil {
		t.Fatal(err)
	}
	addr := manager.Address()
	plain := listenerClient(t, "")
	requireListenerResponse(t, plain, "http://"+addr, "HTTP")

	// Change the protocol from a live request: it must finish without waiting
	// for its own server to shut down, and must ignore the occupied legacy port.
	expectListenerSettings(mock, true, legacyAddr, cert, key)
	requireListenerResponse(t, plain, "http://"+addr+"/switch", "switched")
	if manager.Address() != addr {
		t.Fatalf("HTTPS moved from %s to %s", addr, manager.Address())
	}
	secure := listenerClient(t, cert)
	requireListenerResponse(t, secure, "https://"+addr, "HTTPS")
	response, err := plain.Get("http://" + addr)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("plain HTTP on HTTPS port returned %d", response.StatusCode)
	}

	newCert, newKey := listenerCertificate(t)
	expectListenerSettings(mock, true, legacyAddr, newCert, newKey)
	if err = manager.Reload(db); err != nil {
		t.Fatal(err)
	}
	secure = listenerClient(t, newCert)
	requireListenerResponse(t, secure, "https://"+addr, "HTTPS")

	expectListenerSettings(mock, false, legacyAddr, newCert, newKey)
	requireListenerResponse(t, secure, "https://"+addr+"/switch", "switched")
	requireListenerResponse(t, plain, "http://"+addr, "HTTP")
	if manager.Address() != addr {
		t.Fatal("disabling HTTPS changed the service port")
	}
	if response, err = secure.Get("https://" + addr); err == nil {
		response.Body.Close()
		t.Fatal("HTTP mode still accepted HTTPS")
	}
	manager.Close()
	if err = manager.Reload(nil); !errors.Is(err, http.ErrServerClosed) {
		t.Fatalf("reload after close: %v", err)
	}
	released, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatalf("Close did not release the service port: %v", err)
	}
	released.Close()
}

func TestListenerKeepsHTTPSAfterFailedReload(t *testing.T) {
	db, mock := listenerDB(t)
	cert, key := listenerCertificate(t)
	manager := NewListenerManager("127.0.0.1:0", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.TLS == nil {
			t.Error("request unexpectedly downgraded to HTTP")
		}
		io.WriteString(w, "secure")
	}))
	t.Cleanup(manager.Close)
	expectListenerSettings(mock, true, ":8443", cert, key)
	if err := manager.Reload(db); err != nil {
		t.Fatal(err)
	}
	client := listenerClient(t, cert)
	url := "https://" + manager.Address()
	expectListenerSettings(mock, true, ":8443", "invalid certificate", key)
	if err := manager.Reload(db); err == nil {
		t.Fatal("invalid replacement certificate was accepted")
	}
	requireListenerResponse(t, client, url, "secure")
	if err := os.Remove("data/master.key"); err != nil {
		t.Fatal(err)
	}
	expectListenerSettings(mock, true, ":8443", cert, key)
	if err := manager.Reload(db); err == nil || !strings.Contains(err.Error(), "data/master.key") {
		t.Fatalf("missing master key: %v", err)
	}
	if _, err := os.Stat("data/master.key"); !os.IsNotExist(err) {
		t.Fatal("a missing master key was silently replaced")
	}
	requireListenerResponse(t, client, url, "secure")
}

func TestListenerRejectsInvalidHTTPSAtStartup(t *testing.T) {
	db, mock := listenerDB(t)
	manager := NewListenerManager("127.0.0.1:0", http.NotFoundHandler())
	t.Cleanup(manager.Close)
	expectListenerSettings(mock, true, ":8443", "", "")
	if err := manager.Reload(db); err == nil {
		t.Fatal("HTTPS started without a certificate")
	}
	if manager.current != nil {
		t.Fatal("invalid HTTPS configuration opened an HTTP listener")
	}
}

func TestListenerReportsBindAndServeFailures(t *testing.T) {
	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer occupied.Close()
	blocked := NewListenerManager(occupied.Addr().String(), http.NotFoundHandler())
	t.Cleanup(blocked.Close)
	if err = blocked.Reload(nil); err == nil {
		t.Fatal("occupied port did not return a startup error")
	}
	manager := NewListenerManager("127.0.0.1:0", http.NotFoundHandler())
	t.Cleanup(manager.Close)
	if err = manager.Reload(nil); err != nil {
		t.Fatal(err)
	}
	manager.mu.Lock()
	manager.current.listener.Close()
	manager.mu.Unlock()
	select {
	case err := <-manager.Errors():
		if err == nil {
			t.Fatal("listener failure was not reported")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("listener failure left the process waiting without a listener")
	}
}
