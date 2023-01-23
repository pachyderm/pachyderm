package log

import (
	"context"
	"net"
	"net/http"

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
			Info(ctx, "incoming http request", zap.Stringer("url", r.URL), zap.String("method", r.Method), zap.String("host", r.Host), id)
			ctx = ChildLogger(ctx, "", WithFields(id))
			ctx = metadata.NewOutgoingContext(ctx, metadata.MD{"x-request-id": requestID})
			r = r.WithContext(ctx)
			orig.ServeHTTP(w, r)
		})
	}
}
