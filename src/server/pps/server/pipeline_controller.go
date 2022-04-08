package server

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	middleware_auth "github.com/pachyderm/pachyderm/v2/src/internal/middleware/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing/extended"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs/pretty"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/v2/src/version"

	opentracing "github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

func max(is ...int) int {
	if len(is) == 0 {
		return 0
	}
	max := is[0]
	for _, i := range is {
		if i > max {
			max = i
		}
	}
	return max
}

type pcManager struct {
	sync.Mutex
	pcs map[string]*pipelineController
}

func newPcManager() *pcManager {
	return &pcManager{
		pcs: make(map[string]*pipelineController),
	}
}

type sideEffect int32

const (
	sideEffect_CREATE_RESOURCES       sideEffect = 0
	sideEffect_DELETE_RESOURCES       sideEffect = 1
	sideEffect_FINISH_OUTPUT_COMMITS  sideEffect = 2
	sideEffect_RESTART                sideEffect = 3
	sideEffect_SCALE_UP               sideEffect = 4
	sideEffect_SCALE_DOWN             sideEffect = 5
	sideEffect_START_PIPELINE_MONITOR sideEffect = 6
	sideEffect_STOP_PIPELINE_MONITOR  sideEffect = 7
	sideEffect_START_CRASHING_MONITOR sideEffect = 8
	sideEffect_STOP_CRASHING_MONITOR  sideEffect = 9
)

type sideEffectHandler func(ctx context.Context, pc *pipelineController, pi *pps.PipelineInfo, rc *v1.ReplicationController) error

var sideEffectHandlers map[sideEffect]sideEffectHandler = map[sideEffect]sideEffectHandler{
	sideEffect_CREATE_RESOURCES: func(ctx context.Context, pc *pipelineController, pi *pps.PipelineInfo, _ *v1.ReplicationController) error {
		return pc.kd.CreatePipelineResources(ctx, pi)
	},
	sideEffect_DELETE_RESOURCES: func(_ context.Context, pc *pipelineController, _ *pps.PipelineInfo, _ *v1.ReplicationController) error {
		if err := pc.deletePipelineResources(); err != nil {
			// retry, but the pipeline has already failed
			return stepError{
				error: errors.Wrap(err, "error deleting resources for failing pipeline"),
				retry: true,
			}
		}
		return nil
	},
	sideEffect_FINISH_OUTPUT_COMMITS: func(ctx context.Context, pc *pipelineController, pi *pps.PipelineInfo, _ *v1.ReplicationController) error {
		return pc.finishPipelineOutputCommits(ctx, pi)
	},
	sideEffect_RESTART: func(ctx context.Context, pc *pipelineController, pi *pps.PipelineInfo, rc *v1.ReplicationController) error {
		return pc.restartPipeline(ctx, pi, rc, "missing RC")
	},
	sideEffect_SCALE_UP: func(ctx context.Context, pc *pipelineController, pi *pps.PipelineInfo, rc *v1.ReplicationController) error {
		return pc.scaleUpPipeline(ctx, pi, rc)
	},
	sideEffect_SCALE_DOWN: func(ctx context.Context, pc *pipelineController, pi *pps.PipelineInfo, rc *v1.ReplicationController) error {
		return pc.scaleDownPipeline(ctx, pi, rc)
	},
	sideEffect_START_PIPELINE_MONITOR: func(_ context.Context, pc *pipelineController, pi *pps.PipelineInfo, _ *v1.ReplicationController) error {
		pc.startPipelineMonitor(pi)
		return nil
	},
	sideEffect_STOP_PIPELINE_MONITOR: func(_ context.Context, pc *pipelineController, _ *pps.PipelineInfo, _ *v1.ReplicationController) error {
		pc.stopPipelineMonitor()
		return nil
	},
	sideEffect_START_CRASHING_MONITOR: func(_ context.Context, pc *pipelineController, pi *pps.PipelineInfo, _ *v1.ReplicationController) error {
		pc.startCrashingPipelineMonitor(pi)
		return nil
	},
	sideEffect_STOP_CRASHING_MONITOR: func(_ context.Context, pc *pipelineController, _ *pps.PipelineInfo, _ *v1.ReplicationController) error {
		pc.stopCrashingPipelineMonitor()
		return nil
	},
}

