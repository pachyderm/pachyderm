package server

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
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

// pipelineOp contains all of the relevent current state for a pipeline. It's
// used by step() to take any necessary actions
type pipelineOp struct {
	// a pachyderm client wrapping this operation's context (child of the PPS
	// master's context, and cancelled at the end of Start())
	opCtx        context.Context
	opCancel     context.CancelFunc
	pipeline     string
	pipelineInfo *pps.PipelineInfo
	rc           *v1.ReplicationController
	namespace    string
	env          Env
	txEnv        *transactionenv.TransactionEnv
	pipelines    collection.PostgresCollection
	etcdPrefix   string

	monitorCancel         func()
	crashingMonitorCancel func()

	bumpChan  chan struct{}
	opManager *opManager
}

var (
	errRCNotFound   = errors.New("RC not found")
	errUnexpectedRC = errors.New("unexpected RC")
	errTooManyRCs   = errors.New("multiple RCs found for pipeline")
	errStaleRC      = errors.New("RC doesn't match pipeline version (likely stale)")
)

func (m *ppsMaster) newPipelineOp(ctx context.Context, cancel context.CancelFunc, pipeline string) *pipelineOp {
	op := &pipelineOp{
		opCtx:    ctx,
		opCancel: cancel,
		// pipeline name is recorded separately in the case we are running a delete Op and pipelineInfo isn't available in the DB
		pipeline:     pipeline,
		pipelineInfo: &pps.PipelineInfo{},
		namespace:    m.a.namespace,
		env:          m.a.env,
		txEnv:        m.a.txnEnv,
		pipelines:    m.a.pipelines,
		etcdPrefix:   m.a.etcdPrefix,

		bumpChan:  make(chan struct{}, 1),
		opManager: m.om,
	}
	return op
}

