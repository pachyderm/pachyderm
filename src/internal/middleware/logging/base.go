package logging

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"google.golang.org/grpc"
)

type BaseContextInterceptor struct {
	base context.Context
}

var (
	_ grpc.UnaryServerInterceptor  = new(BaseContextInterceptor).UnaryServerInterceptor
	_ grpc.StreamServerInterceptor = new(BaseContextInterceptor).StreamServerInterceptor
)

func NewBaseContextInterceptor(ctx context.Context) *BaseContextInterceptor {
	return &BaseContextInterceptor{base: ctx}
}

func (bi *BaseContextInterceptor) UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, retErr error) {
	ctx = log.CombineLogger(ctx, bi.base)
	return handler(ctx, req)
}

type contextStream struct {
	ctx context.Context
	grpc.ServerStream
}

func (s *contextStream) Context() context.Context {
	return s.ctx
}

func (bi *BaseContextInterceptor) StreamServerInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (retErr error) {
	wrapper := &contextStream{
		ServerStream: stream,
		ctx:          log.CombineLogger(stream.Context(), bi.base),
	}
	return handler(srv, wrapper)
}