// pipelineController contains all of the relevent current state for a pipeline. It's
// used by step() to take any necessary actions
type pipelineController struct {
	// a pachyderm client wrapping this operation's context (child of the PPS
	// master's context, and cancelled at the end of Start())
	ctx        context.Context
	cancel     context.CancelFunc
	pipeline   string
	namespace  string
	env        Env
	txEnv      *transactionenv.TransactionEnv
	pipelines  collection.PostgresCollection
	etcdPrefix string

	monitorCancel         func()
	crashingMonitorCancel func()

	bumpChan chan time.Time
	pcMgr    *pcManager
	kd       *kubeDriver
}

var (
	errRCNotFound   = errors.New("RC not found")
	errUnexpectedRC = errors.New("unexpected RC")
	errTooManyRCs   = errors.New("multiple RCs found for pipeline")
	errStaleRC      = errors.New("RC doesn't match pipeline version (likely stale)")
)

func (m *ppsMaster) newPipelineController(ctx context.Context, cancel context.CancelFunc, pipeline string) *pipelineController {
	pc := &pipelineController{
		ctx:    ctx,
		cancel: cancel,
		// pipeline name is recorded separately in the case we are running a delete operation and pipelineInfo isn't available in the DB
		pipeline:   pipeline,
		namespace:  m.a.namespace,
		env:        m.a.env,
		txEnv:      m.a.txnEnv,
		pipelines:  m.a.pipelines,
		etcdPrefix: m.a.etcdPrefix,
		kd:         m.kd,
		bumpChan:   make(chan time.Time, 1),
		pcMgr:      m.pcMgr,
	}
	return pc
}

// Bump signals the pipelineController goro to either process the  latest state
// of the pipeline now, or once its done with its current processing iteration.
//
// This function is expected to be called while holding the pcMgr lock.
// This helps guard the critical section when the goroutine is cleaned up
// as part of a pipeline deletion event in pipelineController.tryFinish().
//
// Note: since pc.bumpChan is a buffered channel of length 1, Bump() will
// add to the channel if it's empty, and do nothing otherwise.
func (pc *pipelineController) Bump(ts time.Time) {
	select {
	case pc.bumpChan <- ts:
	default:
	}
}

// Start calls step for the given pipeline in a backoff loop.
// When it encounters a stepError, returned by most pipeline controller helper
// functions, it uses the fields to decide whether to retry the step or,
// if the retries have been exhausted, fail the pipeline.
//
// Other errors are simply logged and ignored, assuming that some future polling
// of the pipeline will succeed.
func (pc *pipelineController) Start(timestamp time.Time) {
	pc.Bump(timestamp)
	for {
		select {
		case <-pc.ctx.Done():
			return
		case ts := <-pc.bumpChan:
			// pc.step returns true if the pipeline was deleted, and the controller can try to shutdown
			if isDelete, err := pc.step(ts); err != nil {
				log.Errorf("PPS master: failed to run step for pipeline %s: %v", pc.pipeline, err)
			} else if isDelete {
				pc.tryFinish()
			}
		}
	}
}

// finishes the pc if it isn't bumped
func (pc *pipelineController) tryFinish() {
	pc.pcMgr.Lock()
	defer pc.pcMgr.Unlock()
	select {
	case ts := <-pc.bumpChan:
		pc.Bump(ts)
	default:
		pc.cancel()
		delete(pc.pcMgr.pcs, pc.pipeline)
	}
}

