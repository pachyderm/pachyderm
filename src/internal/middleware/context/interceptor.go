package context

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/middleware/util"
	"google.golang.org/grpc"
)

const defaultMethodTimeout time.Duration = 300 * time.Second

var customTimeoutMethods = map[string]time.Duration{}

func setTimeout(fullMethod string, ctx context.Context) context.Context {
	if timeout, ok := customTimeoutMethods[fullMethod]; ok {
		if timeout == 0 {
			return ctx
		}
		newCtx, _ := context.WithTimeout(ctx, timeout)
		return newCtx
	}
	newCtx, _ := context.WithTimeout(ctx, defaultMethodTimeout)
	return newCtx
}

func UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	ctx = setTimeout(info.FullMethod, ctx)
	return handler(ctx, req)
}

func StreamServerInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := setTimeout(info.FullMethod, stream.Context())
	streamProxy := util.ServerStreamWrapper{Stream: stream, Ctx: ctx}
	return handler(srv, streamProxy)
}
