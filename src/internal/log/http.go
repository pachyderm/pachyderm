package log

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

// AddLoggerToHTTPServer mutates the provided *http.Server so that (*http.Request).Context() returns
// a context that can be logged to.
//
// TODO(jonathan): Move this to pctx.
func AddLoggerToHTTPServer(rctx context.Context, name string, s *http.Server) {
	ctx := ChildLogger(rctx, name, WithServerID())
	s.BaseContext = func(l net.Listener) context.Context {
		return ctx
	}
	s.ErrorLog = NewStdLogAt(ctx, DebugLevel)
	if s.Handler != nil {
		orig := s.Handler
		s.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			requestID := r.Header.Values("x-request-id")
			if len(requestID) == 0 {
				requestID = append(requestID, uuid(interactiveTrace).String())
			}
			id := zap.Strings("x-request-id", requestID)

			if ua := r.Header.Get("user-agent"); ua != "Envoy/HC" {
				// Print info about the request if this is not an Envoy health check.
				url := r.URL.Path
				if len(url) > 16384 {
					url = url[:16384]
					url += fmt.Sprintf("... (%d bytes)", len(url))
				}
				Info(ctx, "incoming http request", zap.String("path", url), zap.String("method", r.Method), zap.String("host", r.Host), zap.String("peer", r.RemoteAddr), id)
			}

			ctx = ChildLogger(ctx, "", WithFields(id))
			ctx = metadata.NewOutgoingContext(ctx, metadata.MD{"x-request-id": requestID})
			r = r.WithContext(ctx)
			orig.ServeHTTP(w, r)
		})
	}
}

// RequestID returns the RequestID associated with this HTTP request.  This is added by the
// middleware above.
func RequestID(ctx context.Context) string {
	if md, ok := metadata.FromOutgoingContext(ctx); ok {
		if parts := md.Get("x-request-id"); len(parts) > 0 {
			return strings.Join(parts, ";")
		}
	}
	return ""
}