// step fetches 'pipelineInfo', the latest pipeline pointer in the pipeline collection, and
// 1. retrieves its full pipeline spec and RC into the 'Details' field
// 2. makes whatever changes are needed to bring the RC in line with the (new) spec
// 3. updates 'pipelineInfo', if needed, to reflect the action it just took
//
// returns true if the pipeline is deleted, and the pipelineController can try to shutdown
func (pc *pipelineController) step(timestamp time.Time) (isDelete bool, retErr error) {
	log.Debugf("PPS master: processing event for %q", pc.pipeline)
	// Handle tracing
	span, _ := extended.AddSpanToAnyPipelineTrace(pc.ctx, pc.env.EtcdClient, pc.pipeline, "/pps.Master/ProcessPipelineUpdate")
	if !timestamp.IsZero() {
		tracing.TagAnySpan(span, "update-time", timestamp)
	} else {
		tracing.TagAnySpan(span, "pollpipelines-or-autoscaling-event", "true")
	}
	defer tracing.FinishAnySpan(span, "err", retErr)
	// derive the latest pipelineInfo with a corresponding auth'd context
	pi, ctx, err := pc.fetchState()
	if err != nil {
		// if we fail to create a new step, there was an error querying the pipeline info, and there's nothing we can do
		log.Errorf("PPS master: failed to set up step data to handle event for pipeline '%s': %v", pc.pipeline, errors.Wrapf(err, "failing pipeline %q", pc.pipeline))
		return false, err
	} else if pi == nil {
		// interpret the event as a delete operation
		if err := pc.deletePipelineResources(); err != nil {
			log.Errorf("PPS master: error deleting pipelineController resources for pipeline '%s': %v", pc.pipeline, err)
			return true, errors.Wrapf(err, "error deleting pipelineController resources for pipeline '%s'", pc.pipeline)
		}
		return true, nil
	}
	var stepErr stepError
	// TODO(msteffen) should this fail the pipeline? (currently getRC will restart
	// the pipeline indefinitely)
	rc, restart, err := pc.getRC(ctx, pi)
	if restart {
		return false, pc.restartPipeline(ctx, pi, rc, "could not get RC.")
	}
	if err != nil && !errors.Is(err, errRCNotFound) {
		return false, err
	}
	targetState, sideEffects, err := evaluate(pi, rc)
	if err != nil {
		return false, err
	}
	errCount := 0
	err = backoff.RetryNotify(func() error {
		// Create/Modify/Delete pipeline resources as needed per new state
		return pc.apply(ctx, pi, rc, targetState, sideEffects)
	}, backoff.NewExponentialBackOff(), func(err error, d time.Duration) error {
		errCount++
		if errors.As(err, &stepErr) {
			if stepErr.retry && errCount < maxErrCount {
				log.Errorf("PPS master: error updating resources for pipeline %q: %v; retrying in %v",
					pc.pipeline, err, d)
				return nil
			}
		}
		return errors.Wrapf(err, "could not update resource for pipeline %q", pc.pipeline)
	})
	if errors.As(err, &stepErr) && stepErr.failPipeline {
		failError := pc.setPipelineState(ctx,
			pi.SpecCommit,
			pps.PipelineState_PIPELINE_FAILURE,
			fmt.Sprintf("could not update resources after %d attempts: %v", errCount, err),
		)
		if failError != nil {
			log.Errorf("PPS master: error failing pipeline '%s' after step error (%v): %v", pc.pipeline, stepErr, failError)
			return false, errors.Wrapf(failError, "error failing pipeline %q  after step error (%v)", pc.pipeline, stepErr)
		}
	}
	return false, err
}

// returns PipelineInfo corresponding to the latest pipeline state, a context loaded with the pipeline's auth info, and error
// NOTE: returns nil, nil, nil if the step is found to be a delete operation
func (pc *pipelineController) fetchState() (*pps.PipelineInfo, context.Context, error) {
	// query pipelineInfo
	var pi *pps.PipelineInfo
	var err error
	if pi, err = pc.tryLoadLatestPipelineInfo(); err != nil && collection.IsErrNotFound(err) {
		// if the pipeline info is not found, interpret the operation as a delete
		return nil, nil, nil
	} else if err != nil {
		return nil, nil, err
	}
	tracing.TagAnySpan(pc.ctx,
		"current-state", pi.State.String(),
		"spec-commit", pretty.CompactPrintCommitSafe(pi.SpecCommit))
	// add pipeline auth
	// the pipelineController's context is authorized as pps master, but we want to switch to the pipeline itself
	// first clear the cached WhoAmI result from the context
	pachClient := pc.env.GetPachClient(middleware_auth.ClearWhoAmI(pc.ctx))
	pachClient.SetAuthToken(pi.AuthToken)
	return pi, pachClient.Ctx(), nil
}

