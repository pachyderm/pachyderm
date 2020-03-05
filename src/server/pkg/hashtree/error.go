package hashtree

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"

	go_errors "github.com/pkg/errors"
)

// hashTreeError is a custom error type, returned by HashTree's methods
type hashTreeError struct {
	code       ErrCode
	s          string
	stackTrace go_errors.StackTrace
}

func (e *hashTreeError) Error() string {
	return e.s
}

func (e *hashTreeError) StackTrace() go_errors.StackTrace {
	return e.stackTrace
}

// Code returns the "error code"  of 'err' if it was returned by one of the
// HashTree methods, or "Unknown" if 'err' was emitted by some other function
// (error codes are defined in interface.go)
func Code(err error) ErrCode {
	if err == nil {
		return OK
	}

	hte, ok := err.(*hashTreeError)
	if !ok {
		return Unknown
	}
	return hte.code
}

// errorf is analogous to errors.Errorf, but generates hashTreeErrors instead of
// errorStrings.
func errorf(c ErrCode, fmtStr string, args ...interface{}) error {
	return &hashTreeError{
		code:       c,
		s:          fmt.Sprintf(fmtStr, args...),
		stackTrace: errors.Callers(),
	}
}
