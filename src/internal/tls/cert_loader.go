package tls

import (
	"crypto/tls"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"go.uber.org/zap"
)

// CertLoader provides simple hot TLS certificate reloading by checking for a renewed certificate at a configurable interval
type CertLoader struct {
	certPath        string
	keyPath         string
	refreshInterval time.Duration

	// cert is the current cached *tls.Certificate. It should only be accessed with atomic methods because it may be updated by the cert reloading routine.
	cert     unsafe.Pointer
	stopChan chan interface{}
	stopped  bool
}

// NewCertLoader creates a new CertLoader to refresh the specified TLS key at a fixed interval
func NewCertLoader(certPath, keyPath string, refreshInterval time.Duration) *CertLoader {
	return &CertLoader{
		certPath:        certPath,
		keyPath:         keyPath,
		refreshInterval: refreshInterval,
	}
}

// LoadAndStart ensures the current TLS certificate is loaded and starts the reload routine to poll for renewed certificates
func (l *CertLoader) LoadAndStart() error {
	if err := l.loadCertificate(); err != nil {
		return err
	}
	go l.reloadRoutine()
	return nil
}

// Stop signals the reloading routine to stop
func (l *CertLoader) Stop() {
	if l.stopped {
		return
	}
	l.stopped = true
	close(l.stopChan)
}

// GetCertificate gets the currently cached certificate and fulfills
func (l *CertLoader) GetCertificate(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
	certPtr := atomic.LoadPointer(&l.cert)
	cert := (*tls.Certificate)(certPtr)
	if cert == nil {
		return nil, errors.Errorf("no cached TLS certificate available")
	}
	return cert, nil
}

func (l *CertLoader) reloadRoutine() {
	t := time.NewTicker(l.refreshInterval)
	for {
		select {
		case <-t.C:
			err := l.loadCertificate()
			if err != nil {
				log.Error(pctx.TODO(), "Unable to load TLS certificate", zap.Error(err))
			}
		case <-l.stopChan:
			return
		}
	}
}

func (l *CertLoader) loadCertificate() error {
	log.Debug(pctx.TODO(), "Reloading TLS keypair", zap.String("certPath", l.certPath), zap.String("keyPath", l.keyPath))
	cert, err := tls.LoadX509KeyPair(l.certPath, l.keyPath)
	if err != nil {
		return errors.Wrapf(err, "unable to load keypair")
	}
	atomic.StorePointer(&l.cert, unsafe.Pointer(&cert))
	return nil
}
