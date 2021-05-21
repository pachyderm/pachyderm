package server

import (
	"github.com/gogo/protobuf/types"
	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/driver"
)

// WorkerInterface is an interface for getting or canceling the
// currently-running task in the worker process.
type WorkerInterface interface {
	GetStatus() (*pps.WorkerStatus, error)
	Cancel(pipelineJobID string, datumFilter []string) bool
}

// APIServer implements the worker API
type APIServer struct {
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
func (a *APIServer) Status(ctx context.Context, _ *types.Empty) (*pps.WorkerStatus, error) {
	status, err := a.workerInterface.GetStatus()
	if err != nil {
		return nil, err
	}
	status.WorkerID = a.workerName
	return status, nil
}

// Cancel cancels the currently running datum
func (a *APIServer) Cancel(ctx context.Context, request *CancelRequest) (*CancelResponse, error) {
	success := a.workerInterface.Cancel(request.PipelineJobID, request.DataFilters)
	return &CancelResponse{Success: success}, nil
}
