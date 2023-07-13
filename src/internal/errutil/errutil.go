package errutil

import (
	"net"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pacherr"
)

var ErrBreak = pacherr.ErrBreak

// IsAlreadyExistError returns true if err is due to trying to create a
// resource that already exists. It uses simple string matching, it's not
// terribly smart.
func IsAlreadyExistError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "already exists")
}

// IsNotFoundError returns true if err is due to a resource not being found. It
// uses simple string matching, it's not terribly smart.
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "not found")
}

// IsWriteToOutputBranchError returns true if the err is due to an attempt to
// write to an output repo/branch
func IsWriteToOutputBranchError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "cannot start a commit on an output branch")
}

// IsNotADirectoryError returns true if the err is due to an attempt to put a
// file on path that has a non-directory parent. These errors come from the
// hashtree package; while it provides an error code, we can't check against
// that because we'd then have to import hashtree, and hashtree imports
// errutil, leading to circular imports.
func IsNotADirectoryError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "but it's not a directory")
}

// IsInvalidPathError returns true if the err is due to an invalid path
func IsInvalidPathError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "only printable ASCII characters allowed") ||
		strings.Contains(err.Error(), "not allowed in path")
}

// IsNetRetryable returns true if the error is a temporary network error.
func IsNetRetryable(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Temporary() //nolint:staticcheck
}

// IsDatabseDisconnect returns true if the error represents a database disconnect
func IsDatabaseDisconnect(err error) bool {
	return strings.Contains(err.Error(), "broken pipe") || strings.Contains(err.Error(), "unexpected EOF")
}
