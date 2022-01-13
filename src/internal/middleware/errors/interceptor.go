package errors

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor translates errors for unary RPCs
func UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	res, err := handler(ctx, req)
	return res, unwrapGRPC(err)
}

// StreamServerInterceptor translates errors for streaming RPCs
func StreamServerInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	err := handler(srv, stream)
	return unwrapGRPC(err)
}

func unwrapGRPC(err error) error {
	cursor := err
	for cursor != nil {
		cursor = errors.Unwrap(cursor)
		if res, ok := status.FromError(cursor); ok && res != nil {
			return cursor //nolint:wrapcheck
		}
	}
	return err
}
