package server

import (
	"context"
	"errors"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
)

type WorkerPool interface {
	// Process blocks until a worker accepts the given datum set.
	Submit(datumSet []*pfs.FileInfo) error
	// Wait blocks until all datum sets that have been submitted up
	// until this point have been processed.
	Wait() error
}

func (a *apiServer) workerPool(ctx context.Context, jobInfo *pps.JobInfo) (WorkerPool, error) {
	return nil, errors.New("TODO")
}