func evaluate(pi *pps.PipelineInfo, rc *v1.ReplicationController) (pps.PipelineState, []sideEffect, error) {
	if pi.State == pps.PipelineState_PIPELINE_FAILURE {
		return pps.PipelineState_PIPELINE_FAILURE,
			[]sideEffect{
				sideEffect_FINISH_OUTPUT_COMMITS,
				sideEffect_DELETE_RESOURCES,
			}, nil
	}
	if pi.State == pps.PipelineState_PIPELINE_STARTING || pi.State == pps.PipelineState_PIPELINE_RESTARTING {
		sideEffects := make([]sideEffect, 0)
		if rc == nil {
			sideEffects = append(sideEffects, sideEffect_CREATE_RESOURCES)
		}
		if pi.Stopped {
			return pps.PipelineState_PIPELINE_PAUSED, sideEffects, nil
		}
		sideEffects = append(sideEffects, sideEffect_STOP_CRASHING_MONITOR)
		if pi.Details.Autoscaling && pi.State == pps.PipelineState_PIPELINE_STARTING {
			return pps.PipelineState_PIPELINE_STANDBY, sideEffects, nil
		}
		return pps.PipelineState_PIPELINE_RUNNING, sideEffects, nil
	}
	if rc == nil {
		// may happen if an external system deletes the RC
		return pi.State, []sideEffect{sideEffect_RESTART}, nil
	}
	if pi.State == pps.PipelineState_PIPELINE_PAUSED {
		if !pi.Stopped {
			// StartPipeline has been called (so spec commit is updated), but new spec
			// commit hasn't been propagated to PipelineInfo or RC yet
			if pi.Details.Autoscaling {
				return pps.PipelineState_PIPELINE_STANDBY, nil, nil
			}
			return pps.PipelineState_PIPELINE_RUNNING, nil, nil
		}
		return pi.State, []sideEffect{
			// don't want cron commits or STANDBY state changes while pipeline is
			// stopped
			sideEffect_STOP_PIPELINE_MONITOR,
			sideEffect_STOP_CRASHING_MONITOR,
			// default: scale down if pause/standby hasn't propagated to collection yet
			sideEffect_SCALE_DOWN,
		}, nil
	}
	if pi.Stopped {
		return pps.PipelineState_PIPELINE_PAUSED, nil, nil
	}
	switch pi.State {
	case pps.PipelineState_PIPELINE_RUNNING:
		return pi.State, []sideEffect{
			sideEffect_STOP_CRASHING_MONITOR,
			sideEffect_START_PIPELINE_MONITOR,
			// default: scale up if pipeline start hasn't propagated to the collection yet
			// Note: mostly this should do nothing, as this runs several times per job
			sideEffect_SCALE_UP,
		}, nil
	case pps.PipelineState_PIPELINE_STANDBY:
		return pi.State, []sideEffect{
			sideEffect_STOP_CRASHING_MONITOR,
			// Make sure pipelineMonitor is running to pull it out of standby
			sideEffect_START_PIPELINE_MONITOR,
			// default: scale down if standby hasn't propagated to kube RC yet
			sideEffect_SCALE_DOWN,
		}, nil
	case pps.PipelineState_PIPELINE_CRASHING:
		return pi.State, []sideEffect{
			// start a monitor to poll k8s and update us when it goes into a running state
			sideEffect_START_CRASHING_MONITOR,
			sideEffect_START_PIPELINE_MONITOR,
			// Surprisingly, scaleUpPipeline() is necessary, in case a pipelines is
			// quickly transitioned to CRASHING after coming out of STANDBY. Because the
			// pipeline controller reads the current state of the pipeline after each
			// event (to avoid getting backlogged), it might never actually see the
			// pipeline in RUNNING. However, if the RC is never scaled up, the pipeline
			// can never come out of CRASHING, so do it here in case it never happened.
			//
			// In general, CRASHING is actually almost identical to RUNNING (except for
			// the monitorCrashing goro)
			sideEffect_SCALE_UP,
		}, nil
	}
	return 0, nil, errors.New("could not evaluate pipeline transition")
}

