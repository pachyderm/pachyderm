package log

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"runtime/trace"
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
			ctx, task := trace.NewTask(r.Context(), "http server "+r.Method+" "+name)
			defer task.End()

			requestID := r.Header.Values("x-request-id")
			if len(requestID) == 0 {
				requestID = append(requestID, uuid(interactiveTrace).String())
			}
			childFields := []zap.Field{
				zap.Strings("x-request-id", requestID),
			}
			ctx = ChildLogger(ctx, "", WithFields(childFields...))
			ctx = metadata.NewOutgoingContext(ctx, metadata.MD{"x-request-id": requestID})

			// don't log Envoy health checks
			isHealthCheck := r.Header.Get("user-agent") == "Envoy/HC"
			url := r.URL.Path
			if len(url) > 16384 {
				url = url[:16384]
				url += fmt.Sprintf("... (%d bytes)", len(url))
			}
			var requestFields, responseFields []zap.Field
			if !isHealthCheck {
				requestFields = []zap.Field{
					zap.String("path", url),
					zap.String("method", r.Method),
					zap.String("host", r.Host),
					zap.String("peer", r.RemoteAddr),
				}
				Info(ctx, "incoming http request", requestFields...)
			}

			r = r.WithContext(ctx)

			var bodySize, respStatusCode int
			bodyBuf := make([]byte, 0, 1000)
			sw := httpsnoop.Wrap(w, httpsnoop.Hooks{
				WriteHeader: func(next httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
					return func(code int) {
						trace.Logf(ctx, "http", "write headers status=%d", code)
						respStatusCode = code
						next(code)
					}
				},
				Write: func(next httpsnoop.WriteFunc) httpsnoop.WriteFunc {
					return func(b []byte) (int, error) {
						if respStatusCode == 0 {
							respStatusCode = http.StatusOK
						}
						n, err := next(b)
						if len(bodyBuf) < cap(bodyBuf) {
							l := len(bodyBuf)
							m := min(n, cap(bodyBuf)-l)
							bodyBuf = bodyBuf[:l+m] // extend bodyBuf
							copy(bodyBuf[l:], b[:m])
						}
						bodySize += n
						trace.Logf(ctx, "http", "write response %d bytes", len(b))
						return n, err
					}
				},
			})

			orig.ServeHTTP(sw, r)
			if !isHealthCheck {
				responseFields = append(requestFields,
					zap.Int("code", respStatusCode),
					zap.String("headers", fmt.Sprintf("%v", w.Header())),
					zap.Binary("body", bodyBuf),
					zap.Int("bodySizeBytes", bodySize),
				)
				if respStatusCode >= 400 {
					// For error responses, log body again, in cleartext, to make it easy
					// to read/search
					//
					// TODO(msteffen): This was a slightly awkward compromise between
					// logging bodyBuf in base64, so that binary responses don't mess up
					// the logs, and logging bodyBuf in cleartext, to make finding errors
					// easy. See https://github.com/pachyderm/pachyderm/pull/9430#issuecomment-1767570600
					// for the relevant discussion. It may make sense to revise this
					// approach if it's found to be impractical.
					//
					// Alternatives proposed:
					// - Always log the body and always in base64, using zap.Binary
					//   - always works
					// - Always log the body and always in cleartext, using zap.ByteString
					//   - zap.ByteString does appear to escape non-printing chars with
					//     strconv.Quote, but this is not guaranteed in the docs.
					// - Log bodyBuf as cleartext (with key e.g. 'bodyText') if all chars
					//   are printable (per strconv.IsPrint), and as base64 (with key e.g.
					//   'body') otherwise.
					// - Log bodyBuf as cleartext (with key e.g. 'bodyText') if
					//   respStatusCode is an error, and as base64 (with key e.g. 'body')
					//   otherwise.
					responseFields = append(responseFields, zap.ByteString("errorBodyText", bodyBuf))
				}
				Info(ctx, "http response", responseFields...)
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
