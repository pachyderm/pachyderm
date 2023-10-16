package log

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/felixge/httpsnoop"
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

			var url string
			if ua := r.Header.Get("user-agent"); ua != "Envoy/HC" {
				// Print info about the request if this is not an Envoy health check.
				url = r.URL.Path
				if len(url) > 16384 {
					url = url[:16384]
					url += fmt.Sprintf("... (%d bytes)", len(url))
				}
				Info(ctx, "incoming http request", zap.String("path", url), zap.String("method", r.Method), zap.String("host", r.Host), zap.String("peer", r.RemoteAddr), id)
			}

			ctx = ChildLogger(ctx, "", WithFields(id))
			ctx = metadata.NewOutgoingContext(ctx, metadata.MD{"x-request-id": requestID})
			r = r.WithContext(ctx)

			var respStatusCode int
			var errbuf []byte
			w = httpsnoop.Wrap(w, httpsnoop.Hooks{
				WriteHeader: func(next httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
					return func(code int) {
						respStatusCode = code
						next(code)
					}
				},
				Write: func(next httpsnoop.WriteFunc) httpsnoop.WriteFunc {
					return func(b []byte) (int, error) {
						if respStatusCode == 0 {
							respStatusCode = 200
						}
						if respStatusCode >= 400 {
							if errbuf == nil {
								// TODO would it be better to statically alloc this? (e.g.
								// [1000]byte above?)
								errbuf = make([]byte, 0, 1000)
							}
							if len(errbuf) < cap(errbuf) {
								m := len(errbuf)
								n := min(len(b), cap(errbuf)-m)
								errbuf = errbuf[:m+n] // extend errbuf
								copy(errbuf[m:], b[:n])
							}
						}
						return next(b)
					}
				},
			})

			orig.ServeHTTP(w, r)
			if ua := r.Header.Get("user-agent"); ua != "Envoy/HC" {
				Info(ctx, "http response", zap.String("method", r.Method), zap.String("host", r.Host), zap.String("path", url), zap.String("peer", r.RemoteAddr), id, zap.Int("status-code", respStatusCode), zap.ByteString("error-msg", errbuf))
			}
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
