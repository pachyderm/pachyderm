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

var (
	errDatumRecovered = errors.New("the datum errored, and the error was handled successfully")
	statsTagSuffix    = "_stats"
)

// APIServer implements the worker API
type APIServer struct {
	pachClient *client.APIClient
	etcdClient *etcd.Client
	etcdPrefix string

	// Provides common functions used by worker code
	driver driver.Driver

	// Information needed to process input data and upload output
	pipelineInfo *pps.PipelineInfo

	// The k8s pod name of this worker
	workerName string

	statusMu sync.Mutex

	// The currently running job ID
	jobID string
	// The currently running data
	data []*common.Input
	// The time we started the currently running
	started time.Time
	// Func to cancel the currently running datum
	cancel func()
	// Stats about the execution of the job
	stats *pps.ProcessStats
	// queueSize is the number of items enqueued
	queueSize int64

	// The namespace in which pachyderm is deployed
	namespace string

	// Only one datum can be running at a time because they need to be
	// accessing /pfs, runMu enforces this
	runMu sync.Mutex

	// hashtreeStorage is the where we store on disk hashtrees
	hashtreeStorage string

	// numShards is the number of filesystem shards for the output of this pipeline
	numShards int64
	// claimedShard communicates the context for the shard that was claimed
	claimedShard chan context.Context
	// shardCtx is the context for the shard this worker has claimed
	shardCtx context.Context
	// shard is the shard this worker has claimed
	shard int64
	// chunkCache caches chunk hashtrees during a job and can merge them (chunkStatsCache applies to stats)
	chunkCache, chunkStatsCache *hashtree.MergeCache
	// datumCache caches datum hashtrees during a job and can merge them (datumStatsCache applies to stats)
	datumCache, datumStatsCache *hashtree.MergeCache
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
		pipelineInfo:    pipelineInfo,
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
		return a.chunkStatsCache.Get(request.Id, grpcutil.NewStreamingBytesWriter(server), filter)
	}
	return a.chunkCache.Get(request.Id, grpcutil.NewStreamingBytesWriter(server), filter)
}

func (a *APIServer) datum() []*pps.InputFile {
	var result []*pps.InputFile
	for _, datum := range a.data {
		result = append(result, &pps.InputFile{
			Path: datum.FileInfo.File.Path,
			Hash: datum.FileInfo.Hash,
		})
	}
	return result
}

// TODO: move this to common
func userCodeEnv(jobID string, outputCommitID string, inputDir string, data []*common.Input) []string {
	result := os.Environ()
	for _, input := range data {
		result = append(result, fmt.Sprintf("%s=%s", input.Name, filepath.Join(inputDir, input.Name, input.FileInfo.File.Path)))
		result = append(result, fmt.Sprintf("%s_COMMIT=%s", input.Name, input.FileInfo.File.Commit.ID))
	}
	result = append(result, fmt.Sprintf("%s=%s", client.JobIDEnv, jobID))
	result = append(result, fmt.Sprintf("%s=%s", client.OutputCommitIDEnv, outputCommitID))
	return result
}
