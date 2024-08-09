package pjsdb

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

var (
	ErrParentNotFound          = errors.New("parent job not found")
	ErrJobContextAlreadyExists = errors.New("job context already exists")
)

type InvalidFilesetError struct {
	ID, Reason string
}

func (err *InvalidFilesetError) GRPCStatus() *status.Status {
	return status.New(codes.NotFound, err.Error())
}

func (err *InvalidFilesetError) Error() string {
	return fmt.Sprintf("invalid fileset '%s', reason: %s", err.ID, err.Reason)
}

type JobNotFoundError struct {
	ID          JobID
	ProgramHash string
}

func (err *JobNotFoundError) GRPCStatus() *status.Status {
	return status.New(codes.NotFound, err.Error())
}

func (err *JobNotFoundError) Error() string {
	return fmt.Sprintf("job not found (id=%d, program_hash=%v)", err.ID, err.ProgramHash)
}
