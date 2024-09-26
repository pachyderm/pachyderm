// Package errors provides error-intercepting gRPC middleware.
package errors

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"

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
	// strip off the initial stack trace layer, in case this is just a once-wrapped auth error
	cursor := err
	for cursor != nil {
		if _, ok := cursor.(errors.StackTracer); ok {
			cursor = errors.Unwrap(cursor)
		} else {
			break
		}
	}
	if cursor == nil && err != nil {
		// this should never happen, we unwrapped an error consisting only of stack traces
		// reset to the original error, we have no idea what this is
		cursor = err
	}
	if _, ok := status.FromError(cursor); ok {
		// handles nil as well
		return cursor
	}
	code := codes.Unknown
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
	}
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
