package hashtree

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
)

// hashTreeError is a custom error type, returned by HashTree's methods
type hashTreeError struct {
	code       ErrCode
	s          string
	stackTrace errors.StackTrace
}

func (e *hashTreeError) Error() string {
	return e.s
}

func (e *hashTreeError) StackTrace() errors.StackTrace {
	return e.stackTrace
}

// Code returns the "error code"  of 'err' if it was returned by one of the
// HashTree methods, or "Unknown" if 'err' was emitted by some other function
// (error codes are defined in interface.go)
func Code(err error) ErrCode {
	if err == nil {
		return OK
	}

	var hte *hashTreeError
	if errors.As(err, &hte) {
		return hte.code
	}
	return Unknown
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
