package collection

import (
	"fmt"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrNotFound indicates that a key was not found when it was expected to
// exist.
type ErrNotFound struct {
	Type          string
	Key           string
	customMessage string
}

func (err ErrNotFound) Is(other error) bool {
	_, ok := other.(ErrNotFound)
	return ok
}

func (err ErrNotFound) Error() string {
	if err.customMessage != "" {
		return err.customMessage
	}
	return fmt.Sprintf("%s %s not found", strings.TrimPrefix(err.Type, DefaultPrefix), err.Key)
}

func (e ErrNotFound) GRPCStatus() *status.Status {
	return status.New(codes.NotFound, e.Error())
}

// IsErrNotFound determines if an error is an ErrNotFound error
func IsErrNotFound(err error) bool {
	return errors.Is(err, ErrNotFound{})
}

// ErrExists indicates that a key was found to exist when it was expected not
// to.
type ErrExists struct {
	Type          string
	Key           string
	customMessage string
}

func (err ErrExists) Is(other error) bool {
	_, ok := other.(ErrExists)
	return ok
}

func (err ErrExists) Error() string {
	if err.customMessage != "" {
		return err.customMessage
	}
	return fmt.Sprintf("%s %s already exists", strings.TrimPrefix(err.Type, DefaultPrefix), err.Key)
}

func (e ErrExists) GRPCStatus() *status.Status {
	return status.New(codes.AlreadyExists, e.Error())
}

// IsErrExists determines if an error is an ErrExists error
func IsErrExists(err error) bool {
	return errors.Is(err, ErrExists{})
}

// ErrNotUnique indicates that an indexed query expected to have exactly one
// result but had more than one result.
type ErrNotUnique struct {
	Type  string
	Index string
	Value string
}

func (err ErrNotUnique) Is(other error) bool {
	_, ok := other.(ErrNotUnique)
	return ok
}

func (err ErrNotUnique) Error() string {
	return fmt.Sprintf("more than one matching row found in %s on index %s with value %s", err.Type, err.Index, err.Value)
}

// IsErrNotUnique determines if an error is an ErrNotUnique error
func IsErrNotUnique(err error) bool {
	return errors.Is(err, ErrNotUnique{})
}
