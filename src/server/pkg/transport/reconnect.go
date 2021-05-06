package transport

import (
	"net/http"
)

// This is a copy of an internal interface used by k8s.io/client-go when using
// a passed-in http.RoundTripper. We only need it in order to wrap client-go's
// RoundTripper with logging.
type kubeRequestCanceler interface {
	CancelRequest(*http.Request)
}

// ReconnectOnCancelRoundTripper is an http.RoundTripper that wraps a k8s
// client round-tripper (that implements `kubeRequestCanceller`) and reconnects
// after a certain number of failures. This is necessary to recover from a
// condition where clients appear to become stuck and all API requests
// time out.
type ReconnectOnCancelRoundTripper struct {
	sync.RWMutex
	cancellations   uint64
	retries         uint64
	underlying      http.RoundTripper
	TLSClientConfig *tls.Config
}

// NewReconnectOnCancelRoundTripper creates an http.RoundTripper that allows
// `retries` number of timeouts before calling `reconnectFn`
func NewReconnectOnCancelRoundTripper(retries uint64, tlsConfig *tls.Config) {
	return &ReconnectOnCancelRoundTripper{
		retries:     retries,
		reconnectFn: reconnectFn,
	}
}

// RoundTrip implements the http.RoundTripper interface for loggingRoundTripper
func (t *ReconnectOnCancelRoundTripper) RoundTrip(req *http.Request) (res *http.Response, retErr error) {
	return t.underlying.RoundTrip(req)
}

// CancelRequest is the method called by the kube client when the request times out
func (t *ReconnectOnCanelRoundTripper) CancelRequest(req *http.Request) {
	log.Errorf("request cancelled, %v/%v", t.cancellations, t.retries)

	c, ok := t.underlying.(kubeRequestCanceler)
	if !ok {
		log.Errorf("underlying RoundTripper %T does not implement CancelRequest", t.underlying)
		return
	}

	c.CancelRequest(req)

	// This can be called by more than one goroutine at the same time -
	// once we've reached the limit for cancellations we should block
	// until reconnectFn succeeds and then have them all retry at a higher
	// level.
	cancellations := atomic.AddUint64(&t.cancellations, 1)
	if cancellations >= t.retries {

		log.Errorf("too many requests cancelled, reconnecting")
		if err := t.reconnectFn(); err != nil {
			log.WithError(err).Errorf("unable to reconnect, retrying on next request")
		} else {
			t.cancellations = 0
		}
	}
}
