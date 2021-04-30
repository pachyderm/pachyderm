package serviceenv

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

// loggingRoundTripper is an internal implementation of http.RoundTripper with
// which we wrap our kubernetes clients' own transport, so that we can log our
// kubernetes requests and responses.
type loggingRoundTripper struct {
	underlying http.RoundTripper
}

// RoundTrip implements the http.RoundTripper interface for loggingRoundTripper
func (t *loggingRoundTripper) RoundTrip(req *http.Request) (res *http.Response, retErr error) {
	logEntry := log.WithFields(map[string]interface{}{
		"type":   "kube/APIServer request",
		"method": req.Method,
		"url":    req.URL.String(),
	})
	logEntry.Debug()
	defer func(start time.Time) {
		fields := map[string]interface{}{
			"type":     "kube/APIServer response",
			"duration": time.Since(start),
			"method":   req.Method,
			"url":      req.URL.String(),
			"err":      retErr,
		}
		if res != nil {
			fields["status"] = res.Status
			if res.StatusCode >= 400 {
				// error response -- log the error message
				bodytext, err := ioutil.ReadAll(res.Body)
				if err != nil {
					fields["body"] = string(bodytext)
				} else {
					fields["body"] = fmt.Sprintf("error reading body: %v", err)
				}
			}
		}
		logEntry := log.WithFields(fields)
		logEntry.Debug("kube/APIServer response")
	}(time.Now())
	return t.underlying.RoundTrip(req)
}

// wrapWithLoggingTransport is k8s.io/client-go/transport.WrapperFunc that wraps
// a pre-existing http.RoundTripper in loggingRoundTripper. We pass this to our
// kubernetes client via rest.Config.WrapTransport to log our kubernetes RPCs
func wrapWithLoggingTransport(rt http.RoundTripper) http.RoundTripper {
	return &loggingRoundTripper{
		underlying: rt,
	}
}
