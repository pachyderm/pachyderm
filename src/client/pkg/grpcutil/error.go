package grpcutil

import (
	"errors"

	"google.golang.org/grpc/status"
)

// StripGRPCCode removes GRPC error code information from 'err' if it came from
// GRPC (and returns it unchanged otherwise)
func StripGRPCCode(err error) error {
	if err == nil {
		return nil
	}
	if s, ok := status.FromError(err); err != nil && ok {
		return errors.New(s.Message())
	}
	return err
}
