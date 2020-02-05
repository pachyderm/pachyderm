package service

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/types"
	"golang.org/x/net/context"
	kube "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pkg/tracing"
	"github.com/pachyderm/pachyderm/src/client/pkg/tracing/extended"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/worker/common"
	"github.com/pachyderm/pachyderm/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/src/server/worker/logs"
	"github.com/pachyderm/pachyderm/src/server/worker/stats"
)

const (
	shardTTL          = 30
	noShard           = int64(-1)
	parentTreeBufSize = 50 * (1 << (10 * 2))
)

type ctxKey int

const (
	shardKey ctxKey = iota
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

// Status returns the status of the current worker.
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
	filter := hashtree.NewFilter(a.numShards, request.Shard)
	if request.Stats {
		if driver.ChunkStatsCache().Has(request.JobID, request.Tag) {
			return driver.ChunkStatsCache().Get(request.JobID, request.Tag, grpcutil.NewStreamingBytesWriter(server), filter)
		}
	}
	if driver.ChunkCache().Has(request.JobID, request.Tag) {
		return driver.ChunkCache().Get(request.JobID, request.Tag, grpcutil.NewStreamingBytesWriter(server), filter)
	}
	return fmt.Errorf("hashtree chunk not found")
}
