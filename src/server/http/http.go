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
	"github.com/pachyderm/pachyderm/v2/src/internal/jsonschema"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"go.uber.org/zap"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	pfsgw "github.com/pachyderm/pachyderm/v2/src/pfs"
	ppsgw "github.com/pachyderm/pachyderm/v2/src/pps"

	admingw "github.com/pachyderm/pachyderm/v2/src/admin"
	authgw "github.com/pachyderm/pachyderm/v2/src/auth"
	debuggw "github.com/pachyderm/pachyderm/v2/src/debug"
	enterprisegw "github.com/pachyderm/pachyderm/v2/src/enterprise"
	identitygw "github.com/pachyderm/pachyderm/v2/src/identity"
	licensegw "github.com/pachyderm/pachyderm/v2/src/license"
	proxygw "github.com/pachyderm/pachyderm/v2/src/proxy"
	transactiongw "github.com/pachyderm/pachyderm/v2/src/transaction"
	versiongw "github.com/pachyderm/pachyderm/v2/src/version/versionpb"
	workergw "github.com/pachyderm/pachyderm/v2/src/worker"

)

// Server is an http.Server that serves public requests.
type Server struct {
	mux    http.Handler // For testing.
	server *http.Server // For ListenAndServe.
}

// New creates a new API server, with an http.Server to actually serve traffic.
func New(port uint16, pachClientFactory func(ctx context.Context) *client.APIClient) *Server {
	mux := http.NewServeMux()

	// Archive server.
	handler := &archiveserver.Server{
		ClientFactory: pachClientFactory,
	}
	mux.Handle("/archive/", CSRFWrapper(handler))

	// Health check.
	mux.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("healthy\n")) //nolint:errcheck
	}))

	// JSON schemas.
	mux.Handle("/jsonschema/", http.StripPrefix("/jsonschema/", http.FileServer(http.FS(jsonschema.FS))))

	log.Info(context.Background(), "httpserver: start restapi registrations with new gateway mux")
	// GRPC gateway.
	gwmux := NewGatewayMux(pachClientFactory)
	mux.Handle("/api/", http.StripPrefix("/api", gwmux))

	return &Server{
		mux: mux,
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		},
	}
}

func NewGatewayMux(pachClientFactory func(context.Context) *client.APIClient) http.Handler {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	log.Info(ctx, "restapi: get pach client")
	client := pachClientFactory(ctx)

	log.Info(ctx, "restapi: create new mux that passes all headers except contentlength")
	mux := runtime.NewServeMux(runtime.WithIncomingHeaderMatcher(func(s string) (string, bool) {
		if s != "Content-Length" {
			return s, true
		}
		return s, false
	}))

	err := ppsgw.RegisterAPIHandler(ctx, mux, client.ClientConn())
	if err != nil {
		return nil
	}

	err = pfsgw.RegisterAPIHandler(ctx, mux, client.ClientConn())
	if err != nil {
		return nil
	}

	err = workergw.RegisterWorkerHandler(ctx, mux, client.ClientConn())
	if err != nil {
		return nil
	}

	err = proxygw.RegisterAPIHandler(ctx, mux, client.ClientConn())
	if err != nil {
		return nil
	}

	err = admingw.RegisterAPIHandler(ctx, mux, client.ClientConn())
	if err != nil {
		return nil
	}

	err = authgw.RegisterAPIHandler(ctx, mux, client.ClientConn())
	if err != nil {
		return nil
	}

	err = licensegw.RegisterAPIHandler(ctx, mux, client.ClientConn())
	if err != nil {
		return nil
	}

	err = identitygw.RegisterAPIHandler(ctx, mux, client.ClientConn())
	if err != nil {
		return nil
	}

	err = debuggw.RegisterDebugHandler(ctx, mux, client.ClientConn())
	if err != nil {
		return nil
	}

	err = enterprisegw.RegisterAPIHandler(ctx, mux, client.ClientConn())
	if err != nil {
		return nil
	}

	err = transactiongw.RegisterAPIHandler(ctx, mux, client.ClientConn())
	if err != nil {
		return nil
	}

	log.Info(ctx, "restapi: register version api handler")
	err = versiongw.RegisterAPIHandler(ctx, mux, client.ClientConn())
	if err != nil {
		log.Info(ctx, "restapi: version api registration failed")
		return nil
	}
	log.Info(ctx, "restapi: version api registration success")

	return mux
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
func (h *Server) ListenAndServe(ctx context.Context) error {
	log.AddLoggerToHTTPServer(ctx, "pachhttp", h.server)
	errCh := make(chan error, 1)
	go func() {
		errCh <- h.server.ListenAndServe()
	}()
	select {
	case <-ctx.Done():
		log.Info(ctx, "terminating pachhttp server", zap.Error(context.Cause(ctx)))
		return errors.EnsureStack(h.server.Shutdown(ctx))
	case err := <-errCh:
		return err
	}
}
