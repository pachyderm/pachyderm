package serviceenv

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"

	utilnet "k8s.io/apimachinery/pkg/util/net"
)

const cancellationsBeforeReset = 4

type timeoutTransport struct {
	sync.Mutex
	timeout    time.Duration
	tlsConfig  *tls.Config
	underlying *http.Transport
}

func newTimeoutTransport(timeout time.Duration, tlsConfig *tls.Config) *timeoutTransport {
	t := &timeoutTransport{
		tlsConfig: tlsConfig,
		timeout:   timeout,
	}
	t.underlying = t.newUnderlying()
	return t
}

// RoundTrip implements the http.RoundTripper interface for loggingRoundTripper
func (t *timeoutTransport) RoundTrip(req *http.Request) (res *http.Response, retErr error) {
	// Just add a context with a timeout to see if that works
	ctx, _ := context.WithTimeout(context.Background(), t.timeout)
	return t.underlying.RoundTrip(req.WithContext(ctx))
}

func (t *timeoutTransport) newUnderlying() *http.Transport {
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
