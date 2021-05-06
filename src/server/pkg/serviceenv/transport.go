package serviceenv

import (
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	utilnet "k8s.io/apimachinery/pkg/util/net"
)

const cancellationsBeforeReset = 4

type reconnectTransport struct {
	sync.RWMutex
	tlsConfig     *tls.Config
	underlying    *http.Transport
	cancellations int
}

func newReconnectTransport(tlsConfig *tls.Config) *reconnectTransport {
	t := &reconnectTransport{
		tlsConfig: tlsConfig,
	}
	t.underlying = t.newUnderlying()
	return t
}

// RoundTrip implements the http.RoundTripper interface for loggingRoundTripper
func (t *reconnectTransport) RoundTrip(req *http.Request) (res *http.Response, retErr error) {
	// Acquire a read lock - block if we're currently resetting the underlying transport
	t.RLock()
	trans := t.underlying
	t.RUnlock()
	return trans.RoundTrip(req)
}

func (t *reconnectTransport) CancelRequest(req *http.Request) {
	// Acquire a write lock to cancel - this is probably overkill
	t.Lock()
	defer t.Unlock()
	log.Errorf("cancelling request: %v/%v", t.cancellations, cancellationsBeforeReset)

	// Cancel the underlying request
	t.underlying.CancelRequest(req)
	t.cancellations++

	if t.cancellations >= cancellationsBeforeReset {
		log.Errorf("recreating transport due to too many timed out connections")
		// Close all keep-alive connections in the current transport.
		// TOOD: this doesn't cancel the outstanding requests, we should track them and cancel them all here too.
		t.underlying.CloseIdleConnections()
		t.cancellations = 0
		t.underlying = t.newUnderlying()
	}
}

func (t *reconnectTransport) newUnderlying() *http.Transport {
	dial := (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext

	return utilnet.SetTransportDefaults(&http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     t.tlsConfig,
		MaxIdleConnsPerHost: 25,
		DialContext:         dial,
		DisableCompression:  false,
		TLSNextProto:        make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
	})
}
