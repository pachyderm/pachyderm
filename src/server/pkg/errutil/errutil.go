package errutil

import (
	"strings"
)

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
