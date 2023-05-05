package archiveserver

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"go.uber.org/zap"
)

// HTTP is an http.Server that serves the archiveserver endpoints.
type HTTP struct {
	mux    http.Handler // For testing.
	server *http.Server // For ListenAndServe.
}

// NewHTTP creates a new Archive Server and an HTTP server to serve it on.
func NewHTTP(port uint16, pachClientFactory func(ctx context.Context) *client.APIClient) *HTTP {
	mux := http.NewServeMux()
	handler := &Server{
		pachClientFactory: pachClientFactory,
	}
	mux.Handle("/archive/", CSRFWrapper(handler))
	mux.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("healthy\n")) //nolint:errcheck
	}))
	return &HTTP{
		mux: mux,
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		},
	}
}

// CSRFWrapper is an http.Handler that provides CSRF protection to the underlying handler.
func CSRFWrapper(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("origin")
		if origin != "" {
			u, err := url.Parse(origin)
			if err != nil {
				origin = "error(origin): " + err.Error()
			} else if u.Host == "" {
				origin = "error(origin): no host"
			} else {
				origin = u.Host
			}
		} else {
			// No Origin header, try Referer.
			if r := r.Header.Get("referer"); r != "" {
				u, err := url.Parse(r)
				if err != nil {
					origin = "error(referer): " + err.Error() // We must deny in this case.
				} else if u.Host == "" {
					origin = "error(referer): no host"
				} else {
					origin = u.Host
				}
			}
			// Neither header exists.
		}
		if origin == "" {
			log.Debug(r.Context(), "csrf: no origin or referer header; assuming cli; allow")
			h.ServeHTTP(w, r)
			return
		}
		if origin != r.Host {
			log.Info(r.Context(), "csrf: origin/host mismatch; deny", zap.String("resolved_origin", origin), zap.Strings("origin", r.Header.Values("origin")), zap.Strings("referer", r.Header.Values("referer")), zap.String("host", r.Host))
			http.Error(w, "csrf: origin/host mismatch", http.StatusForbidden)
			return
		}
		// Origin and Host match; allow.
		h.ServeHTTP(w, r)
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
