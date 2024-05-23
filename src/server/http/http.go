// Package http is a browser-targeted HTTP server for Pachyderm.  It includes what used to be called
// the "archive server" (for /archive/ downloads) and the "file server" (for /pfs/ downloads).  Any
// future user-facing HTTP services should be added here.
//
// The general design here is for the servers to be relatively untrusted.  They build a PachClient
// based on the HTTP request, and then make ordinary API calls.
package http

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pachyderm/pachyderm/v2/src/internal/archiveserver"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/fileserver"
	"github.com/pachyderm/pachyderm/v2/src/internal/jsonschema"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/restgateway"
	"go.uber.org/zap"
)

// Server is an http.Server that serves public requests.
type Server struct {
	mux    http.Handler // For testing.
	Server *http.Server // For ListenAndServe.
}

// New creates a new API server, with an http.Server to actually serve traffic.
func New(ctx context.Context, port uint16, pachClientFactory func(ctx context.Context) *client.APIClient) (*Server, error) {
	ctx = pctx.Child(ctx, "httpserver")
	mux := http.NewServeMux()

	// Archive server.
	handler := &archiveserver.Server{
		ClientFactory: pachClientFactory,
	}
	mux.Handle("/archive/", CSRFWrapper(handler))

	// File server.
	fileHandler := &fileserver.Server{
		ClientFactory: pachClientFactory,
	}
	mux.Handle("/pfs/", CSRFWrapper(fileHandler))

	// Health check.
	mux.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("healthy\n")) //nolint:errcheck
	}))

	// JSON schemas.
	mux.Handle("/jsonschema/", http.StripPrefix("/jsonschema/", http.FileServer(http.FS(jsonschema.FS))))

	// GRPC gateway.
	client := pachClientFactory(ctx)
	gwmux, err := restgateway.NewMux(ctx, client.ClientConn())
	if err != nil {
		return nil, errors.Wrap(err, "init rest gateway mux")
	}
	mux.Handle("/api/", http.StripPrefix("/api", gwmux))
	log.Info(ctx, "completed grpc gateway rest api registrations")

	return &Server{
		mux: mux,
		Server: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		},
	}, nil
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
			log.Info(r.Context(), "csrf: origin/host mismatch; delete cookies", zap.String("resolved_origin", origin), zap.Strings("origin", r.Header.Values("origin")), zap.Strings("referer", r.Header.Values("referer")), zap.String("host", r.Host))
			r.Header.Del("cookie")
		}
		// Origin and Host match; allow.
		h.ServeHTTP(w, r)
	}
}

// ListenAndServe begins serving the server, and returns when the context is canceled or the server
// dies on its own.
func (h *Server) ListenAndServe(ctx context.Context) error {
	log.AddLoggerToHTTPServer(ctx, "pachhttp", h.Server)
	errCh := make(chan error, 1)
	go func() {
		errCh <- h.Server.ListenAndServe()
	}()
	select {
	case <-ctx.Done():
		log.Info(ctx, "terminating pachhttp server", zap.Error(context.Cause(ctx)))
		return errors.EnsureStack(h.Server.Shutdown(ctx))
	case err := <-errCh:
		return err
	}
}
