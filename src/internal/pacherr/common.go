package pacherr

import (
	"fmt"
	"os"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

type ErrNotExist struct {
	Collection string
	ID         string
}

func NewNotExist(collection, id string) error {
	return &ErrNotExist{
		Collection: collection,
		ID:         id,
	}
}

func (e *ErrNotExist) Error() string {
	return fmt.Sprintf("%s does not contain item: (%s)", e.Collection, e.ID)
}

func (e *ErrNotExist) GRPCStatus() *status.Status {
	return status.New(codes.NotFound, e.Error())
}

func IsNotExist(err error) bool {
	target := &ErrNotExist{}
	return errors.As(err, target) || os.IsNotExist(err)
}

type ErrExists struct {
	Collection string
	ID         string
}

func NewExists(collection, id string) error {
	return &ErrExists{
		Collection: collection,
		ID:         id,
	}
}

func (e *ErrExists) Error() string {
	return fmt.Sprintf("%s already contains an item: (%s)", e.Collection, e.ID)
}

func (e *ErrExists) GRPCStatus() *status.Status {
	return status.New(codes.AlreadyExists, e.Error())
}

func IsExists(err error) bool {
	target := &ErrExists{}
	return errors.As(err, target)
}

var (
	// ErrBreak is an error used to break out of call back based iteration,
	// should be swallowed by iteration functions and treated as successful
	// iteration.
	ErrBreak = errors.Errorf("BREAK")
)
