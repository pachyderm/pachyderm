package log

import (
	"context"
	"net"
	"net/http"
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
	}
	if diff := cmp.Diff(h.Logs(), want, formatLogs(simple)); diff != "" {
		t.Errorf("logs (-got +want):\n%s", diff)
	}
}
