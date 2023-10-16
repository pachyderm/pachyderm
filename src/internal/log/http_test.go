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
			if r.URL.Path == "/tea" {
				w.Write([]byte("OK"))
				return
			}
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

	t.Run("HealthyResponsesDontLogBody", func(t *testing.T) {
		h.Clear()
		if _, err := http.Get("http://" + l.Addr().String() + "/tea"); err != nil {
			t.Fatalf("get: %v", err)
		}

		want := []string{
			"http: info: incoming http request",
			"http: info: http response",
		}
		if diff := cmp.Diff(h.Logs(), want, formatLogs(simple)); diff != "" {
			t.Errorf("logs (-got +want):\n%s", diff)
		}
		for _, l := range h.Logs() {
			if rawStatus, ok := l.Keys["status-code"]; ok {
				status, ok := rawStatus.(float64)
				if !ok {
					t.Errorf("expected http error msg to be float64, but was: %T", rawStatus)
				}
				if status != 200 {
					// can't use require.Equal due to import cycle
					t.Errorf("expected status-code 200, but was: %d", int(status))
				}
			}
			if rawErrmsg, ok := l.Keys["err-msg"]; ok {
				errmsg, ok := rawErrmsg.(string)
				if !ok {
					t.Errorf("expected http error msg to be string, but was: %T", rawErrmsg)
				}
				if errmsg != "" {
					t.Errorf("expected empty http error msg, but was:\n%s\n(len: %d)", errmsg, len(errmsg))
				}
			}
		}
	})
	t.Run("ErrorResponsesLogged", func(t *testing.T) {
		h.Clear()
		if _, err := http.Get("http://" + l.Addr().String() + "/coffee"); err != nil {
			t.Fatalf("get: %v", err)
		}

		want := []string{
			"http: info: incoming http request",
			"http: info: http response",
		}
		if diff := cmp.Diff(h.Logs(), want, formatLogs(simple)); diff != "" {
			t.Errorf("logs (-got +want):\n%s", diff)
		}
		for _, l := range h.Logs() {
			if rawStatus, ok := l.Keys["status-code"]; ok {
				status, ok := rawStatus.(float64)
				if !ok {
					t.Errorf("expected http error msg to be float64, but was: %T", rawStatus)
				}
				if status != http.StatusTeapot {
					t.Errorf("expected status-code %d, but was: %d", http.StatusTeapot, int(status))
				}
			}
			if rawErrmsg, ok := l.Keys["err-msg"]; ok {
				errmsg, ok := rawErrmsg.(string)
				if !ok {
					t.Errorf("expected http error msg to be string, but was: %T", rawErrmsg)
				}
				if errmsg != strings.Repeat("a", 1000) {
					t.Errorf("expected http error msg to be a*1000, but was:\n%s\n(len: %d)", errmsg, len(errmsg))
				}
			}
		}
	})
}
