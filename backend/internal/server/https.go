package server

import (
	"auth-api/internal/repo"
	"auth-api/internal/secretkey"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jmoiron/sqlx"
)

// ListenerManager serves HTTP or HTTPS on the same configured address, matching
// the original executable. The legacy settings.https_addr is not a second port.
type ListenerManager struct {
	mu      sync.Mutex
	addr    string
	handler http.Handler
	current *listenerState
	closed  bool
	errors  chan error
	workers sync.WaitGroup
}

type listenerState struct {
	server      *http.Server
	listener    net.Listener
	tls         bool
	certificate atomic.Pointer[tls.Certificate]
}

func NewListenerManager(addr string, handler http.Handler) *ListenerManager {
	return &ListenerManager{addr: addr, handler: handler, errors: make(chan error, 1)}
}

func (m *ListenerManager) Errors() <-chan error { return m.errors }

func (m *ListenerManager) Address() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.current != nil {
		return m.current.listener.Addr().String()
	}
	return m.addr
}

// Reload validates the new TLS settings before changing the active listener.
// A nil database starts the HTTP installation wizard.
func (m *ListenerManager) Reload(db *sqlx.DB) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return http.ErrServerClosed
	}
	cert, err := loadServerCertificate(db)
	if err != nil {
		return err
	}
	old := m.current
	if old != nil && old.tls == (cert != nil) {
		// Certificate rotation does not need to interrupt existing connections.
		old.certificate.Store(cert)
		return nil
	}
	addr := m.addr
	if old != nil {
		// Keep the actual port even when the initial address requested port zero.
		addr = old.listener.Addr().String()
		m.current = nil
		_ = old.listener.Close()
		m.drainLocked(old)
	}
	if err = m.startLocked(addr, cert); err != nil {
		if old != nil {
			// If rebinding fails, try to preserve the previous working protocol.
			if restoreErr := m.startLocked(addr, old.certificate.Load()); restoreErr != nil {
				m.report(fmt.Errorf("restore listener after protocol change: %w", restoreErr))
			}
		}
		return err
	}
	return nil
}

func loadServerCertificate(db *sqlx.DB) (*tls.Certificate, error) {
	if db == nil {
		return nil, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	settings, err := repo.NewSettingsRepo(db).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("read HTTPS settings: %w", err)
	}
	if settings.HTTPSEnabled != 1 {
		return nil, nil
	}
	var encoded string
	if err = db.GetContext(ctx, &encoded, repo.SQL("SELECT https_key FROM {settings} WHERE id=1")); err != nil {
		return nil, fmt.Errorf("read HTTPS private key: %w", err)
	}
	if settings.HTTPSCert == "" || encoded == "" {
		return nil, errors.New("HTTPS is enabled but its certificate or private key is missing")
	}
	// Never generate a replacement key for an existing encrypted certificate.
	master, err := os.ReadFile("data/master.key")
	if err != nil {
		return nil, fmt.Errorf("read HTTPS master key data/master.key: %w", err)
	}
	key, err := secretkey.Decrypt(master, encoded)
	if err != nil {
		return nil, fmt.Errorf("decrypt HTTPS private key with data/master.key: %w", err)
	}
	cert, err := tls.X509KeyPair([]byte(settings.HTTPSCert), key)
	if err != nil {
		return nil, fmt.Errorf("load HTTPS certificate: %w", err)
	}
	return &cert, nil
}

// The caller holds m.mu until the new server becomes the active generation.
func (m *ListenerManager) startLocked(addr string, cert *tls.Certificate) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}
	s := &listenerState{listener: listener, tls: cert != nil}
	s.certificate.Store(cert)
	s.server = &http.Server{Addr: addr, Handler: m.handler, ReadHeaderTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second}
	protocol := "HTTP"
	if s.tls {
		protocol = "HTTPS"
		s.server.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			ClientAuth: tls.RequestClientCert,
			GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
				return s.certificate.Load(), nil
			},
		}
	}
	m.current = s
	m.workers.Add(1)
	go func() {
		defer m.workers.Done()
		var serveErr error
		if s.tls {
			serveErr = s.server.ServeTLS(listener, "", "")
		} else {
			serveErr = s.server.Serve(listener)
		}
		m.mu.Lock()
		unexpected := !m.closed && m.current == s
		m.mu.Unlock()
		if unexpected && serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			m.report(fmt.Errorf("%s listener %s stopped: %w", protocol, addr, serveErr))
		}
	}()
	log.Printf("%s listening on %s", protocol, listener.Addr())
	return nil
}

// Drain asynchronously: Reload can be called by a request on the old server,
// and waiting for that request's Shutdown here would deadlock its response.
func (m *ListenerManager) drainLocked(s *listenerState) {
	s.server.SetKeepAlivesEnabled(false)
	m.workers.Add(1)
	go func() {
		defer m.workers.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.server.Shutdown(ctx); err != nil {
			_ = s.server.Close()
		}
	}()
}

func (m *ListenerManager) report(err error) {
	select {
	case m.errors <- err:
	default:
	}
}

func (m *ListenerManager) Close() {
	m.mu.Lock()
	m.closed = true
	if m.current != nil {
		s := m.current
		m.current = nil
		_ = s.listener.Close()
		m.drainLocked(s)
	}
	m.mu.Unlock()
	m.workers.Wait()
}
