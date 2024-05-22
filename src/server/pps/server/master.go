package server

import (
	"context"
	"fmt"
	"path"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dlock"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	middleware_auth "github.com/pachyderm/pachyderm/v2/src/internal/middleware/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

const (
	masterLockPath = "_master_lock"
	maxErrCount    = 3 // gives all retried operations ~4.5s total to finish
)

var (
	failures = map[string]bool{
		// from k8s.io/kubernetes/pkg/kubelet/images/types.go
		"ErrImagePull":            true,
		"InvalidImageName":        true,
		"ImagePullBackOff":        true,
		v1.PodReasonUnschedulable: true,
		// from k8s.io/kubernetes/pkg/kubelet/kuberuntime/kuberuntime_container.go
		"CreateContainerConfigError": true,
		"CreateContainerError":       true,
	}

	zero     int32 // used to turn down RCs in scaleDownWorkersForPipeline
	falseVal bool  // used to delete RCs in deletePipelineResources and restartPipeline()
)

type pipelineKey string

func toKey(p *pps.Pipeline) pipelineKey {
	return pipelineKey(fmt.Sprintf("%s/%s", p.GetProject().GetName(), p.GetName()))
}

func fromKey(k pipelineKey) (*pps.Pipeline, error) {
	parts := strings.Split(string(k), "/")
	// TODO: this will need to change when hierarchical projects are
	// enabled; then keys with longer parts will be permissible
	if len(parts) != 2 {
		return nil, errors.Errorf("invalid pipeline key %s", k)
	}
	return newPipeline(parts[0], parts[1]), nil
}

func newPipeline(projectName, pipelineName string) *pps.Pipeline {
	return &pps.Pipeline{
		Project: &pfs.Project{Name: projectName},
		Name:    pipelineName,
	}
}

type pipelineEvent struct {
	pipeline  *pps.Pipeline
	timestamp time.Time
}

type stepError struct {
	error
	retry        bool
	failPipeline bool
}

func newRetriableError(err error, message string) error {
	retry, failPipeline := true, true
	if errors.Is(err, context.Canceled) {
		retry = false
		failPipeline = false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		retry = true
		failPipeline = false
	}
	return stepError{
		error:        errors.Wrap(err, message),
		retry:        retry,
		failPipeline: failPipeline,
	}
}

func (s stepError) Unwrap() error {
	return s.error
}

type ppsMaster struct {
	env Env
	// for checking worker status during crashing monitor
	etcdPrefix string
	// masterCtx is a context that is cancelled if
	// the current pps master loses its master status
	masterCtx context.Context
	// fields for the pollPipelines, pollPipelinePods, and watchPipelines goros
	pollPipelinesMu sync.Mutex
	pollCancel      func() // protected by pollPipelinesMu
	pollPodsCancel  func() // protected by pollPipelinesMu
	watchCancel     func() // protected by pollPipelinesMu
	pcMgr           *pcManager
	kd              InfraDriver
	sd              PipelineStateDriver
	// channel through which pipeline events are passed
	eventCh         chan *pipelineEvent
	scaleUpInterval time.Duration
	crashingBackoff time.Duration
}

func newMaster(ctx context.Context, env Env, etcdPrefix string, kd InfraDriver, sd PipelineStateDriver) *ppsMaster {
	return &ppsMaster{
		masterCtx:       ctx,
		env:             env,
		etcdPrefix:      etcdPrefix,
		pcMgr:           newPcManager(),
		kd:              kd,
		sd:              sd,
		scaleUpInterval: time.Second * 30,
		crashingBackoff: time.Second * 15,
	}
}

// The master process is responsible for creating/deleting workers as
// pipelines are created/removed.
func (a *apiServer) master(ctx context.Context) {
	masterLock := dlock.NewDLock(a.env.EtcdClient, path.Join(a.etcdPrefix, masterLockPath))
	backoff.RetryUntilCancel(ctx, func() error { //nolint:errcheck
		ctx, cancel := pctx.WithCancel(pctx.Child(ctx, "master", pctx.WithServerID()))
		// set internal auth for basic operations
		ctx = middleware_auth.AsInternalUser(ctx, "pps-master")
		defer cancel()
		ctx, err := masterLock.Lock(ctx)
		if err != nil {
			return errors.EnsureStack(err)
		}
		defer masterLock.Unlock(ctx) //nolint:errcheck
		log.Info(ctx, "PPS master: launching master process")
		kd := newKubeDriver(a.env.KubeClient, a.env.Config)
		sd := newPipelineStateDriver(a.env.DB, a.pipelines, a.txnEnv, a.env.PFSServer)
		go gcDetUsers(ctx, a.getDetConfig(), time.Minute, sd, a.env.KubeClient.CoreV1().Secrets(a.namespace))
		m := newMaster(ctx, a.env, a.etcdPrefix, kd, sd)
		m.run()
		return errors.Wrapf(context.Cause(ctx), "ppsMaster.Run() exited unexpectedly")
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		log.Error(ctx, "PPS master: error running the master process; retrying",
			zap.Error(err), zap.Duration("retryIn", d))
		return nil
	})
}

func (m *ppsMaster) setPipelineCrashing(ctx context.Context, specCommit *pfs.Commit, reason string) error {
	if err := m.sd.SetState(ctx, specCommit, pps.PipelineState_PIPELINE_CRASHING, reason); err != nil {
		return errors.Wrapf(err, "failed to set pipeline to crashing state")
	}
	return nil
}

// run() ingests pipeline events from `m.eventCh` that are generated by watching the DB and
// from polling k8s. It then distributes the work to pipelineController goroutines that refresh
// the worker(s) state for each pipeline.
//
// Notes:
// - Since each event `e` instructs the master to set a particular pipeline `p` to its most recently
// declared state, and because pipelines don't share k8s resources, we can run a single goroutine
// for each `p` to increase throughput across pipelines.
//
// - In the case where many events are queued for a given pipeline, we can skip to the
// most recent event in the next `step` the pipelineController executes. This is done using `pipelineController.Bump()`,
// i.e. when an active pipelineController completes execution, it will re-execute with the
// most recently declared state if it has been bumped.
func (m *ppsMaster) run() {
	// close m.eventCh after all cancels have returned and therefore all pollers
	// (which are what write to m.eventCh) have exited
	m.eventCh = make(chan *pipelineEvent, 1)
	defer close(m.eventCh)
	defer m.cancelPCs()
	// start pollers in the background--cancel functions ensure poll/monitor
	// goroutines all definitely stop (either because cancelXYZ returns or because
	// the binary panics)
	m.startPipelinePoller()
	defer m.cancelPipelinePoller()
	m.startPipelinePodsPoller()
	defer m.cancelPipelinePodsPoller()
	m.startPipelineWatcher()
	defer m.cancelPipelineWatcher()

eventLoop:
	for {
		select {
		case e := <-m.eventCh:
			func(e *pipelineEvent) {
				log.Debug(m.masterCtx, "pipelineEvent", zap.Stringer("pipeline", e.pipeline))
				m.pcMgr.Lock()
				defer m.pcMgr.Unlock()
				key := toKey(e.pipeline)
				if pc, ok := m.pcMgr.pcs[key]; ok {
					pc.Bump(e.timestamp) // raises flag in pipelineController to run again whenever it finishes
				} else {
					// pc's ctx is cancelled in pipelineController.tryFinish(), to avoid leaking resources
					pcCtx, pcCancel := pctx.WithCancel(m.masterCtx)
					pc = m.newPipelineController(pcCtx, pcCancel, e.pipeline)
					m.pcMgr.pcs[key] = pc
					go pc.Start(e.timestamp)
				}
			}(e)
		case <-m.masterCtx.Done():
			break eventLoop
		}
	}
}

// setPipelineState is a PPS-master-specific helper that wraps
// ppsutil.SetPipelineState in a trace
func setPipelineState(ctx context.Context, db *pachsql.DB, pipelines collection.PostgresCollection, specCommit *pfs.Commit, state pps.PipelineState, reason string) (retErr error) {
	log.Debug(ctx, "set pipeline state", zap.String("pipeline", specCommit.GetBranch().GetRepo().GetName()), zap.Stringer("state", state))
	span, ctx := tracing.AddSpanToAnyExisting(ctx,
		"/pps.Master/SetPipelineState", "project", specCommit.Repo.Project.GetName(), "pipeline", specCommit.Repo.Name, "new-state", state)
	defer func() {
		tracing.TagAnySpan(span, "err", retErr)
		tracing.FinishAnySpan(span)
	}()
	return ppsutil.SetPipelineState(ctx, db, pipelines,
		specCommit, nil, state, reason)
}

func (m *ppsMaster) cancelPCs() {
	m.pcMgr.Lock()
	defer m.pcMgr.Unlock()
	for _, pc := range m.pcMgr.pcs {
		pc.cancel()
	}
}
