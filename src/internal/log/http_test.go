package log

import (
	"context"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
)

func TestHTTPServer(t *testing.T) {
	ctx, h := testWithCaptureParallel(t, zap.Development())
	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Info(r.Context(), "log from handler")
			w.WriteHeader(http.StatusNoContent)
		}),
	}
	AddLoggerToHTTPServer(ctx, "http", server)

	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Logf("listening on %v", l.Addr().String())
	go server.Serve(l)                                          //nolint:errcheck
	t.Cleanup(func() { server.Shutdown(context.Background()) }) //nolint:errcheck

	if _, err := http.Get("http://" + l.Addr().String()); err != nil {
		t.Fatalf("get: %v", err)
	}

	want := []string{
		"http: info: incoming http request",
		"http: info: log from handler",
		"http: info: http response",
	}
	if diff := cmp.Diff(h.Logs(), want, formatLogs(simple)); diff != "" {
		t.Errorf("logs (-got +want):\n%s", diff)
	}
}

func TestHTTPServerError(t *testing.T) {
	ctx, h := testWithCaptureParallel(t, zap.Development())
	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// test that response log lines only include the first 1000b of the
			// response body.
			http.Error(w, strings.Repeat("a", 2000), http.StatusTeapot)
		}),
	}
	AddLoggerToHTTPServer(ctx, "http", server)

	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Logf("listening on %v", l.Addr().String())
	go server.Serve(l)                                          //nolint:errcheck
	t.Cleanup(func() { server.Shutdown(context.Background()) }) //nolint:errcheck

	if _, err := http.Get("http://" + l.Addr().String() + "/brew?drink=coffee"); err != nil {
		t.Fatalf("get: %v", err)
	}

	want := []string{
		"http: info: incoming http request",
		"http: info: http response",
	}
	if diff := cmp.Diff(h.Logs(), want, formatLogs(simple)); diff != "" {
		t.Errorf("logs (-got +want):\n%s", diff)
	}
	/*
		type msg struct {
			Time     time.Time
			Severity string
			Logger   string
			Caller   string
			Message  string
			Keys     map[string]any
			Orig     json.RawMessage
		}
	*/
	for _, l := range h.Logs() {
		if errmsg, ok := l.Keys["err-msg"]; ok {
			errmsg, ok := errmsg.(string)
			if !ok {
				t.Errorf("expected http error msg to be string, but was: %T", errmsg)
			}
			if errmsg != strings.Repeat("a", 1000) {
				// can't use require.Equal due to import cycle
				t.Errorf("expected http error msg to be a*1000, but was:\n%s\n(len: %d)", errmsg, len(errmsg))
			}
		}
	}
}