// Bump signals the pipelineOp goro to either process the  latest state
// of the pipeline now, or once its done with its current processing iteration.
//
// This function is expected to be called while holding the opManager lock.
// This helps guard the critical section when the goroutine is cleaned up
// as part of a pipeline deletion event in pipelineOp.tryFinish().
//
// Note: since op.bumpChan is a buffered channel of length 1, Bump() will
// add to the channel if it's empty, and do nothing otherwise.
func (op *pipelineOp) Bump() {
	select {
	case op.bumpChan <- struct{}{}:
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
func (op *pipelineOp) Start(timestamp time.Time) {
	op.Bump()
	for {
		select {
		case <-op.opCtx.Done():
			return
		case <-op.bumpChan:
			var errCount int
			var stepErr stepError
			err := backoff.RetryNotify(func() error {
				// Create/Modify/Delete pipeline resources as needed per new state
				return op.step(timestamp)
			}, backoff.NewExponentialBackOff(), func(err error, d time.Duration) error {
				errCount++
				if errors.As(err, &stepErr) {
					if stepErr.retry && errCount < maxErrCount {
						log.Errorf("PPS master: error updating resources for pipeline %q: %v; retrying in %v",
							op.pipeline, err, d)
						return nil
					}
				}
				return errors.Wrapf(err, "could not update resource for pipeline %q", op.pipeline)
			})

			// we've given up on the step, check if the error indicated that the pipeline should fail
			if err != nil {
				if errors.As(err, &stepErr) && stepErr.failPipeline {
					failError := op.setPipelineFailure(op.opCtx, fmt.Sprintf("could not update resources after %d attempts: %v", errCount, err))
					if failError != nil {
						log.Errorf("PPS master: error starting a pipelineOp for pipeline '%s': %v", op.pipeline,
							errors.Wrapf(failError, "error failing pipeline %q (%v)", op.pipeline, err))
					}
				}
				log.Errorf("PPS master: error starting a pipelineOp for pipeline '%s': %v", op.pipeline,
					errors.Wrapf(err, "failing pipeline %q", op.pipeline))
			}
		}
	}
}

// finishes the op if it isn't bumped
func (op *pipelineOp) tryFinish() {
	op.opManager.Lock()
	defer op.opManager.Unlock()
	select {
	case <-op.bumpChan:
		op.Bump()
	default:
		op.opCancel()
		delete(op.opManager.activeOps, op.pipeline)
	}
}

// step takes 'pipelineInfo', a newly-changed pipeline pointer in the pipeline collection, and
// 1. retrieves its full pipeline spec and RC into the 'Details' field
// 2. makes whatever changes are needed to bring the RC in line with the (new) spec
// 3. updates 'pipelineInfo', if needed, to reflect the action it just took
func (op *pipelineOp) step(timestamp time.Time) (retErr error) {
	log.Debugf("PPS master: processing event for %q", op.pipelineInfo.Pipeline)

	// Handle tracing
	span, _ := extended.AddSpanToAnyPipelineTrace(op.opCtx,
		op.env.EtcdClient, op.pipeline, "/pps.Master/ProcessPipelineUpdate")
	if !timestamp.IsZero() {
		tracing.TagAnySpan(span, "update-time", timestamp)
	} else {
		tracing.TagAnySpan(span, "pollpipelines-event", "true")
	}
	defer func() {
		tracing.FinishAnySpan(span, "err", retErr)
	}()

	// set op.pipelineInfo
	if err := op.tryLoadLatestPipelineInfo(); err != nil && collection.IsErrNotFound(err) {
		// if the pipeline info is not found, interpret the operation as a delete
		if err := op.deletePipelineResources(); err != nil {
			log.Errorf("PPS master: error deleting pipelineOp resources for pipeline '%s': %v", op.pipeline,
				errors.Wrapf(err, "failing pipeline %q", op.pipeline))
		}
		op.tryFinish()
		return nil
	} else if err != nil {
		return err
	}

	tracing.TagAnySpan(op.opCtx,
		"current-state", op.pipelineInfo.State.String(),
		"spec-commit", pretty.CompactPrintCommitSafe(op.pipelineInfo.SpecCommit))
	// add pipeline auth
	// the pipelineOp's context is authorized as pps master, but we want to switch to the pipeline itself
	// first clear the cached WhoAmI result from the context
	pachClient := op.env.GetPachClient(middleware_auth.ClearWhoAmI(op.opCtx))
	pachClient.SetAuthToken(op.pipelineInfo.AuthToken)
	stepCtx := pachClient.Ctx()

	// set op.rc
	// TODO(msteffen) should this fail the pipeline? (currently getRC will restart
	// the pipeline indefinitely)
	if err := op.getRC(stepCtx, noExpectation); err != nil && !errors.Is(err, errRCNotFound) {
		return err
	}

	// Process the pipeline event
	return op.run(stepCtx)
}

func (op *pipelineOp) run(stepCtx context.Context) error {
	// Bring 'pipeline' into the correct state by taking appropriate action
	switch op.pipelineInfo.State {
	case pps.PipelineState_PIPELINE_STARTING, pps.PipelineState_PIPELINE_RESTARTING:
		if op.rc != nil && !op.rcIsFresh() {
			// old RC is not down yet
			return op.restartPipeline(stepCtx, "stale RC") // step() will be called again after collection write
		} else if op.rc == nil {
			// default: old RC (if any) is down but new RC is not up yet
			if err := op.createPipelineResources(stepCtx); err != nil {
				return err
			}
		}
		if op.pipelineInfo.Stopped {
			return op.setPipelineState(stepCtx, pps.PipelineState_PIPELINE_PAUSED, "")
		}
		op.stopCrashingPipelineMonitor()
		// trigger another event
		target := pps.PipelineState_PIPELINE_RUNNING
		if op.pipelineInfo.Details.Autoscaling && op.pipelineInfo.State == pps.PipelineState_PIPELINE_STARTING {
			// start in standby
			target = pps.PipelineState_PIPELINE_STANDBY
		}
		return op.setPipelineState(stepCtx, target, "")
	case pps.PipelineState_PIPELINE_RUNNING:
		if !op.rcIsFresh() {
			return op.restartPipeline(stepCtx, "stale RC") // step() will be called again after collection write
		}
		if op.pipelineInfo.Stopped {
			return op.setPipelineState(stepCtx, pps.PipelineState_PIPELINE_PAUSED, "")
		}

		op.stopCrashingPipelineMonitor()
		op.startPipelineMonitor()
		// default: scale up if pipeline start hasn't propagated to the collection yet
		// Note: mostly this should do nothing, as this runs several times per job
		return op.scaleUpPipeline(stepCtx)
	case pps.PipelineState_PIPELINE_STANDBY:
		if !op.rcIsFresh() {
			return op.restartPipeline(stepCtx, "stale RC") // step() will be called again after collection write
		}
		if op.pipelineInfo.Stopped {
			return op.setPipelineState(stepCtx, pps.PipelineState_PIPELINE_PAUSED, "")
		}

		op.stopCrashingPipelineMonitor()
		// Make sure pipelineMonitor is running to pull it out of standby
		op.startPipelineMonitor()
		// default: scale down if standby hasn't propagated to kube RC yet
		return op.scaleDownPipeline(stepCtx)
	case pps.PipelineState_PIPELINE_PAUSED:
		if !op.rcIsFresh() {
			return op.restartPipeline(stepCtx, "stale RC") // step() will be called again after collection write
		}
		if !op.pipelineInfo.Stopped {
			// StartPipeline has been called (so spec commit is updated), but new spec
			// commit hasn't been propagated to PipelineInfo or RC yet
			target := pps.PipelineState_PIPELINE_RUNNING
			if op.pipelineInfo.Details.Autoscaling {
				target = pps.PipelineState_PIPELINE_STANDBY
			}
			return op.setPipelineState(stepCtx, target, "")
		}
		// don't want cron commits or STANDBY state changes while pipeline is
		// stopped
		op.stopPipelineMonitor()
		op.stopCrashingPipelineMonitor()
		// default: scale down if pause/standby hasn't propagated to collection yet
		return op.scaleDownPipeline(stepCtx)
	case pps.PipelineState_PIPELINE_FAILURE:
		// pipeline fails if it encounters an unrecoverable error
		if err := op.finishPipelineOutputCommits(stepCtx); err != nil {
			return err
		}
		// deletePipelineResources calls cancelMonitor() and cancelCrashingMonitor()
		// in addition to deleting the RC, so those calls aren't necessary here.
		if err := op.deletePipelineResources(); err != nil {
			// retry, but the pipeline has already failed
			return stepError{
				error: errors.Wrap(err, "error deleting resources for failing pipeline"),
				retry: true,
			}
		}
		return nil
	case pps.PipelineState_PIPELINE_CRASHING:
		if !op.rcIsFresh() {
			return op.restartPipeline(stepCtx, "stale RC") // step() will be called again after collection write
		}
		if op.pipelineInfo.Stopped {
			return op.setPipelineState(stepCtx, pps.PipelineState_PIPELINE_PAUSED, "")
		}
		// start a monitor to poll k8s and update us when it goes into a running state
		op.startPipelineMonitor()
		op.startCrashingPipelineMonitor()
		// Surprisingly, scaleUpPipeline() is necessary, in case a pipelines is
		// quickly transitioned to CRASHING after coming out of STANDBY. Because the
		// pipeline controller reads the current state of the pipeline after each
		// event (to avoid getting backlogged), it might never actually see the
		// pipeline in RUNNING. However, if the RC is never scaled up, the pipeline
		// can never come out of CRASHING, so do it here in case it never happened.
		//
		// In general, CRASHING is actually almost identical to RUNNING (except for
		// the monitorCrashing goro)
		return op.scaleUpPipeline(stepCtx)
	}
	return nil
}

func (op *pipelineOp) tryLoadLatestPipelineInfo() error {
	errCnt := 0
	err := backoff.RetryNotify(func() error {
		return op.loadLatestPipelineInfo()
	}, backoff.NewExponentialBackOff(), func(err error, d time.Duration) error {
		errCnt++
		// Don't put the pipeline in a failing state if we're in the middle
		// of activating auth, retry in a bit
		if (auth.IsErrNotAuthorized(err) || auth.IsErrNotSignedIn(err)) && errCnt <= maxErrCount {
			log.Warnf("PPS master: could not retrieve pipelineInfo for pipeline %q: %v; retrying in %v",
				op.pipeline, err, d)
			return nil
		}
		return stepError{
			error: errors.Wrapf(err, "could not load pipelineInfo for pipeline %q", op.pipeline),
			retry: false,
		}
	})
	return err
}

func (op *pipelineOp) loadLatestPipelineInfo() error {
	*op.pipelineInfo = pps.PipelineInfo{}
	specCommit, err := ppsutil.FindPipelineSpecCommit(op.opCtx, op.env.PFSServer, *op.txEnv, op.pipeline)
	if err != nil {
		return errors.Wrapf(err, "could not find spec commit for pipeline %q", op.pipeline)
	}
	if err := op.pipelines.ReadOnly(op.opCtx).Get(specCommit, op.pipelineInfo); err != nil {
		return errors.Wrapf(err, "could not retrieve pipeline info for %q", op.pipeline)
	}
	return nil
}

// getRC reads the RC associated with 'op's pipeline. op.pipelineInfo must be
// set already. 'expectation' indicates whether the PPS master expects an RC to
// exist--if set to 'rcExpected', getRC will restart the pipeline if no RC is
// found after three retries. If set to 'noRCExpected', then getRC will return
// after the first "not found" error. If set to noExpectation, then getRC will
// retry the kubeclient.List() RPC, but will not restart the pipeline if no RC
// is found
//
// Unlike other functions in this file, getRC takes responsibility for restarting
// op's pipeline if it can't read the pipeline's RC (or if the RC is stale or
// redundant), and then returns an error to the caller to indicate that the
// caller shouldn't continue with other operations
func (op *pipelineOp) getRC(ctx context.Context, expectation rcExpectation) (retErr error) {
	span, _ := tracing.AddSpanToAnyExisting(ctx,
		"/pps.Master/GetRC", "pipeline", op.pipeline)
	defer func(span opentracing.Span) {
		tracing.TagAnySpan(span, "err", fmt.Sprintf("%v", retErr))
		tracing.FinishAnySpan(span)
	}(span)

	selector := fmt.Sprintf("%s=%s", pipelineNameLabel, op.pipeline)

	// count error types separately, so that this only errors if the pipeline is
	// stuck and not changing
	var notFoundErrCount, unexpectedErrCount, staleErrCount, tooManyErrCount,
		otherErrCount int
	return backoff.RetryNotify(func() error {
		// List all RCs, so stale RCs from old pipelines are noticed and deleted
		rcs, err := op.env.KubeClient.CoreV1().ReplicationControllers(op.namespace).List(
			metav1.ListOptions{LabelSelector: selector})
		if err != nil && !errutil.IsNotFoundError(err) {
			return err
		}
		if len(rcs.Items) == 0 {
			op.rc = nil
			return errRCNotFound
		}

		op.rc = &rcs.Items[0]
		switch {
		case len(rcs.Items) > 1:
			// select stale RC if possible, so that we delete it in restartPipeline
			for i := range rcs.Items {
				op.rc = &rcs.Items[i]
				if !op.rcIsFresh() {
					break
				}
			}
			return errTooManyRCs
		case !op.rcIsFresh():
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
				return op.restartPipeline(ctx, fmt.Sprintf("could not get RC after %d attempts: %v", errCount, err))
			}
			return err //return whatever the most recent error was
		}
		log.Errorf("PPS master: error retrieving RC for %q: %v; retrying in %v", op.pipeline, err, d)
		return nil
	})
}

// rcIsFresh returns a boolean indicating whether op.rc has the right labels
// corresponding to op.pipelineInfo. If this returns false, it likely means the
// current RC is using e.g. an old spec commit or something.
func (op *pipelineOp) rcIsFresh() bool {
	if op.rc == nil {
		log.Errorf("PPS master: RC for %q is nil", op.pipelineInfo.Pipeline.Name)
		return false
	}
	expectedName := ""
	if op.pipelineInfo != nil {
		expectedName = ppsutil.PipelineRcName(op.pipelineInfo.Pipeline.Name, op.pipelineInfo.Version)
	}

	// establish current RC properties
	rcName := op.rc.ObjectMeta.Name
	rcPachVersion := op.rc.ObjectMeta.Annotations[pachVersionAnnotation]
	rcAuthTokenHash := op.rc.ObjectMeta.Annotations[hashedAuthTokenAnnotation]
	rcPipelineVersion := op.rc.ObjectMeta.Annotations[pipelineVersionAnnotation]
	switch {
	case rcAuthTokenHash != hashAuthToken(op.pipelineInfo.AuthToken):
		log.Errorf("PPS master: auth token in %q is stale %s != %s",
			op.pipelineInfo.Pipeline.Name, rcAuthTokenHash, hashAuthToken(op.pipelineInfo.AuthToken))
		return false
	case rcPipelineVersion != strconv.FormatUint(op.pipelineInfo.Version, 10):
		log.Errorf("PPS master: pipeline version in %q looks stale %s != %d",
			op.pipelineInfo.Pipeline.Name, rcPipelineVersion, op.pipelineInfo.Version)
		return false
	case rcPachVersion != version.PrettyVersion():
		log.Errorf("PPS master: %q is using stale pachd v%s != current v%s",
			op.pipelineInfo.Pipeline.Name, rcPachVersion, version.PrettyVersion())
		return false
	case expectedName != "" && rcName != expectedName:
		log.Errorf("PPS master: %q has an unexpected (likely stale) name %q != %q",
			op.pipelineInfo.Pipeline.Name, rcName, expectedName)
	}
	return true
}

// setPipelineState set's op's state in the collection to 'state'. This will trigger a
// collection watch event and cause step() to eventually run again.
func (op *pipelineOp) setPipelineState(ctx context.Context, state pps.PipelineState, reason string) error {
	if err := setPipelineState(ctx, op.env.DB, op.pipelines, op.pipelineInfo.SpecCommit, state, reason); err != nil {
		// don't bother failing if we can't set the state
		return stepError{
			error: errors.Wrapf(err, "could not set pipeline state to %v"+
				"(you may need to restart pachd to un-stick the pipeline)", state),
			retry: true,
		}
	}
	return nil
}

// createPipelineResources creates the RC and any services for op's pipeline.
func (op *pipelineOp) createPipelineResources(ctx context.Context) error {
	log.Infof("PPS master: creating resources for pipeline %q", op.pipelineInfo.Pipeline.Name)
	if err := op.createWorkerSvcAndRc(ctx, op.pipelineInfo); err != nil {
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

// startPipelineMonitor spawns a monitorPipeline() goro for this pipeline (if
// one doesn't exist already), which manages standby and cron inputs, and
// updates the the pipeline state.
// Note: this is called by every run through step(), so must be idempotent
func (op *pipelineOp) startPipelineMonitor() {
	// since *op.pipelineInfo may be modified on a Bump, and startMonitor() passes its
	// input pipelineInfo to a monitor goroutine, send a pointer to a copy of op.pipelineInfo to
	// avoid a race condition
	pi := *op.pipelineInfo
	if op.monitorCancel == nil {
		op.monitorCancel = op.startMonitor(op.opCtx, &pi)
	}
	op.pipelineInfo.Details.WorkerRc = op.rc.ObjectMeta.Name
}

func (op *pipelineOp) startCrashingPipelineMonitor() {
	// since *op.pipelineInfo may be modified on a Bump, and startMonitor() passes its
	// input pipelineInfo to a monitor goroutine, send a pointer to a copy of op.pipelineInfo to
	// avoid a race condition
	pi := *op.pipelineInfo
	if op.crashingMonitorCancel == nil {
		op.crashingMonitorCancel = op.startCrashingMonitor(op.opCtx, &pi)
	}
}

func (op *pipelineOp) stopPipelineMonitor() {
	if op.monitorCancel != nil {
		op.monitorCancel()
		op.monitorCancel = nil
	}
}

func (op *pipelineOp) stopCrashingPipelineMonitor() {
	if op.crashingMonitorCancel != nil {
		op.crashingMonitorCancel()
		op.crashingMonitorCancel = nil
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
// in the case where op's pipeline is already in FAILURE. If it returns an error in
// that case, the pps master will log the error and move on to the next pipeline
// event. This pipeline's output commits will stay open until another watch
// event arrives for the pipeline and finishPipelineOutputCommits is retried.
func (op *pipelineOp) finishPipelineOutputCommits(ctx context.Context) (retErr error) {
	log.Infof("PPS master: finishing output commits for pipeline %q", op.pipelineInfo.Pipeline.Name)

	pachClient := op.env.GetPachClient(ctx)
	if span, _ctx := tracing.AddSpanToAnyExisting(ctx,
		"/pps.Master/FinishPipelineOutputCommits", "pipeline", op.pipelineInfo.Pipeline.Name); span != nil {
		pachClient = pachClient.WithCtx(_ctx) // copy span back into pachClient
		defer func() {
			tracing.TagAnySpan(span, "err", fmt.Sprintf("%v", retErr))
			tracing.FinishAnySpan(span)
		}()
	}
	pachClient.SetAuthToken(op.pipelineInfo.AuthToken)

	if err := pachClient.ListCommitF(client.NewRepo(op.pipelineInfo.Pipeline.Name), client.NewCommit(op.pipelineInfo.Pipeline.Name, op.pipelineInfo.Details.OutputBranch, ""), nil, 0, false, func(commitInfo *pfs.CommitInfo) error {
		return pachClient.StopJob(op.pipelineInfo.Pipeline.Name, commitInfo.Commit.ID)
	}); err != nil {
		if errutil.IsNotFoundError(err) {
			return nil // already deleted
		}
		return errors.Wrapf(err, "could not finish output commits of pipeline %q", op.pipelineInfo.Pipeline.Name)
	}
	return nil
}

// updateRC is a helper for {scaleUp,scaleDown}Pipeline. It includes all of the
// logic for writing an updated RC spec to kubernetes, and updating/retrying if
// k8s rejects the write. It presents a strange API, since the the RC being
// updated is already available to the caller in op.rc, but update() may be
// called muliple times if the k8s write fails. It may be helpful to think of
// the rc passed to update() as mutable, while op.rc is immutable.
func (op *pipelineOp) updateRC(update func(rc *v1.ReplicationController)) error {
	rc := op.env.KubeClient.CoreV1().ReplicationControllers(op.namespace)

	newRC := *op.rc
	// Apply op's update to rc
	update(&newRC)
	// write updated RC to k8s
	if _, err := rc.Update(&newRC); err != nil {
		return newRetriableError(err, "error updating RC")
	}
	return nil
}

// scaleUpPipeline edits the RC associated with op's pipeline & spins up the
// configured number of workers.
func (op *pipelineOp) scaleUpPipeline(ctx context.Context) (retErr error) {
	log.Debugf("PPS master: ensuring correct k8s resources for %q", op.pipelineInfo.Pipeline.Name)
	span, _ := tracing.AddSpanToAnyExisting(ctx,
		"/pps.Master/ScaleUpPipeline", "pipeline", op.pipelineInfo.Pipeline.Name)
	defer func() {
		if retErr != nil {
			log.Errorf("PPS master: error scaling up: %v", retErr)
		}
		tracing.TagAnySpan(span, "err", retErr)
		tracing.FinishAnySpan(span)
	}()

	// compute target pipeline parallelism
	parallelism := uint64(1)
	if op.pipelineInfo.Details.ParallelismSpec != nil {
		parallelism = op.pipelineInfo.Details.ParallelismSpec.Constant
	}

	// update pipeline RC
	return op.updateRC(func(rc *v1.ReplicationController) {
		if rc.Spec.Replicas != nil && *op.rc.Spec.Replicas > 0 {
			return // prior attempt succeeded
		}
		rc.Spec.Replicas = new(int32)
		if op.pipelineInfo.Details.Autoscaling {
			*rc.Spec.Replicas = 1
		} else {
			*rc.Spec.Replicas = int32(parallelism)
		}
	})
}

// scaleDownPipeline edits the RC associated with op's pipeline & spins down the
// configured number of workers.
func (op *pipelineOp) scaleDownPipeline(ctx context.Context) (retErr error) {
	log.Infof("PPS master: scaling down workers for %q", op.pipelineInfo.Pipeline.Name)
	span, _ := tracing.AddSpanToAnyExisting(ctx,
		"/pps.Master/ScaleDownPipeline", "pipeline", op.pipelineInfo.Pipeline.Name)
	defer func() {
		if retErr != nil {
			log.Errorf("PPS master: error scaling down: %v", retErr)
		}
		tracing.TagAnySpan(span, "err", retErr)
		tracing.FinishAnySpan(span)
	}()

	return op.updateRC(func(rc *v1.ReplicationController) {
		if rc.Spec.Replicas != nil && *op.rc.Spec.Replicas == 0 {
			return // prior attempt succeeded
		}
		rc.Spec.Replicas = &zero
	})
}

// restartPipeline updates the RC/service associated with op's pipeline, and
// then sets its state to RESTARTING. Note that restartPipeline only deletes
// op.rc if it's stale--a prior bug was that it would delete all of op's
// resources, and then get stuck in a loop deleting and recreating op's RC if
// the cluster was busy and the RC was taking too long to start.
//
// restartPipeline is an error-handling
// codepath, so it's guaranteed to return an error (typically wrapping 'reason',
// though if the restart process fails that error will take precendence) so that
// callers can use it like so:
//
// if errorState {
//   return op.restartPipeline("entered error state")
// }
func (op *pipelineOp) restartPipeline(ctx context.Context, reason string) error {
	if op.rc != nil && !op.rcIsFresh() {
		// delete old RC, monitorPipeline goro, and worker service
		if err := op.deletePipelineResources(); err != nil {
			return newRetriableError(err, "error deleting resources for restart")
		}
	}
	// create up-to-date RC
	if err := op.createPipelineResources(ctx); err != nil {
		return errors.Wrap(err, "error creating resources for restart")
	}
	if err := op.setPipelineState(ctx, pps.PipelineState_PIPELINE_RESTARTING, ""); err != nil {
		return errors.Wrap(err, "error restarting pipeline")
	}

	return errors.Errorf("restarting pipeline %q: %s", op.pipelineInfo.Pipeline.Name, reason)
}

// deletePipelineResources deletes the RC and services associated with op's
// pipeline. It doesn't return a stepError, leaving retry behavior to the caller
func (op *pipelineOp) deletePipelineResources() (retErr error) {
	log.Infof("PPS master: deleting resources for pipeline %q", op.pipeline)
	span, ctx := tracing.AddSpanToAnyExisting(op.opCtx,
		"/pps.Master/DeletePipelineResources", "pipeline", op.pipeline)
	defer func() {
		tracing.TagAnySpan(ctx, "err", retErr)
		tracing.FinishAnySpan(span)
	}()

	// Cancel any running monitorPipeline call
	op.stopPipelineMonitor()
	// Same for cancelCrashingMonitor
	op.stopCrashingPipelineMonitor()

	kubeClient := op.env.KubeClient
	namespace := op.namespace

	// Delete any services associated with op.pipeline
	selector := fmt.Sprintf("%s=%s", pipelineNameLabel, op.pipeline)
	opts := &metav1.DeleteOptions{
		OrphanDependents: &falseVal,
	}
	services, err := kubeClient.CoreV1().Services(namespace).List(metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return errors.Wrapf(err, "could not list services")
	}
	for _, service := range services.Items {
		if err := kubeClient.CoreV1().Services(namespace).Delete(service.Name, opts); err != nil {
			if !errutil.IsNotFoundError(err) {
				return errors.Wrapf(err, "could not delete service %q", service.Name)
			}
		}
	}

	// Delete any secrets associated with op.pipeline
	secrets, err := kubeClient.CoreV1().Secrets(namespace).List(metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return errors.Wrapf(err, "could not list secrets")
	}
	for _, secret := range secrets.Items {
		if err := kubeClient.CoreV1().Secrets(namespace).Delete(secret.Name, opts); err != nil {
			if !errutil.IsNotFoundError(err) {
				return errors.Wrapf(err, "could not delete secret %q", secret.Name)
			}
		}
	}

	// Finally, delete op.pipeline's RC, which will cause pollPipelines to stop
	// polling it.
	rcs, err := kubeClient.CoreV1().ReplicationControllers(namespace).List(metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return errors.Wrapf(err, "could not list RCs")
	}
	for _, rc := range rcs.Items {
		if err := kubeClient.CoreV1().ReplicationControllers(namespace).Delete(rc.Name, opts); err != nil {
			if !errutil.IsNotFoundError(err) {
				return errors.Wrapf(err, "could not delete RC %q", rc.Name)
			}
		}
	}

	return nil
}

func (op *pipelineOp) setPipelineFailure(ctx context.Context, reason string) error {
	return op.setPipelineState(ctx, pps.PipelineState_PIPELINE_FAILURE, reason)
}
