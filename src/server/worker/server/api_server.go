package server

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/driver"
	workerapi "github.com/pachyderm/pachyderm/v2/src/worker"
	"google.golang.org/protobuf/types/known/emptypb"
)

// WorkerInterface is an interface for getting or canceling the
// currently-running task in the worker process.
type WorkerInterface interface {
	GetStatus() (*pps.WorkerStatus, error)
	Cancel(jobID string, datumFilter []string) bool
	NextDatum(context.Context, error) ([]string, error)
}

// APIServer implements the worker API
type APIServer struct {
	workerapi.UnimplementedWorkerServer

	driver          driver.Driver
	workerInterface WorkerInterface
	workerName      string // The k8s pod name of this worker
}

// NewAPIServer creates an APIServer for a given pipeline
func NewAPIServer(driver driver.Driver, workerInterface WorkerInterface, workerName string) *APIServer {
	return &APIServer{
		driver:          driver,
		workerInterface: workerInterface,
		workerName:      workerName,
	}
}

// Status returns the status of the current worker task.
func (a *APIServer) Status(ctx context.Context, _ *emptypb.Empty) (*pps.WorkerStatus, error) {
	status, err := a.workerInterface.GetStatus()
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	status.WorkerId = a.workerName
	return status, nil
}

// Cancel cancels the currently running datum
func (a *APIServer) Cancel(ctx context.Context, request *workerapi.CancelRequest) (*workerapi.CancelResponse, error) {
	success := a.workerInterface.Cancel(request.JobId, request.DataFilters)
	return &workerapi.CancelResponse{Success: success}, nil
}

func (a *APIServer) NextDatum(ctx context.Context, request *workerapi.NextDatumRequest) (*workerapi.NextDatumResponse, error) {
	var err error
	if request.Error != "" {
		err = errors.New(request.Error)
	}
	env, err := a.workerInterface.NextDatum(ctx, err)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &workerapi.NextDatumResponse{Env: env}, nil
}
