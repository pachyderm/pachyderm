package errors

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor translates errors for unary RPCs
func UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	res, err := handler(ctx, req)
	return res, errorForGRPC(err)
}

// StreamServerInterceptor translates errors for streaming RPCs
func StreamServerInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	err := handler(srv, stream)
	return errorForGRPC(err)
}

func errorForGRPC(err error) error {
	if _, ok := status.FromError(err); ok {
		// handles nil as well
		return err //nolint:wrapcheck
	}
	code := codes.Unknown
	cursor := err
	for cursor != nil {
		if res, ok := status.FromError(cursor); ok {
			code = res.Code()
			break
		}
		cursor = errors.Unwrap(cursor)
	}
	return &gRPCStatusError{
		code: code,
		err:  err,
	} //nolint:wrapcheck
}

type gRPCStatusError struct {
	err  error
	code codes.Code
}

func (e *gRPCStatusError) GRPCStatus() *status.Status {
	return status.New(e.code, e.Error())
}

func (e *gRPCStatusError) Error() string {
	return e.err.Error()
}

func (e *gRPCStatusError) Unwrap() error {
	return e.err
}
