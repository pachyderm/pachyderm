package restgateway

import (
	"context"
	"flag"
	"net/http"
	"net/url"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"go.uber.org/zap"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

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

var (
	// pachd-grpc server port 1650
	grpcServerEndpoint = flag.String("grpc-server-endpoint", "localhost:1650", "gRPC server endpoint")
)

// Server is an http.Server that serves public requests.
type Server struct {
	mux    runtime.ServeMux
	server *http.Server
}

// New creates a new API server, with an http.Server to actually serve traffic.
func New(port uint16, pachClientFactory func(ctx context.Context) *client.APIClient) *runtime.ServeMux {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	client := pachClientFactory(ctx)

	mux := runtime.NewServeMux(runtime.WithIncomingHeaderMatcher(func(s string) (string, bool) {
		if s != "Content-Length" {
			return s, true
		}
		return s, false
	}))

	// Health check.
	//mux.HandlePath("POST", "/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	mux.HandlePath("GET", "/healthz", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("healthy\n")) //nolint:errcheck
	})

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err := ppsgw.RegisterAPIHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts)
	if err != nil {
		return nil
	}

	err = pfsgw.RegisterAPIHandler(ctx, mux, client.ClientConn())
	if err != nil {
		return nil
	}

	err = workergw.RegisterWorkerHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts)
	if err != nil {
		return nil
	}

	err = proxygw.RegisterAPIHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts)
	if err != nil {
		return nil
	}

	err = admingw.RegisterAPIHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts)
	if err != nil {
		return nil
	}

	err = authgw.RegisterAPIHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts)
	if err != nil {
		return nil
	}

	err = licensegw.RegisterAPIHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts)
	if err != nil {
		return nil
	}

	err = identitygw.RegisterAPIHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts)
	if err != nil {
		return nil
	}

	err = debuggw.RegisterDebugHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts)
	if err != nil {
		return nil
	}

	err = enterprisegw.RegisterAPIHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts)
	if err != nil {
		return nil
	}

	err = transactiongw.RegisterAPIHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts)
	if err != nil {
		return nil
	}

	err = versiongw.RegisterAPIHandler(ctx, mux, client.ClientConn())
	if err != nil {
		return nil
	}
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
	log.AddLoggerToHTTPServer(ctx, "restgateway", h.server)
	errCh := make(chan error, 1)
	go func() {
		errCh <- h.server.ListenAndServe()
	}()
	select {
	case <-ctx.Done():
		log.Info(ctx, "restgateway server exiting", zap.Error(context.Cause(ctx)))
		return errors.EnsureStack(h.server.Shutdown(ctx))
	case err := <-errCh:
		return err
	}
}
