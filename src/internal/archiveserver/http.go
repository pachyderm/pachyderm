package archiveserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"go.uber.org/zap"
)

// HTTP is an http.Server that serves the archiveserver endpoints.
type HTTP struct {
	handler *Server        // For testing.
	mux     *http.ServeMux // For testing.
	server  *http.Server   // For ListenAndServe.
}

// NewHTTP creates a new Archive Server and an HTTP server to serve it on.
func NewHTTP(port uint16, pachClientFactory func(ctx context.Context) *client.APIClient) *HTTP {
	mux := http.NewServeMux()
	handler := &Server{
		pachClientFactory: pachClientFactory,
	}
	mux.Handle("/download/", handler)
	mux.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("healthy\n")) //nolint:errcheck
	}))
	return &HTTP{
		handler: handler,
		mux:     mux,
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		},
	}
}

// ListenAndServe begins serving the server, and returns when the context is canceled or the server
// dies on its own.
func (h *HTTP) ListenAndServe(ctx context.Context) error {
	log.AddLoggerToHTTPServer(ctx, "download", h.server)
	errCh := make(chan error, 1)
	go func() {
		errCh <- h.server.ListenAndServe()
	}()
	select {
	case <-ctx.Done():
		log.Info(ctx, "terminating download server", zap.Error(ctx.Err()))
		return errors.EnsureStack(h.server.Shutdown(ctx))
	case err := <-errCh:
		return err
	}
}
