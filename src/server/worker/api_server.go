package worker

import (
	"errors"
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

// APIServer implements the worker API
type APIServer struct {
	pachClient *client.APIClient
	etcdClient *etcd.Client
	etcdPrefix string

	// Provides common functions used by worker code
	driver driver.Driver

	// The k8s pod name of this worker
	workerName string

	statusMu sync.Mutex

	// Stats about the execution of the job
	stats *pps.ProcessStats
	// queueSize is the number of items enqueued
	queueSize int64

	// The namespace in which pachyderm is deployed
	namespace string
	// clients are the worker clients (used for the shuffle step by mergers)
	clients map[string]Client
}

// NewAPIServer creates an APIServer for a given pipeline
func NewAPIServer(pachClient *client.APIClient, etcdClient *etcd.Client, etcdPrefix string, pipelineInfo *pps.PipelineInfo, workerName string, namespace string, hashtreeStorage string) (*APIServer, error) {
	stats.InitPrometheus()

	span, ctx := extended.AddPipelineSpanToAnyTrace(pachClient.Ctx(),
		etcdClient, pipelineInfo.Pipeline.Name, "/worker/Start")
	oldPachClient := pachClient // don't use tracing in apiServer.pachClient
	if span != nil {
		pachClient = pachClient.WithCtx(ctx)
	}
	defer tracing.FinishAnySpan(span)

	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	kubeClient, err := kube.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	d, err := driver.NewDriver(pipelineInfo, oldPachClient, driver.NewKubeWrapper(kubeClient), etcdClient, etcdPrefix)
	if err != nil {
		return nil, err
	}

	server := &APIServer{
		pachClient:      oldPachClient,
		etcdClient:      etcdClient,
		etcdPrefix:      etcdPrefix,
		driver:          d,
		workerName:      workerName,
		namespace:       namespace,
		hashtreeStorage: hashtreeStorage,
		claimedShard:    make(chan context.Context, 1),
		shard:           noShard,
		clients:         make(map[string]Client),
	}
	logger := logs.NewStatlessLogger(pipelineInfo)

	numShards, err := ppsutil.GetExpectedNumHashtrees(pipelineInfo.HashtreeSpec)
	if err != nil {
		logger.Logf("error getting number of shards, default to 1 shard: %v", err)
		numShards = 1
	}
	server.numShards = numShards
	root := filepath.Join(hashtreeStorage, uuid.NewWithoutDashes())
	if err := os.MkdirAll(filepath.Join(root, "chunk", "stats"), 0777); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(root, "datum", "stats"), 0777); err != nil {
		return nil, err
	}
	server.chunkCache = hashtree.NewMergeCache(filepath.Join(root, "chunk"))
	server.chunkStatsCache = hashtree.NewMergeCache(filepath.Join(root, "chunk", "stats"))
	server.datumCache = hashtree.NewMergeCache(filepath.Join(root, "datum"))
	server.datumStatsCache = hashtree.NewMergeCache(filepath.Join(root, "datum", "stats"))
	var noDocker bool
	if _, err := os.Stat("/var/run/docker.sock"); err != nil {
		noDocker = true
	}
	if pipelineInfo.Transform.Image != "" && !noDocker {
		docker, err := docker.NewClientFromEnv()
		if err != nil {
			return nil, err
		}
		image, err := docker.InspectImage(pipelineInfo.Transform.Image)
		if err != nil {
			return nil, fmt.Errorf("error inspecting image %s: %+v", pipelineInfo.Transform.Image, err)
		}
		if pipelineInfo.Transform.User == "" {
			pipelineInfo.Transform.User = image.Config.User
		}
		if pipelineInfo.Transform.WorkingDir == "" {
			pipelineInfo.Transform.WorkingDir = image.Config.WorkingDir
		}
		if server.pipelineInfo.Transform.Cmd == nil {
			if len(image.Config.Entrypoint) == 0 {
				ppsutil.FailPipeline(ctx, etcdClient, d.Pipelines(),
					pipelineInfo.Pipeline.Name,
					"nothing to run: no transform.cmd and no entrypoint")
			}
			server.pipelineInfo.Transform.Cmd = image.Config.Entrypoint
		}
	}
	go server.master()
	go server.worker()
	return server, nil
}

// Status returns the status of the current worker.
func (a *APIServer) Status(ctx context.Context, _ *types.Empty) (*pps.WorkerStatus, error) {
	a.statusMu.Lock()
	defer a.statusMu.Unlock()
	started, err := types.TimestampProto(a.started)
	if err != nil {
		return nil, err
	}
	result := &pps.WorkerStatus{
		JobID:     a.jobID,
		WorkerID:  a.workerName,
		Started:   started,
		Data:      nil, // TODO: implement this
		QueueSize: atomic.LoadInt64(&a.queueSize),
	}
	return result, nil
}

// Cancel cancels the currently running datum
func (a *APIServer) Cancel(ctx context.Context, request *CancelRequest) (*CancelResponse, error) {
	a.statusMu.Lock()
	defer a.statusMu.Unlock()
	if request.JobID != a.jobID {
		return &CancelResponse{Success: false}, nil
	}
	if !common.MatchDatum(request.DataFilters, a.datum()) {
		return &CancelResponse{Success: false}, nil
	}
	a.cancel()
	// clear the status since we're no longer processing this datum
	a.jobID = ""
	a.data = nil
	a.started = time.Time{}
	a.cancel = nil
	return &CancelResponse{Success: true}, nil
}

// GetChunk returns the merged datum hashtrees of a particular chunk (if available)
func (a *APIServer) GetChunk(request *GetChunkRequest, server Worker_GetChunkServer) error {
	filter := hashtree.NewFilter(a.numShards, request.Shard)
	if request.Stats {
		return driver.ChunkStatsCache().Get(request.JobID, request.Tag, grpcutil.NewStreamingBytesWriter(server), filter)
	}
	return driver.ChunkCache().Get(request.JobID, request.Tag, grpcutil.NewStreamingBytesWriter(server), filter)
}
