package errutil

import (
	"fmt"
	"strings"
)

var (
	// ErrBreak is an error used to break out of call back based iteration,
	// should be swallowed by iteration functions and treated as successful
	// iteration.
	ErrBreak = fmt.Errorf("BREAK")

	// ErrUnusedPeek is an error used to indicate that a callback function did
	// not use a parameter that was peeked off of something. Make this comment better.
	ErrUnusedPeek = fmt.Errorf("UNUSED_PEEK")

	ErrStop = fmt.Errorf("STOP")
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
