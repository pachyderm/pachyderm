package errutil

import (
	"strings"
)

// IsAlreadyExistError returns true if an error is due to trying to create a
// resource that already exists. It uses simple string matching, it's not
// terrible smart.
func IsAlreadyExistError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "already exists")
}
