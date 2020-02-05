package worker

import (
	"context"
	"path"
	"time"

	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/dlock"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/work"
	"github.com/pachyderm/pachyderm/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/src/server/worker/logs"
	"github.com/pachyderm/pachyderm/src/server/worker/pipeline/service"
	"github.com/pachyderm/pachyderm/src/server/worker/pipeline/spout"
	"github.com/pachyderm/pachyderm/src/server/worker/pipeline/transform"
	"github.com/pachyderm/pachyderm/src/server/worker/server"
)

const (
	masterLockPath = "_master_worker_lock"
)

type Worker struct {
	pachClient      *client.APIClient // TODO: does this need to be separate from driver.PachClient()?
	etcdClient      *etcd.Client
	etcdPrefix      string
	driver          driver.Driver // Provides common functions used by worker code
	namespace       string        // The namespace in which pachyderm is deployed
	transformWorker *transform.Worker
}

func NewWorker(pachClient *client.APIClient, etcdClient *etcd.Client, etcdPrefix string, pipelineInfo *pps.PipelineInfo, workerName string, namespace string, hashtreePath string) (*Worker, error) {
	stats.InitPrometheus()

	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	kubeClient, err := kube.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

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
		if pipelineInfo.Transform.Cmd == nil {
			if len(image.Config.Entrypoint) == 0 {
				ppsutil.FailPipeline(ctx, etcdClient, d.Pipelines(),
					pipelineInfo.Pipeline.Name,
					"nothing to run: no transform.cmd and no entrypoint")
			}
			pipelineInfo.Transform.Cmd = image.Config.Entrypoint
		}
	}

	d, err := driver.NewDriver(
		pipelineInfo,
		oldPachClient,
		driver.NewKubeWrapper(kubeClient),
		etcdClient,
		etcdPrefix,
		filepath.Join(hashtreePath, uuid.NewWithoutDashes()),
		client.PPSInputPrefix,
	)
	if err != nil {
		return nil, err
	}

	worker := &Worker{
		pachClient: oldPachClient,
		etcdClient: etcdClient,
		etcdPrefix: etcdPrefix,
		driver:     d,
		namespace:  namespace,
		apiServer:  server.NewAPIServer(d, workerName),
	}

	go worker.master()
	go worker.worker()
	return worker, nil
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

func (w *Worker) worker() {
	ctx := w.driver.PachClient().Ctx()
	logger := logs.NewStatlessLogger(w.driver.PipelineInfo())

	backoff.RetryUntilCancel(ctx, func() error {
		return w.driver.NewTaskWorker().Run(
			w.driver.PachClient().Ctx(),
			func(ctx context.Context, task *work.Task, subtask *work.Task) error {
				return transform.Worker(w.driver, logger, task, subtask)
			},
		)
	}, backoff.NewConstantBackOff(200*time.Millisecond), func(err error, d time.Duration) error {
		logger.Logf("worker failed: %v; retrying in %v", err, d)
		return nil
	})
}

func (w *Worker) master() {
	logger := logs.NewMasterLogger(w.pipelineInfo)

	masterLock := dlock.NewDLock(w.etcdClient, path.Join(w.etcdPrefix, masterLockPath, w.pipelineInfo.Pipeline.Name, w.pipelineInfo.Salt))
	b := backoff.NewInfiniteBackOff()
	// Setting a high backoff so that when this master fails, the other
	// workers are more likely to become the master.
	// Also, we've observed race conditions where StopPipeline would cause
	// a master to restart before it's deleted.  PPS would then get confused
	// by the restart and create the workers again, because the restart would
	// bring the pipeline state from PAUSED to RUNNING.  By setting a high
	// retry interval, the master would be deleted before it gets a chance
	// to restart.
	b.InitialInterval = 10 * time.Second
	backoff.RetryNotify(func() error {
		// We use pachClient.Ctx here because it contains auth information.
		ctx, cancel := context.WithCancel(w.pachClient.Ctx())
		defer cancel() // make sure that everything this loop might spawn gets cleaned up
		ctx, err := masterLock.Lock(ctx)
		if err != nil {
			return err
		}
		defer masterLock.Unlock(ctx)

		// Create a new driver that uses a new cancelable pachClient
		return runSpawner(w.driver.WithCtx(ctx), logger)
	}, b, func(err error, d time.Duration) error {
		if auth.IsErrNotAuthorized(err) {
			logger.Logf("failing %q due to auth rejection", w.pipelineInfo.Pipeline.Name)
			return ppsutil.FailPipeline(w.pachClient.Ctx(), w.etcdClient, w.driver.Pipelines(),
				w.pipelineInfo.Pipeline.Name, "worker master could not access output "+
					"repo to watch for new commits")
		}
		logger.Logf("master: error running the master process, retrying in %v: %v", d, err)
		return nil
	})
}

type spawnerFunc func(driver.Driver, logs.TaggedLogger) error

// Run runs the spawner for a given pipeline.  This switches between several
// underlying functions based on the configuration in pipelineInfo (e.g. if
// it is a service, a spout, or a transform pipeline).
func runSpawner(driver driver.Driver, logger logs.TaggedLogger) error {
	pipelineType, runFn := func() (string, spawnerFunc) {
		switch {
		case driver.PipelineInfo().Service != nil:
			return "service", service.Run
		case driver.PipelineInfo().Spout != nil:
			return "spout", spout.Run
		default:
			return "transform", transform.Run
		}
	}()

	logger.Logf("Launching %v spawner process", pipelineType)
	err := runFn(driver, logger)
	if err != nil {
		logger.Logf("error running the %v spawner process: %v", pipelineType, err)
	}
	return err
}