func (pc *pipelineController) apply(ctx context.Context, pi *pps.PipelineInfo, rc *v1.ReplicationController, target pps.PipelineState, sideEffects []sideEffect) error {
	for _, se := range sideEffects {
		if err := sideEffectHandlers[se](ctx, pc, pi, rc); err != nil {
			return err
		}
	}
	if target != pi.State {
		return pc.setPipelineState(ctx, pi.SpecCommit, target, "")
	}
	return nil
}

func (pc *pipelineController) tryLoadLatestPipelineInfo() (*pps.PipelineInfo, error) {
	pi := &pps.PipelineInfo{}
	errCnt := 0
	err := backoff.RetryNotify(func() error {
		return pc.loadLatestPipelineInfo(pi)
	}, backoff.NewExponentialBackOff(), func(err error, d time.Duration) error {
		errCnt++
		// Don't put the pipeline in a failing state if we're in the middle
		// of activating auth, retry in a bit
		if (auth.IsErrNotAuthorized(err) || auth.IsErrNotSignedIn(err)) && errCnt <= maxErrCount {
			log.Warnf("PPS master: could not retrieve pipelineInfo for pipeline %q: %v; retrying in %v",
				pc.pipeline, err, d)
			return nil
		}
		return stepError{
			error: errors.Wrapf(err, "could not load pipelineInfo for pipeline %q", pc.pipeline),
			retry: false,
		}
	})
	return pi, err
}

func (pc *pipelineController) loadLatestPipelineInfo(message *pps.PipelineInfo) error {
	specCommit, err := ppsutil.FindPipelineSpecCommit(pc.ctx, pc.env.PFSServer, *pc.txEnv, pc.pipeline)
	if err != nil {
		return errors.Wrapf(err, "could not find spec commit for pipeline %q", pc.pipeline)
	}
	if err := pc.pipelines.ReadOnly(pc.ctx).Get(specCommit, message); err != nil {
		return errors.Wrapf(err, "could not retrieve pipeline info for %q", pc.pipeline)
	}
	return nil
}

// rcIsFresh returns a boolean indicating whether rc has the right labels
// corresponding to pipelineInfo. If this returns false, it likely means the
// current RC is using e.g. an old spec commit or something.
func rcIsFresh(pi *pps.PipelineInfo, rc *v1.ReplicationController) bool {
	if rc == nil {
		log.Errorf("PPS master: RC for %q is nil", pi.Pipeline.Name)
		return false
	}
	expectedName := ppsutil.PipelineRcName(pi.Pipeline.Name, pi.Version)
	// establish current RC properties
	rcName := rc.ObjectMeta.Name
	rcPachVersion := rc.ObjectMeta.Annotations[pachVersionAnnotation]
	rcAuthTokenHash := rc.ObjectMeta.Annotations[hashedAuthTokenAnnotation]
	rcPipelineVersion := rc.ObjectMeta.Annotations[pipelineVersionAnnotation]
	rcSpecCommit := rc.ObjectMeta.Annotations[pipelineSpecCommitAnnotation]
	switch {
	case rcPipelineVersion != strconv.FormatUint(pi.Version, 10):
		log.Infof("PPS master: pipeline version in %q looks stale %s != %d",
			pi.Pipeline.Name, rcPipelineVersion, pi.Version)
		return false
	case rcSpecCommit != pi.SpecCommit.ID:
		log.Infof("PPS master: pipeline spec commit in %q looks stale %s != %s",
			pi.Pipeline.Name, rcSpecCommit, pi.SpecCommit.ID)
		return false
	case rcPachVersion != version.PrettyVersion():
		log.Infof("PPS master: %q is using stale pachd v%s != current v%s",
			pi.Pipeline.Name, rcPachVersion, version.PrettyVersion())
		return false
	case rcName != expectedName:
		log.Infof("PPS master: %q has an unexpected (likely stale) name %q != %q",
			pi.Pipeline.Name, rcName, expectedName)
	case rcAuthTokenHash != hashAuthToken(pi.AuthToken):
		log.Infof("PPS master: auth token in %q is stale %s != %s",
			pi.Pipeline.Name, rcAuthTokenHash, hashAuthToken(pi.AuthToken))
		return false
	}
	return true
}

