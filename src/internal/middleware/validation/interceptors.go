package validation

import (
	"context"
	"fmt"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type validatable interface {
	ValidateAll() error
}

type branchNillable interface {
	NilBranch()
}

func UnaryServerInterceptor(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if r, ok := req.(validatable); ok {
		if err := r.ValidateAll(); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "validate request: %v", err)
		}
	} else if _, ok := req.(*emptypb.Empty); ok {
		// Expected.
	} else if _, ok := req.(*grpc_health_v1.HealthCheckRequest); ok {
		// Expected.
	} else {
		log.DPanic(ctx, "no validation routine on request message", zap.String("type", fmt.Sprintf("%T", req)))
	}

	resp, err := handler(ctx, req)
	if err != nil {
		return nil, err
	}
	if b, ok := resp.(branchNillable); ok {
		b.NilBranch()
	}
	return resp, nil
}

var _ grpc.UnaryServerInterceptor = UnaryServerInterceptor

type streamWrapper struct {
	grpc.ServerStream
	IsServerStream bool
}

var _ grpc.ServerStream = new(streamWrapper)

// RecvMsg implements grpc.ServerStream.
func (w *streamWrapper) RecvMsg(m any) error {
	if err := w.ServerStream.RecvMsg(m); err != nil {
		return err //nolint:wrapcheck
	}
	if r, ok := m.(validatable); ok {
		if err := r.ValidateAll(); err != nil {
			return status.Errorf(codes.InvalidArgument, "validate recv: %v", err)
		}
	} else {
		log.DPanic(w.Context(), "no validation routine on request message", zap.String("type", fmt.Sprintf("%T", m)))
	}
	return nil
}

func (w *streamWrapper) SendMsg(m any) error {
	// google grpc library wraps client stream and puts server stream. We don't want our implementation to apply to client streams.
	if w.IsServerStream {
		if b, ok := m.(branchNillable); ok {
			b.NilBranch()
		}
	}
	return errors.EnsureStack(w.ServerStream.SendMsg(m))
}

func StreamServerInterceptor(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return handler(srv, &streamWrapper{ServerStream: stream, IsServerStream: info.IsServerStream})
}

var _ grpc.StreamServerInterceptor = StreamServerInterceptor
