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

// loggingResponseWriter logs the first 1kb of HTTP response data if the HTTP
// header is an error status (4xx or 5xx)
type loggingResponseWriter struct {
	// w is the actual HTTP ResponseWriter that sends data to the client
	w http.ResponseWriter
	// statusCode is the response code of the request
	statusCode int
	// buf will hold the first 1k of the response body (for logging) if the server
	// returns an error response
	buf []byte
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{
		w: w,
	}
}

func (l *loggingResponseWriter) Header() http.Header {
	return l.w.Header()
}

func (l *loggingResponseWriter) Write(bs []byte) (int, error) {
	if l.statusCode == 0 {
		l.statusCode = 200
	}
	if l.statusCode >= 400 {
		if l.buf == nil {
			// TODO would it be better to statically alloc this? (e.g. [1000]byte in
			// the struct)
			l.buf = make([]byte, 0, 1000)
		}
		if len(l.buf) < cap(l.buf) {
			m := len(l.buf)
			n := min(len(bs), cap(l.buf)-m)
			l.buf = l.buf[:m+n] // extend l.buf
			copy(l.buf[m:], bs[:n])
		}
	}
	return l.w.Write(bs)
}

func (l *loggingResponseWriter) WriteHeader(statusCode int) {
	if l.statusCode == 0 {
		l.statusCode = statusCode
	}
	l.w.WriteHeader(statusCode)
}

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
			lw := newLoggingResponseWriter(w)
			orig.ServeHTTP(lw, r)
			if ua := r.Header.Get("user-agent"); ua != "Envoy/HC" {
				Info(ctx, "http response", zap.String("method", r.Method), zap.String("host", r.Host), zap.String("path", url), zap.String("peer", r.RemoteAddr), id, zap.Int("status-code", lw.statusCode), zap.ByteString("error-msg", lw.buf))
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
