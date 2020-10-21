package server

import (
	"github.com/gogo/protobuf/types"
	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	"github.com/pachyderm/pachyderm/src/server/worker/driver"
)

// WorkerInterface is an interface for getting or canceling the
// currently-running task in the worker process.
type WorkerInterface interface {
	GetStatus() (*pps.WorkerStatus, error)
	Cancel(jobID string, datumFilter []string) bool
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
	success := a.workerInterface.Cancel(request.JobID, request.DataFilters)
	return &CancelResponse{Success: success}, nil
}

// GetChunk returns the merged datum hashtrees of a particular chunk (if available)
func (a *APIServer) GetChunk(request *GetChunkRequest, server Worker_GetChunkServer) error {
	filter := hashtree.NewFilter(a.driver.NumShards(), request.Shard)
	if request.Stats {
		cache := a.driver.ChunkStatsCaches().GetCache(request.JobID)
		if cache != nil && cache.Has(request.ChunkID) {
			return cache.Get(request.ChunkID, grpcutil.NewStreamingBytesWriter(server), filter)
		}
	} else {
		cache := a.driver.ChunkCaches().GetCache(request.JobID)
		if cache != nil && cache.Has(request.ChunkID) {
			return cache.Get(request.ChunkID, grpcutil.NewStreamingBytesWriter(server), filter)
		}
	}
	return errors.New("hashtree chunk not found")
}

// Exec executes an adhoc command
func (a *APIServer) Exec(request *pps.ExecRequest, server Worker_ExecServer) (retErr error) {
	if err := server.Send(&pps.ExecResponse{Stdout: []byte("Hello World!")}); err != nil {
		return err
	}
	return nil
}
