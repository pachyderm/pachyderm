package server

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/client/limit"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	middleware_auth "github.com/pachyderm/pachyderm/v2/src/internal/middleware/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing/extended"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs/pretty"
	"github.com/pachyderm/pachyderm/v2/src/version"

	opentracing "github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type rcExpectation byte

const (
	noExpectation rcExpectation = iota
	noRCExpected
	rcExpected
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
	// the limiter intends to guard the k8s API server from being overwhelmed by many concurrent requests
	// that could arise from many concurrent pipelineController goros.
	limiter limit.ConcurrencyLimiter
}

func newPcManager(maxConcurrentK8sRequests int) *pcManager {
	return &pcManager{
		pcs:     make(map[string]*pipelineController),
		limiter: limit.New(maxConcurrentK8sRequests),
	}
}

// pcStep captures the state that's used within a single handling of a pipeline event
type pcStep struct {
	pc           *pipelineController
	pipelineInfo *pps.PipelineInfo
	rc           *v1.ReplicationController
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

		bumpChan: make(chan time.Time, 1),
		pcMgr:    m.pcMgr,
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

// step takes 'pipelineInfo', a newly-changed pipeline pointer in the pipeline collection, and
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
		tracing.TagAnySpan(span, "pollpipelines-event", "true")
	}
	defer tracing.FinishAnySpan(span, "err", retErr)

	step, ctx, err := pc.newStep()
	if err != nil {
		// if we fail to create a new step, there was an error querying the pipeline info, and there's nothing we can do
		log.Errorf("PPS master: failed to set up step data to handle event for pipeline '%s': %v", pc.pipeline, errors.Wrapf(err, "failing pipeline %q", pc.pipeline))
		return false, err
	} else if step == nil {
		// interpret the event as a delete operation
		if err := pc.deletePipelineResources(); err != nil {
			log.Errorf("PPS master: error deleting pipelineController resources for pipeline '%s': %v", pc.pipeline, err)
			return true, errors.Wrapf(err, "error deleting pipelineController resources for pipeline '%s'", pc.pipeline)
		}
		return true, nil
	}
	var stepErr stepError
	errCount := 0
	err = backoff.RetryNotify(func() error {
		// set pc.rc
		// TODO(msteffen) should this fail the pipeline? (currently getRC will restart
		// the pipeline indefinitely)
		if err := step.getRC(ctx, noExpectation); err != nil && !errors.Is(err, errRCNotFound) {
			return err
		}

		// Create/Modify/Delete pipeline resources as needed per new state
		return step.updateRcAndState(ctx)
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
		failError := step.setPipelineFailure(ctx, fmt.Sprintf("could not update resources after %d attempts: %v", errCount, err))
		if failError != nil {
			log.Errorf("PPS master: error failing pipeline '%s' after step error (%v): %v", pc.pipeline, stepErr, failError)
			return false, errors.Wrapf(failError, "error failing pipeline %q  after step error (%v)", pc.pipeline, stepErr)
		}
	}
	return false, err
}

// returns nil, nil, nil if the step is found to be a delete operation
func (pc *pipelineController) newStep() (*pcStep, context.Context, error) {

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
	ctx := pachClient.Ctx()

	step := &pcStep{
		pc:           pc,
		pipelineInfo: pi,
	}

	return step, ctx, nil
}

