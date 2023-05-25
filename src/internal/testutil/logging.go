package testutil

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httputil"
	"testing"
)

func LogHttpResponse(t testing.TB, response *http.Response, msg string) {
	t.Helper()
	dump, err := httputil.DumpResponse(response, true)
	if err != nil {
		t.Log("unable to log response body", msg, err, response)
		return
	}
	t.Log(msg, string(dump))
}

type loggingTransport struct {
	t          testing.TB
	underlying http.RoundTripper
}

func (tr *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	body := req.Body != nil
	outBuf := new(bytes.Buffer)
	if body {
		req.Body = io.NopCloser(io.TeeReader(req.Body, outBuf))
	}
	if b, err := httputil.DumpRequest(req, true); err != nil {
		tr.t.Logf("http: DumpRequest: error: %v", err)
	} else {
		tr.t.Logf("http: request: %s", b)
	}
	if body {
		req.Body = io.NopCloser(outBuf)
	}
	resp, err := tr.underlying.RoundTrip(req)
	if err != nil {
		tr.t.Logf("http: error: %v", err)
		return resp, err //nolint:wrapcheck
	}
	inBuf := new(bytes.Buffer)
	resp.Body = io.NopCloser(io.TeeReader(resp.Body, inBuf))
	if b, err := httputil.DumpResponse(resp, true); err != nil {
		tr.t.Logf("http: DumpResponse: error: %v", err)
	} else {
		tr.t.Logf("http: response: %s", b)
	}
	resp.Body = io.NopCloser(inBuf)
	return resp, err //nolint:wrapcheck
}

func NewLoggingHTTPClient(t testing.TB) *http.Client {
	tr := &loggingTransport{t: t, underlying: http.DefaultTransport}
	return &http.Client{
		Transport: tr,
	}
}
