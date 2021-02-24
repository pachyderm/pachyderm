package collection

import (
	"fmt"
	"strings"
)

// ErrNotFound indicates that a key was not found when it was expected to
// exist.
type ErrNotFound struct {
	Type string
	Key  string
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("%s %s not found", strings.TrimPrefix(e.Type, DefaultPrefix), e.Key)
}

// IsErrNotFound determines if an error is an ErrNotFound error
func IsErrNotFound(e error) bool {
	_, ok := e.(ErrNotFound)
	return ok
}

// ErrExists indicates that a key was found to exist when it was expected not
// to.
type ErrExists struct {
	Type string
	Key  string
}

func (e ErrExists) Error() string {
	return fmt.Sprintf("%s %s already exists", strings.TrimPrefix(e.Type, DefaultPrefix), e.Key)
}

// IsErrExists determines if an error is an ErrExists error
func IsErrExists(e error) bool {
	_, ok := e.(ErrExists)
	return ok
}

// ErrMalformedValue indicates that a value was malformed, such as when it was
// supposed to be parseable as an int but wasn't.
type ErrMalformedValue struct {
	Type string
	Key  string
	Val  string
}

func (e ErrMalformedValue) Error() string {
	return fmt.Sprintf("malformed value at %s/%s: %s", strings.TrimPrefix(e.Type, DefaultPrefix), e.Key, e.Val)
}

// IsErrMalformedValue determines if an error is an ErrMalformedValue error
func IsErrMalformedValue(e error) bool {
	_, ok := e.(ErrMalformedValue)
	return ok
}
