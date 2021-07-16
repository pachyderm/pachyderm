package grpcutil

import (
	"strings"

	"google.golang.org/grpc/status"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// We use the 'causer' interface to match errors that were wrapped using our
// errors package.  We aren't concerned with errors that were wrapped in other
// ways (like fmt.Errorf("%w")).
type causer interface {
	Cause() error
}

// ScrubGRPC removes GRPC error code information from 'err' if it came from
// GRPC (and returns it unchanged otherwise).
func ScrubGRPC(err error) error {
	if err == nil {
		return nil
	}

	// For wrapped errors, we need to check if each level is a GRPC error,
	// construct a replacement error, then rewrap it all the way back up the
	// chain.
	errs := []error{err}
	for err != nil {
		cause, ok := err.(causer)
		if !ok {
			break
		}
		err = cause.Cause()
		errs = append(errs, err)
	}

	// Map any GRPC errors to a stripped error
	changed := false
	for i := range errs {
		if s, ok := status.FromError(errs[i]); ok {
			errs[i] = errors.New(s.Message())
			changed = true
		}
	}

	if !changed {
		return errs[0]
	}

	// At least one of the errors was a GRPC error, so walk backwards through the
	// list and reconstruct the original error. This only really works for the
	// specific formatting done by the errors package for wrapped errors, but that
	// is probably ok since Errorf("%w") doesn't produce a structured wrapped
	// error, just a flat one with a new string.
	err = errs[len(errs)-1]
	for i := len(errs) - 2; i >= 0; i-- {
		str := errs[i].Error()
		innerErrStr := ": " + errs[i+1].Error()
		msg := strings.ReplaceAll(str, ": "+innerErrStr, "")
		err = errors.Wrapf(err, msg)
	}

	return err
}