// setPipelineState set's pc's state in the collection to 'state'. This will trigger a
// collection watch event and cause step() to eventually run again.
func (pc *pipelineController) setPipelineState(ctx context.Context, specCommit *pfs.Commit, state pps.PipelineState, reason string) error {
	if err := setPipelineState(ctx, pc.env.DB, pc.pipelines, specCommit, state, reason); err != nil {
		// don't bother failing if we can't set the state
		return stepError{
			error: errors.Wrapf(err, "could not set pipeline state to %v"+
				"(you may need to restart pachd to un-stick the pipeline)", state),
			retry: true,
		}
	}
	return nil
}

// startPipelineMonitor spawns a monitorPipeline() goro for this pipeline (if
// one doesn't exist already), which manages standby and cron inputs, and
// updates the the pipeline state.
// Note: this is called by every run through step(), so must be idempotent
func (pc *pipelineController) startPipelineMonitor(pi *pps.PipelineInfo) {
	if pc.monitorCancel == nil {
		pc.monitorCancel = pc.startMonitor(pc.ctx, pi)
	}
}

func (pc *pipelineController) startCrashingPipelineMonitor(pi *pps.PipelineInfo) {
	if pc.crashingMonitorCancel == nil {
		pc.crashingMonitorCancel = pc.startCrashingMonitor(pc.ctx, pi)
	}
}

func (pc *pipelineController) stopPipelineMonitor() {
	if pc.monitorCancel != nil {
		pc.monitorCancel()
		pc.monitorCancel = nil
	}
}

func (pc *pipelineController) stopCrashingPipelineMonitor() {
	if pc.crashingMonitorCancel != nil {
		pc.crashingMonitorCancel()
		pc.crashingMonitorCancel = nil
	}
}

// finishPipelineOutputCommits finishes any output commits of
// 'pipelineInfo.Pipeline' with an empty tree.
// TODO(msteffen) Note that if the pipeline has any jobs (which can happen if
// the user manually deletes the pipeline's RC, failing the pipeline, after it
// has created jobs) those will not be updated, but they should be FAILED
//
// Unlike other functions in this file, finishPipelineOutputCommits doesn't
// cause retries if it encounters an error. Currently. it's only called by step()
// in the case where pc's pipeline is already in FAILURE. If it returns an error in
// that case, the pps master will log the error and move on to the next pipeline
// event. This pipeline's output commits will stay open until another watch
// event arrives for the pipeline and finishPipelineOutputCommits is retried.
func (pc *pipelineController) finishPipelineOutputCommits(ctx context.Context, pi *pps.PipelineInfo) (retErr error) {
	log.Infof("PPS master: finishing output commits for pipeline %q", pi.Pipeline.Name)
	pachClient := pc.env.GetPachClient(ctx)
	if span, _ctx := tracing.AddSpanToAnyExisting(ctx,
		"/pps.Master/FinishPipelineOutputCommits", "pipeline", pi.Pipeline.Name); span != nil {
		pachClient = pachClient.WithCtx(_ctx) // copy span back into pachClient
		defer func() {
			tracing.TagAnySpan(span, "err", fmt.Sprintf("%v", retErr))
			tracing.FinishAnySpan(span)
		}()
	}
	pachClient.SetAuthToken(pi.AuthToken)
	if err := pachClient.ListCommitF(client.NewRepo(pi.Pipeline.Name), client.NewCommit(pi.Pipeline.Name, pi.Details.OutputBranch, ""), nil, 0, false, func(commitInfo *pfs.CommitInfo) error {
		return pachClient.StopJob(pi.Pipeline.Name, commitInfo.Commit.ID)
	}); err != nil {
		if errutil.IsNotFoundError(err) {
			return nil // already deleted
		}
		return errors.Wrapf(err, "could not finish output commits of pipeline %q", pi.Pipeline.Name)
	}
	return nil
}

