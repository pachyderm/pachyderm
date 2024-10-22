package snapshotdb

import (
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SnapshotNotFoundError struct {
	ID int64
}

func (err *SnapshotNotFoundError) GRPCStatus() *status.Status {
	return status.New(codes.NotFound, err.Error())
}

func (err *SnapshotNotFoundError) Error() string {
	return fmt.Sprintf("snapshot not found (id=%d)", err.ID)
}
