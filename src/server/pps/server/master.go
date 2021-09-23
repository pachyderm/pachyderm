package server

import (
	"context"
	"path"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/dlock"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/middleware/auth"
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
		"InvalidImageName": true,
		"ErrImagePull":     true,
		"Unschedulable":    true,
	}

	zero     int32 // used to turn down RCs in scaleDownWorkersForPipeline
	falseVal bool  // used to delete RCs in deletePipelineResources and restartPipeline()
)

type eventType int

const (
	writeEv eventType = iota
	deleteEv
)

type pipelineEvent struct {
	eventType
	pipeline  string
	timestamp time.Time
}

type stepError struct {
	error
	retry        bool
	failPipeline bool
}

func newRetriableError(err error, message string) error {
	return stepError{
		error:        errors.Wrap(err, message),
		retry:        true,
		failPipeline: true,
	}
}

func (s stepError) Unwrap() error {
	return s.error
}

type ppsMaster struct {
	// The PPS APIServer that owns this struct
	a *apiServer

	// masterCtx is a context that is cancelled if
	// the current pps master loses its master status
	masterCtx context.Context

	// fields for monitorPipeline goros, monitorCrashingPipeline goros, etc.
	monitorCancelsMu       sync.Mutex
	monitorCancels         map[string]func() // protected by monitorCancelsMu
	crashingMonitorCancels map[string]func() // also protected by monitorCancelsMu

	// fields for the pollPipelines, pollPipelinePods, and watchPipelines goros
	pollPipelinesMu sync.Mutex
	pollCancel      func() // protected by pollPipelinesMu
	pollPodsCancel  func() // protected by pollPipelinesMu
	watchCancel     func() // protected by pollPipelinesMu

	opsInProcessMu sync.Mutex
	opsInProcess   map[string]*pipelineOp

	startMonitorPipelineCh chan *pps.PipelineInfo
	stopMonitorPipelineCh  chan *pps.PipelineInfo

	startMonitorCrashingPipelineCh chan *pps.PipelineInfo
	stopMonitorCrashingPipelineCh  chan *pps.PipelineInfo

	// channel through which pipeline events are passed
	eventCh chan *pipelineEvent
}

// The master process is responsible for creating/deleting workers as
// pipelines are created/removed.
func (a *apiServer) master() {
	m := &ppsMaster{
		a:                              a,
		monitorCancels:                 make(map[string]func()),
		crashingMonitorCancels:         make(map[string]func()),
		opsInProcess:                   make(map[string]*pipelineOp),
		startMonitorPipelineCh:         make(chan *pps.PipelineInfo),
		stopMonitorPipelineCh:          make(chan *pps.PipelineInfo),
		startMonitorCrashingPipelineCh: make(chan *pps.PipelineInfo),
		stopMonitorCrashingPipelineCh:  make(chan *pps.PipelineInfo),
	}

	masterLock := dlock.NewDLock(a.env.EtcdClient, path.Join(a.etcdPrefix, masterLockPath))
	backoff.RetryNotify(func() error {
		ctx, cancel := context.WithCancel(context.Background())
		// set internal auth for basic operations
		ctx = auth.AsInternalUser(ctx, "pps-master")
		defer cancel()
		ctx, err := masterLock.Lock(ctx)
		if err != nil {
			return err
		}
		defer masterLock.Unlock(ctx)

		log.Infof("PPS master: launching master process")
		m.masterCtx = ctx
		go m.processMonitorUpdates()
		m.run()
		return errors.Wrapf(ctx.Err(), "ppsMaster.Run() exited unexpectedly")
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		log.Errorf("PPS master: error running the master process: %v; retrying in %v", err, d)
		return nil
	})
	panic("internal error: PPS master has somehow exited. Restarting pod...")
}

func (a *apiServer) setPipelineCrashing(ctx context.Context, specCommit *pfs.Commit, reason string) error {
	return a.setPipelineState(ctx, specCommit, pps.PipelineState_PIPELINE_CRASHING, reason)
}

func (m *ppsMaster) processMonitorUpdates() {
	for {
		select {
		case pi := <-m.startMonitorPipelineCh:
			m.startMonitor(pi)
		case pi := <-m.stopMonitorPipelineCh:
			m.cancelMonitor(pi.Pipeline.Name)
		case pi := <-m.startMonitorCrashingPipelineCh:
			m.startCrashingMonitor(pi)
		case pi := <-m.stopMonitorCrashingPipelineCh:
			m.cancelCrashingMonitor(pi.Pipeline.Name)
		}
	}
}

func (m *ppsMaster) run() {
	// close m.eventCh after all cancels have returned and therefore all pollers
	// (which are what write to m.eventCh) have exited
	m.eventCh = make(chan *pipelineEvent, 1)
	defer close(m.eventCh)
	defer m.cancelAllMonitorsAndCrashingMonitors()
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
				m.opsInProcessMu.Lock()
				defer m.opsInProcessMu.Unlock()

				var ot opType
				switch e.eventType {
				case writeEv:
					ot = writeOp // raises flag in pipelineOp to run again whenever it finishes (doesn't block)
				case deleteEv:
					ot = deleteOp
				}

				if pipelineOp, ok := m.opsInProcess[e.pipeline]; ok {
					pipelineOp.Bump(ot)
				} else {
					// Initialize op ctx (cancelled at the end of pipelineOp.Start(), to avoid leaking
					// resources), whereas masterClient is passed by the
					// PPS master and used in case a monitor needs to be spawned for 'pipeline',
					// whose lifetime is tied to the master rather than this op.
					opCtx, opCancel := context.WithCancel(m.masterCtx)
					pipelineOp, err := m.newPipelineOp(opCtx, opCancel, e.pipeline)
					if err != nil {
						log.Errorf("PPS master: error creating a pipelineOp for pipeline '%s': %v", e.pipeline, err)
					}
					m.opsInProcess[e.pipeline] = pipelineOp
					go pipelineOp.Start(ot, e.timestamp)
				}
			}(e)
		case <-m.masterCtx.Done():
			break eventLoop
		}
	}
}

// setPipelineState is a PPS-master-specific helper that wraps
// ppsutil.SetPipelineState in a trace
func (a *apiServer) setPipelineState(ctx context.Context, specCommit *pfs.Commit, state pps.PipelineState, reason string) (retErr error) {
	span, ctx := tracing.AddSpanToAnyExisting(ctx,
		"/pps.Master/SetPipelineState", "pipeline", specCommit.Branch.Repo.Name, "new-state", state)
	defer func() {
		tracing.TagAnySpan(span, "err", retErr)
		tracing.FinishAnySpan(span)
	}()
	return ppsutil.SetPipelineState(ctx, a.env.DB, a.pipelines,
		specCommit, nil, state, reason)
}

// transitionPipelineState is similar to setPipelineState, except that it sets
// 'from' and logs a different trace
func (a *apiServer) transitionPipelineState(ctx context.Context, specCommit *pfs.Commit, from []pps.PipelineState, to pps.PipelineState, reason string) (retErr error) {
	span, ctx := tracing.AddSpanToAnyExisting(ctx,
		"/pps.Master/TransitionPipelineState", "pipeline", specCommit.Branch.Repo.Name,
		"from-state", from, "to-state", to)
	defer func() {
		tracing.TagAnySpan(span, "err", retErr)
		tracing.FinishAnySpan(span)
	}()
	return ppsutil.SetPipelineState(ctx, a.env.DB, a.pipelines,
		specCommit, from, to, reason)
}
