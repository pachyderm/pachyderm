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
	sync.Mutex
	tlsConfig     *tls.Config
	underlying    *http.Transport
	inFlight      map[cancelKey]interface{}
	cancellations int
}

func newReconnectTransport(tlsConfig *tls.Config) *reconnectTransport {
	t := &reconnectTransport{
		tlsConfig: tlsConfig,
		inFlight:  make(map[cancelKey]interface{}),
	}
	t.underlying = t.newUnderlying()
	return t
}

// RoundTrip implements the http.RoundTripper interface for loggingRoundTripper
func (t *reconnectTransport) RoundTrip(req *http.Request) (res *http.Response, retErr error) {
	// Acquire a lock to get the current transport and register the request
	t.Lock()
	trans := t.underlying
	t.inFlight[cancelKey{req}] = 1
	t.Unlock()

	// When the request is complete, remove it from the set of in-flight requests
	defer func() {
		t.Lock()
		delete(t.inFlight, cancelKey{req})
		t.Unlock()
	}()
	return trans.RoundTrip(req)
}

func (t *reconnectTransport) CancelRequest(req *http.Request) {
	t.Lock()
	defer t.Unlock()

	// the request was for an older underlying transport
	if _, ok := t.inFlight[cancelKey{req}]; !ok {
		log.Infof("got cancelrequest but request is already over")
		return
	}

	// Cancel the underlying request and add it to the count for this connection
	log.Errorf("cancelling request: %v/%v", t.cancellations, cancellationsBeforeReset)
	t.underlying.CancelRequest(req)
	t.cancellations++

	if t.cancellations >= cancellationsBeforeReset {
		log.Errorf("recreating transport due to too many timed out connections")
		// Close all keep-alive connections in the current transport.
		// TOOD: this doesn't cancel the outstanding requests, we should track them and cancel them all here too.
		t.underlying.CloseIdleConnections()
		for inflightReq, _ := range t.inFlight {
			t.underlying.CancelRequest(inflightReq.req)
		}
		t.cancellations = 0
		t.inFlight = make(map[cancelKey]interface{})
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

type cancelKey struct {
	req *http.Request
}