func (step *pcStep) updateRcAndState(ctx context.Context) error {
	// Bring 'pipeline' into the correct state by taking appropriate action
	switch step.pipelineInfo.State {
	case pps.PipelineState_PIPELINE_STARTING, pps.PipelineState_PIPELINE_RESTARTING:
		if step.rc != nil && !step.rcIsFresh() {
			// old RC is not down yet
			return step.restartPipeline(ctx, "stale RC") // step() will be called again after collection write
		} else if step.rc == nil {
			// default: old RC (if any) is down but new RC is not up yet
			if err := step.createPipelineResources(ctx); err != nil {
				return err
			}
		}
		if step.pipelineInfo.Stopped {
			return step.setPipelineState(ctx, pps.PipelineState_PIPELINE_PAUSED, "")
		}
		step.pc.stopCrashingPipelineMonitor()
		// trigger another event
		target := pps.PipelineState_PIPELINE_RUNNING
		if step.pipelineInfo.Details.Autoscaling && step.pipelineInfo.State == pps.PipelineState_PIPELINE_STARTING {
			// start in standby
			target = pps.PipelineState_PIPELINE_STANDBY
		}
		return step.setPipelineState(ctx, target, "")
	case pps.PipelineState_PIPELINE_RUNNING:
		if !step.rcIsFresh() {
			return step.restartPipeline(ctx, "stale RC") // step() will be called again after collection write
		}
		if step.pipelineInfo.Stopped {
			return step.setPipelineState(ctx, pps.PipelineState_PIPELINE_PAUSED, "")
		}

		step.pc.stopCrashingPipelineMonitor()
		step.startPipelineMonitor()
		// default: scale up if pipeline start hasn't propagated to the collection yet
		// Note: mostly this should do nothing, as this runs several times per job
		return step.scaleUpPipeline(ctx)
	case pps.PipelineState_PIPELINE_STANDBY:
		if !step.rcIsFresh() {
			return step.restartPipeline(ctx, "stale RC") // step() will be called again after collection write
		}
		if step.pipelineInfo.Stopped {
			return step.setPipelineState(ctx, pps.PipelineState_PIPELINE_PAUSED, "")
		}

		step.pc.stopCrashingPipelineMonitor()
		// Make sure pipelineMonitor is running to pull it out of standby
		step.startPipelineMonitor()
		// default: scale down if standby hasn't propagated to kube RC yet
		return step.scaleDownPipeline(ctx)
	case pps.PipelineState_PIPELINE_PAUSED:
		if !step.rcIsFresh() {
			return step.restartPipeline(ctx, "stale RC") // step() will be called again after collection write
		}
		if !step.pipelineInfo.Stopped {
			// StartPipeline has been called (so spec commit is updated), but new spec
			// commit hasn't been propagated to PipelineInfo or RC yet
			target := pps.PipelineState_PIPELINE_RUNNING
			if step.pipelineInfo.Details.Autoscaling {
				target = pps.PipelineState_PIPELINE_STANDBY
			}
			return step.setPipelineState(ctx, target, "")
		}
		// don't want cron commits or STANDBY state changes while pipeline is
		// stopped
		step.pc.stopPipelineMonitor()
		step.pc.stopCrashingPipelineMonitor()
		// default: scale down if pause/standby hasn't propagated to collection yet
		return step.scaleDownPipeline(ctx)
	case pps.PipelineState_PIPELINE_FAILURE:
		// pipeline fails if it encounters an unrecoverable error
		if err := step.finishPipelineOutputCommits(ctx); err != nil {
			return err
		}
		// deletePipelineResources calls cancelMonitor() and cancelCrashingMonitor()
		// in addition to deleting the RC, so those calls aren't necessary here.
		if err := step.pc.deletePipelineResources(); err != nil {
			// retry, but the pipeline has already failed
			return stepError{
				error: errors.Wrap(err, "error deleting resources for failing pipeline"),
				retry: true,
			}
		}
		return nil
	case pps.PipelineState_PIPELINE_CRASHING:
		if !step.rcIsFresh() {
			return step.restartPipeline(ctx, "stale RC") // step() will be called again after collection write
		}
		if step.pipelineInfo.Stopped {
			return step.setPipelineState(ctx, pps.PipelineState_PIPELINE_PAUSED, "")
		}
		// start a monitor to poll k8s and update us when it goes into a running state
		step.startPipelineMonitor()
		step.startCrashingPipelineMonitor()
		// Surprisingly, scaleUpPipeline() is necessary, in case a pipelines is
		// quickly transitioned to CRASHING after coming out of STANDBY. Because the
		// pipeline controller reads the current state of the pipeline after each
		// event (to avoid getting backlogged), it might never actually see the
		// pipeline in RUNNING. However, if the RC is never scaled up, the pipeline
		// can never come out of CRASHING, so do it here in case it never happened.
		//
		// In general, CRASHING is actually almost identical to RUNNING (except for
		// the monitorCrashing goro)
		return step.scaleUpPipeline(ctx)
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

// rcIsFresh returns a boolean indicating whether pc.rc has the right labels
// corresponding to pc.pipelineInfo. If this returns false, it likely means the
// current RC is using e.g. an old spec commit or something.
func (step *pcStep) rcIsFresh() bool {
	if step.rc == nil {
		log.Errorf("PPS master: RC for %q is nil", step.pipelineInfo.Pipeline.Name)
		return false
	}
	expectedName := ""
	if step.pipelineInfo != nil {
		expectedName = ppsutil.PipelineRcName(step.pipelineInfo.Pipeline.Name, step.pipelineInfo.Version)
	}

	// establish current RC properties
	rcName := step.rc.ObjectMeta.Name
	rcPachVersion := step.rc.ObjectMeta.Annotations[pachVersionAnnotation]
	rcAuthTokenHash := step.rc.ObjectMeta.Annotations[hashedAuthTokenAnnotation]
	rcPipelineVersion := step.rc.ObjectMeta.Annotations[pipelineVersionAnnotation]
	rcSpecCommit := step.rc.ObjectMeta.Annotations[pipelineSpecCommitAnnotation]
	switch {
	case rcAuthTokenHash != hashAuthToken(step.pipelineInfo.AuthToken):
		log.Errorf("PPS master: auth token in %q is stale %s != %s",
			step.pipelineInfo.Pipeline.Name, rcAuthTokenHash, hashAuthToken(step.pipelineInfo.AuthToken))
		return false
	case rcPipelineVersion != strconv.FormatUint(step.pipelineInfo.Version, 10):
		log.Errorf("PPS master: pipeline version in %q looks stale %s != %d",
			step.pipelineInfo.Pipeline.Name, rcPipelineVersion, step.pipelineInfo.Version)
		return false
	case rcSpecCommit != step.pipelineInfo.SpecCommit.ID:
		log.Errorf("PPS master: pipeline spec commit in %q looks stale %s != %s",
			step.pipelineInfo.Pipeline.Name, rcSpecCommit, step.pipelineInfo.SpecCommit.ID)
		return false
	case rcPachVersion != version.PrettyVersion():
		log.Errorf("PPS master: %q is using stale pachd v%s != current v%s",
			step.pipelineInfo.Pipeline.Name, rcPachVersion, version.PrettyVersion())
		return false
	case expectedName != "" && rcName != expectedName:
		log.Errorf("PPS master: %q has an unexpected (likely stale) name %q != %q",
			step.pipelineInfo.Pipeline.Name, rcName, expectedName)
	}
	return true
}

// setPipelineState set's pc's state in the collection to 'state'. This will trigger a
// collection watch event and cause step() to eventually run again.
func (step *pcStep) setPipelineState(ctx context.Context, state pps.PipelineState, reason string) error {
	if err := setPipelineState(ctx, step.pc.env.DB, step.pc.pipelines, step.pipelineInfo.SpecCommit, state, reason); err != nil {
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
func (step *pcStep) startPipelineMonitor() {
	// since *pc.pipelineInfo may be modified on a Bump, and startMonitor() passes its
	// input pipelineInfo to a monitor goroutine, send a pointer to a copy of pc.pipelineInfo to
	// avoid a race condition
	pi := *step.pipelineInfo
	if step.pc.monitorCancel == nil {
		step.pc.monitorCancel = step.pc.startMonitor(step.pc.ctx, &pi)
	}
	step.pipelineInfo.Details.WorkerRc = step.rc.ObjectMeta.Name
}

func (step *pcStep) startCrashingPipelineMonitor() {
	// since *pc.pipelineInfo may be modified on a Bump, and startMonitor() passes its
	// input pipelineInfo to a monitor goroutine, send a pointer to a copy of pc.pipelineInfo to
	// avoid a race condition
	pi := *step.pipelineInfo
	if step.pc.crashingMonitorCancel == nil {
		step.pc.crashingMonitorCancel = step.pc.startCrashingMonitor(step.pc.ctx, &pi)
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
func (step *pcStep) finishPipelineOutputCommits(ctx context.Context) (retErr error) {
	log.Infof("PPS master: finishing output commits for pipeline %q", step.pipelineInfo.Pipeline.Name)

	pachClient := step.pc.env.GetPachClient(ctx)
	if span, _ctx := tracing.AddSpanToAnyExisting(ctx,
		"/pps.Master/FinishPipelineOutputCommits", "pipeline", step.pipelineInfo.Pipeline.Name); span != nil {
		pachClient = pachClient.WithCtx(_ctx) // copy span back into pachClient
		defer func() {
			tracing.TagAnySpan(span, "err", fmt.Sprintf("%v", retErr))
			tracing.FinishAnySpan(span)
		}()
	}
	pachClient.SetAuthToken(step.pipelineInfo.AuthToken)

	if err := pachClient.ListCommitF(client.NewRepo(step.pipelineInfo.Pipeline.Name), client.NewCommit(step.pipelineInfo.Pipeline.Name, step.pipelineInfo.Details.OutputBranch, ""), nil, 0, false, func(commitInfo *pfs.CommitInfo) error {
		return pachClient.StopJob(step.pipelineInfo.Pipeline.Name, commitInfo.Commit.ID)
	}); err != nil {
		if errutil.IsNotFoundError(err) {
			return nil // already deleted
		}
		return errors.Wrapf(err, "could not finish output commits of pipeline %q", step.pipelineInfo.Pipeline.Name)
	}
	return nil
}

// scaleUpPipeline edits the RC associated with pc's pipeline & spins up the
// configured number of workers.
func (step *pcStep) scaleUpPipeline(ctx context.Context) (retErr error) {
	log.Debugf("PPS master: ensuring correct k8s resources for %q", step.pipelineInfo.Pipeline.Name)
	span, _ := tracing.AddSpanToAnyExisting(ctx,
		"/pps.Master/ScaleUpPipeline", "pipeline", step.pipelineInfo.Pipeline.Name)
	defer func() {
		if retErr != nil {
			log.Errorf("PPS master: error scaling up: %v", retErr)
		}
		tracing.TagAnySpan(span, "err", retErr)
		tracing.FinishAnySpan(span)
	}()

	// compute target pipeline parallelism
	parallelism := uint64(1)
	if step.pipelineInfo.Details.ParallelismSpec != nil {
		parallelism = step.pipelineInfo.Details.ParallelismSpec.Constant
	}

	// update pipeline RC
	return step.updateRC(ctx, func(rc *v1.ReplicationController) {
		if rc.Spec.Replicas != nil && *step.rc.Spec.Replicas > 0 {
			return // prior attempt succeeded
		}
		rc.Spec.Replicas = new(int32)
		if step.pipelineInfo.Details.Autoscaling {
			*rc.Spec.Replicas = 1
		} else {
			*rc.Spec.Replicas = int32(parallelism)
		}
	})
}

// scaleDownPipeline edits the RC associated with pc's pipeline & spins down the
// configured number of workers.
func (step *pcStep) scaleDownPipeline(ctx context.Context) (retErr error) {
	log.Debugf("PPS master: scaling down workers for %q", step.pipelineInfo.Pipeline.Name)
	span, _ := tracing.AddSpanToAnyExisting(ctx,
		"/pps.Master/ScaleDownPipeline", "pipeline", step.pipelineInfo.Pipeline.Name)
	defer func() {
		if retErr != nil {
			log.Errorf("PPS master: error scaling down: %v", retErr)
		}
		tracing.TagAnySpan(span, "err", retErr)
		tracing.FinishAnySpan(span)
	}()

	return step.updateRC(ctx, func(rc *v1.ReplicationController) {
		if rc.Spec.Replicas != nil && *step.rc.Spec.Replicas == 0 {
			return // prior attempt succeeded
		}
		rc.Spec.Replicas = &zero
	})
}

// restartPipeline updates the RC/service associated with pc's pipeline, and
// then sets its state to RESTARTING. Note that restartPipeline only deletes
// pc.rc if it's stale--a prior bug was that it would delete all of pc's
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
func (step *pcStep) restartPipeline(ctx context.Context, reason string) error {
	if step.rc != nil && !step.rcIsFresh() {
		// delete old RC, monitorPipeline goro, and worker service
		if err := step.pc.deletePipelineResources(); err != nil {
			return newRetriableError(err, "error deleting resources for restart")
		}
	}
	// create up-to-date RC
	if err := step.createPipelineResources(ctx); err != nil {
		return errors.Wrap(err, "error creating resources for restart")
	}
	if err := step.setPipelineState(ctx, pps.PipelineState_PIPELINE_RESTARTING, ""); err != nil {
		return errors.Wrap(err, "error restarting pipeline")
	}

	return errors.Errorf("restarting pipeline %q: %s", step.pipelineInfo.Pipeline.Name, reason)
}

func (step *pcStep) setPipelineFailure(ctx context.Context, reason string) error {
	return step.setPipelineState(ctx, pps.PipelineState_PIPELINE_FAILURE, reason)
}

// deletePipelineResources deletes the monitors, k8s RC and k8s services associated with pc's
// pipeline. It doesn't return a stepError, leaving retry behavior to the caller
func (pc *pipelineController) deletePipelineResources() (retErr error) {
	// Cancel any running monitorPipeline call
	pc.stopPipelineMonitor()
	// Same for cancelCrashingMonitor
	pc.stopCrashingPipelineMonitor()
	return pc.deletePipelineKubeResources()
}

//============================================================================================
//
// The below functions all acquire the counting semaphore lock: the pc.pcMgr.limiter to
// prevent too many concurrent requests against kubernetes on behalf of different pipelines.
// The functions below this block should not call the functions above it to avoid deadlocks.
//
//============================================================================================

// getRC reads the RC associated with 'pc's pipeline. pc.pipelineInfo must be
// set already. 'expectation' indicates whether the PPS master expects an RC to
// exist--if set to 'rcExpected', getRC will restart the pipeline if no RC is
// found after three retries. If set to 'noRCExpected', then getRC will return
// after the first "not found" error. If set to noExpectation, then getRC will
// retry the kubeclient.List() RPC, but will not restart the pipeline if no RC
// is found
//
// Unlike other functions in this file, getRC takes responsibility for restarting
// pc's pipeline if it can't read the pipeline's RC (or if the RC is stale or
// redundant), and then returns an error to the caller to indicate that the
// caller shouldn't continue with other operations
func (step *pcStep) getRC(ctx context.Context, expectation rcExpectation) (retErr error) {
	span, _ := tracing.AddSpanToAnyExisting(ctx,
		"/pps.Master/GetRC", "pipeline", step.pc.pipeline)
	defer func(span opentracing.Span) {
		tracing.TagAnySpan(span, "err", fmt.Sprintf("%v", retErr))
		tracing.FinishAnySpan(span)
	}(span)

	step.pc.pcMgr.limiter.Acquire()
	defer step.pc.pcMgr.limiter.Release()

	selector := fmt.Sprintf("%s=%s", pipelineNameLabel, step.pc.pipeline)

	// count error types separately, so that this only errors if the pipeline is
	// stuck and not changing
	var notFoundErrCount, unexpectedErrCount, staleErrCount, tooManyErrCount,
		otherErrCount int
	return backoff.RetryNotify(func() error {
		// List all RCs, so stale RCs from old pipelines are noticed and deleted
		rcs, err := step.pc.env.KubeClient.CoreV1().ReplicationControllers(step.pc.namespace).List(
			ctx,
			metav1.ListOptions{LabelSelector: selector})
		if err != nil && !errutil.IsNotFoundError(err) {
			return errors.EnsureStack(err)
		}
		if len(rcs.Items) == 0 {
			step.rc = nil
			return errRCNotFound
		}

		step.rc = &rcs.Items[0]
		switch {
		case len(rcs.Items) > 1:
			// select stale RC if possible, so that we delete it in restartPipeline
			for i := range rcs.Items {
				step.rc = &rcs.Items[i]
				if !step.rcIsFresh() {
					break
				}
			}
			return errTooManyRCs
		case !step.rcIsFresh():
			return errStaleRC
		case expectation == noRCExpected:
			return errUnexpectedRC
		default:
			return nil
		}
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		if expectation == noRCExpected && errors.Is(err, errRCNotFound) {
			return err // rc has come down successfully--no need to keep looking
		}
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
			missingExpectedRC := expectation == rcExpected && errors.Is(err, errRCNotFound)
			invalidRCState := errors.Is(err, errTooManyRCs) || errors.Is(err, errStaleRC)
			if missingExpectedRC || invalidRCState {
				return step.restartPipeline(ctx, fmt.Sprintf("could not get RC after %d attempts: %v", errCount, err))
			}
			return err //return whatever the most recent error was
		}
		log.Errorf("PPS master: error retrieving RC for %q: %v; retrying in %v", step.pc.pipeline, err, d)
		return nil
	})
}

// updateRC is a helper for {scaleUp,scaleDown}Pipeline. It includes all of the
// logic for writing an updated RC spec to kubernetes, and updating/retrying if
// k8s rejects the write. It presents a strange API, since the the RC being
// updated is already available to the caller in pc.rc, but update() may be
// called muliple times if the k8s write fails. It may be helpful to think of
// the rc passed to update() as mutable, while pc.rc is immutable.
func (step *pcStep) updateRC(ctx context.Context, update func(rc *v1.ReplicationController)) error {

	step.pc.pcMgr.limiter.Acquire()
	defer step.pc.pcMgr.limiter.Release()

	rc := step.pc.env.KubeClient.CoreV1().ReplicationControllers(step.pc.namespace)

	newRC := *step.rc
	// Apply op's update to rc
	update(&newRC)
	// write updated RC to k8s
	if _, err := rc.Update(ctx, &newRC, metav1.UpdateOptions{}); err != nil {
		return newRetriableError(err, "error updating RC")
	}
	return nil
}

// createPipelineResources creates the RC and any services for pc's pipeline.
func (step *pcStep) createPipelineResources(ctx context.Context) error {
	log.Infof("PPS master: creating resources for pipeline %q", step.pipelineInfo.Pipeline.Name)

	step.pc.pcMgr.limiter.Acquire()
	defer step.pc.pcMgr.limiter.Release()

	if err := step.pc.createWorkerSvcAndRc(ctx, step.pipelineInfo); err != nil {
		if errors.As(err, &noValidOptionsErr{}) {
			// these errors indicate invalid pipelineInfo, don't retry
			return stepError{
				error:        errors.Wrap(err, "could not generate RC options"),
				failPipeline: true,
			}
		}
		return newRetriableError(err, "error creating resources")
	}
	return nil
}

// deletePipelineKubeResources deletes the RC and services associated with pc's
// pipeline. It doesn't return a stepError, leaving retry behavior to the caller
func (pc *pipelineController) deletePipelineKubeResources() (retErr error) {
	log.Infof("PPS master: deleting resources for pipeline %q", pc.pipeline)
	span, ctx := tracing.AddSpanToAnyExisting(pc.ctx,
		"/pps.Master/DeletePipelineResources", "pipeline", pc.pipeline)
	defer func() {
		tracing.TagAnySpan(ctx, "err", retErr)
		tracing.FinishAnySpan(span)
	}()

	pc.pcMgr.limiter.Acquire()
	defer pc.pcMgr.limiter.Release()

	kubeClient := pc.env.KubeClient
	namespace := pc.namespace

	// Delete any services associated with pc.pipeline
	selector := fmt.Sprintf("%s=%s", pipelineNameLabel, pc.pipeline)
	opts := metav1.DeleteOptions{
		OrphanDependents: &falseVal,
	}
	services, err := kubeClient.CoreV1().Services(namespace).List(pc.ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return errors.Wrapf(err, "could not list services")
	}
	for _, service := range services.Items {
		if err := kubeClient.CoreV1().Services(namespace).Delete(pc.ctx, service.Name, opts); err != nil {
			if !errutil.IsNotFoundError(err) {
				return errors.Wrapf(err, "could not delete service %q", service.Name)
			}
		}
	}

	// Delete any secrets associated with pc.pipeline
	secrets, err := kubeClient.CoreV1().Secrets(namespace).List(pc.ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return errors.Wrapf(err, "could not list secrets")
	}
	for _, secret := range secrets.Items {
		if err := kubeClient.CoreV1().Secrets(namespace).Delete(pc.ctx, secret.Name, opts); err != nil {
			if !errutil.IsNotFoundError(err) {
				return errors.Wrapf(err, "could not delete secret %q", secret.Name)
			}
		}
	}

	// Finally, delete pc.pipeline's RC, which will cause pollPipelines to stop
	// polling it.
	rcs, err := kubeClient.CoreV1().ReplicationControllers(namespace).List(pc.ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return errors.Wrapf(err, "could not list RCs")
	}
	for _, rc := range rcs.Items {
		if err := kubeClient.CoreV1().ReplicationControllers(namespace).Delete(pc.ctx, rc.Name, opts); err != nil {
			if !errutil.IsNotFoundError(err) {
				return errors.Wrapf(err, "could not delete RC %q", rc.Name)
			}
		}
	}

	return nil
}