// scaleUpPipeline edits the RC associated with pc's pipeline & spins up the
// configured number of workers.
func (pc *pipelineController) scaleUpPipeline(ctx context.Context, pi *pps.PipelineInfo, oldRC *v1.ReplicationController) (retErr error) {
	log.Debugf("PPS master: ensuring correct k8s resources for %q", pi.Pipeline.Name)
	span, _ := tracing.AddSpanToAnyExisting(ctx,
		"/pps.Master/ScaleUpPipeline", "pipeline", pi.Pipeline.Name)
	defer func() {
		if retErr != nil {
			log.Errorf("PPS master: error scaling up: %v", retErr)
		}
		tracing.TagAnySpan(span, "err", retErr)
		tracing.FinishAnySpan(span)
	}()
	// Compute maximum parallelism
	maxScale := int32(1)
	if pi.Details.ParallelismSpec != nil && pi.Details.ParallelismSpec.Constant > 0 {
		maxScale = int32(pi.Details.ParallelismSpec.Constant)
	}
	// update pipeline RC
	return pc.kd.UpdateReplicationController(ctx, oldRC, func(rc *v1.ReplicationController) bool {
		var curScale int32
		if rc.Spec.Replicas != nil && *rc.Spec.Replicas > 0 {
			curScale = *rc.Spec.Replicas
		}
		targetScale := func() int32 {
			if !pi.Details.Autoscaling {
				return maxScale // don't bother if Autoscaling is off
			}
			if curScale == 0 {
				return 1 // make one pod to be the worker master & calculate tasks
			}
			// Master is scheduled; see if tasks have been calculated
			nTasks64, _, err := task.Count(ctx, pc.env.TaskService, driver.TaskNamespace(pi), "")
			nTasks := int32(nTasks64) // for cmp. with curScale, if err != nil
			// Set parallelism
			log.Debugf("Beginning scale-up check for %q, which has %d tasks",
				pi.Pipeline.Name, nTasks)
			switch {
			case err != nil || nTasks == 0:
				log.Errorf("tasks remaining for %q not known (possibly still being calculated): %v",
					pi.Pipeline.Name, err)
				return curScale // leave pipeline alone until until nTasks is available
			case nTasks <= curScale:
				return curScale // can't scale down w/o dropping work
			case nTasks <= maxScale:
				return nTasks
			default:
				return maxScale
			}
		}()
		if targetScale < maxScale {
			// schedule another step in scaleUpInterval, to check the tasks again
			go func() {
				time.Sleep(scaleUpInterval)
				// Normally, it's necessary to acquire the mutex in step.pc.pcMgr and
				// then read the latest pipelineController from pcMgr before calling
				// Bump(), in order to avoid Bumping a dead pipelineController and
				// dropping a Bump event. But in this case, we'd rather drop the Bump
				// event. If this pipeline was recently updated, the new pipeline may
				// not have autoscaling, or it may simply not make sense to trigger an
				// update anymore.
				if pc.ctx.Err() == nil {
					pc.Bump(time.Time{}) // no ts, as it's not a new event
				}
			}()
		}
		if curScale == targetScale {
			return false // no changes necessary
		}
		// Update the # of replicas
		rc.Spec.Replicas = &targetScale
		return true
	})
}

// scaleDownPipeline edits the RC associated with pc's pipeline & spins down the
// configured number of workers.
func (pc *pipelineController) scaleDownPipeline(ctx context.Context, pi *pps.PipelineInfo, rc *v1.ReplicationController) (retErr error) {
	log.Debugf("PPS master: scaling down workers for %q", pi.Pipeline.Name)
	span, _ := tracing.AddSpanToAnyExisting(ctx,
		"/pps.Master/ScaleDownPipeline", "pipeline", pi.Pipeline.Name)
	defer func() {
		if retErr != nil {
			log.Errorf("PPS master: error scaling down: %v", retErr)
		}
		tracing.TagAnySpan(span, "err", retErr)
		tracing.FinishAnySpan(span)
	}()
	return pc.kd.UpdateReplicationController(ctx, rc, func(rc *v1.ReplicationController) bool {
		if rc.Spec.Replicas != nil && *rc.Spec.Replicas == 0 {
			return false // prior attempt succeeded
		}
		rc.Spec.Replicas = &zero
		return true
	})
}

