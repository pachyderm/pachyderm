// Package recovery implements a GRPC server interceptor that recovers from panicking RPCs.
package recovery

import (
	"context"
	"runtime"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	PanicCode = codes.Aborted
	panicLen  = 32768
)

// UnaryServerInterceptor implements grpc.UnaryServerInterceptor.  It recovers from a panicking RPC
// handler.
func UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, retErr error) {
	defer func() {
		if err := recover(); err != nil {
			stack := make([]byte, panicLen)
			n := runtime.Stack(stack, false)
			stack = stack[:n]
			retErr = status.Errorf(PanicCode, "panic: %v\n%s", err, stack)
		}
	}()
	return handler(ctx, req)
}

var _ grpc.UnaryServerInterceptor = UnaryServerInterceptor

// StreamServerInterceptor implements grpc.StreamServerInterceptor.  It recovers from a panicking
// RPC handler.
func StreamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (retErr error) {
	defer func() {
		if err := recover(); err != nil {
			stack := make([]byte, panicLen)
			n := runtime.Stack(stack, false)
			stack = stack[:n]
			retErr = status.Errorf(PanicCode, "panic: %v\n%s", err, stack)
		}
	}()
	return handler(srv, ss)
}

var _ grpc.StreamServerInterceptor = StreamServerInterceptor
