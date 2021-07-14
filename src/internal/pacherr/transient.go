package pacherr

import (
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TransientError struct {
	Err     error
	MinWait time.Duration
}

func WrapTransient(err error, minWait time.Duration) error {
	return &TransientError{
		Err:     err,
		MinWait: minWait,
	}
}

func (e *TransientError) Error() string {
	return e.Err.Error()
}

func (e *TransientError) Unwrap() error {
	return e.Err
}

func (e *TransientError) GRPCStatus() *status.Status {
	// TODO: not sure if codes.Unavailable is appropriate here
	return status.New(codes.Unavailable, e.Error())
}