// restartPipeline updates the RC/service associated with pc's pipeline, and
// then sets its state to RESTARTING. Note that restartPipeline only deletes
// rc if it's stale--a prior bug was that it would delete all of pc's
// resources, and then get stuck in a loop deleting and recreating pc's RC if
// the cluster was busy and the RC was taking too long to start.
//
// restartPipeline is an error-handling
// codepath, so it's guaranteed to return an error (typically wrapping 'reason',
// though if the restart process fails that error will take precendence) so that
// callers can use it like so:
//
// if errorState {
//   return pc.restartPipeline("entered error state")
// }
func (pc *pipelineController) restartPipeline(ctx context.Context, pi *pps.PipelineInfo, rc *v1.ReplicationController, reason string) error {
	if rc != nil && !rcIsFresh(pi, rc) {
		// delete old RC, monitorPipeline goro, and worker service
		if err := pc.deletePipelineResources(); err != nil {
			return newRetriableError(err, "error deleting resources for restart")
		}
	}
	// create up-to-date RC
	if err := pc.kd.CreatePipelineResources(ctx, pi); err != nil {
		return errors.Wrap(err, "error creating resources for restart")
	}
	if err := pc.setPipelineState(ctx, pi.SpecCommit, pps.PipelineState_PIPELINE_RESTARTING, ""); err != nil {
		return errors.Wrap(err, "error restarting pipeline")
	}
	return errors.Errorf("restarting pipeline %q: %s", pi.Pipeline.Name, reason)
}

// deletePipelineResources deletes the monitors, k8s RC and k8s services associated with pc's
// pipeline. It doesn't return a stepError, leaving retry behavior to the caller
func (pc *pipelineController) deletePipelineResources() (retErr error) {
	// Cancel any running monitorPipeline call
	pc.stopPipelineMonitor()
	// Same for cancelCrashingMonitor
	pc.stopCrashingPipelineMonitor()
	return pc.kd.DeletePipelineResources(pc.ctx, pc.pipeline)
}

// Unlike other functions in this file, getRC takes responsibility for restarting
// pc's pipeline if it can't read the pipeline's RC (or if the RC is stale or
// redundant), and then returns an error to the caller to indicate that the
// caller shouldn't continue with other operations
func (pc *pipelineController) getRC(ctx context.Context, pi *pps.PipelineInfo) (rc *v1.ReplicationController, restart bool, retErr error) {
	span, _ := tracing.AddSpanToAnyExisting(ctx,
		"/pps.Master/GetRC", "pipeline", pc.pipeline)
	defer func(span opentracing.Span) {
		tracing.TagAnySpan(span, "err", fmt.Sprintf("%v", retErr))
		tracing.FinishAnySpan(span)
	}(span)
	// count error types separately, so that this only errors if the pipeline is
	// stuck and not changing
	var notFoundErrCount, unexpectedErrCount, staleErrCount, tooManyErrCount,
		otherErrCount int
	err := backoff.RetryNotify(func() error {
		var err error
		rcs, err := pc.kd.ReadReplicationController(pc.ctx, pi)
		if err != nil && !errutil.IsNotFoundError(err) {
			return errors.EnsureStack(err)
		}
		if len(rcs.Items) == 0 {
			rc = nil
			return errRCNotFound
		}
		rc = &rcs.Items[0]
		switch {
		case len(rcs.Items) > 1:
			// select stale RC if possible, so that we delete it in restartPipeline
			for i := range rcs.Items {
				rc = &rcs.Items[i]
				if !rcIsFresh(pi, rc) {
					break
				}
			}
			return errTooManyRCs
		case !rcIsFresh(pi, rc):
			return errStaleRC
		default:
			return nil
		}
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		switch {
		case errors.Is(err, errRCNotFound):
			notFoundErrCount++
		case errors.Is(err, errUnexpectedRC):
			unexpectedErrCount++
		case errors.Is(err, errTooManyRCs):
			tooManyErrCount++
		case errors.Is(err, errStaleRC):
			staleErrCount++ // don't return immediately b/c RC might be changing
		default:
			otherErrCount++
		}
		errCount := max(notFoundErrCount, unexpectedErrCount, staleErrCount,
			tooManyErrCount, otherErrCount)
		if errCount >= maxErrCount {
			invalidRCState := errors.Is(err, errTooManyRCs) || errors.Is(err, errStaleRC)
			if invalidRCState {
				restart = true
				return errutil.ErrBreak
			}
			return err //return whatever the most recent error was
		}
		log.Warnf("PPS master: error retrieving RC for %q: %v; retrying in %v", pc.pipeline, err, d)
		return nil
	})
	return rc, restart, err
}
