package server

import (
	"context"
	"fmt"
	"path"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/dlock"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
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

	// channel through which pipeline events are passed
	eventCh     chan *pipelineEvent
	lastUpdates map[string]time.Time
}

// The master process is responsible for creating/deleting workers as
// pipelines are created/removed.
func (a *apiServer) master() {
	m := &ppsMaster{
		a:                      a,
		monitorCancels:         make(map[string]func()),
		crashingMonitorCancels: make(map[string]func()),
		lastUpdates:            make(map[string]time.Time),
	}

	masterLock := dlock.NewDLock(a.env.GetEtcdClient(), path.Join(a.etcdPrefix, masterLockPath))
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
		m.run()
		return errors.Wrapf(ctx.Err(), "ppsMaster.Run() exited unexpectedly")
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		log.Errorf("PPS master: error running the master process: %v; retrying in %v", err, d)
		return nil
	})
	panic("internal error: PPS master has somehow exited. Restarting pod...")
}

func (a *apiServer) setPipelineFailure(ctx context.Context, specCommit *pfs.Commit, reason string) error {
	return a.setPipelineState(ctx, specCommit, pps.PipelineState_PIPELINE_FAILURE, reason)
}

func (a *apiServer) setPipelineCrashing(ctx context.Context, specCommit *pfs.Commit, reason string) error {
	return a.setPipelineState(ctx, specCommit, pps.PipelineState_PIPELINE_CRASHING, reason)
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
			switch e.eventType {
			case writeEv:
				if !e.timestamp.IsZero() && e.timestamp.Before(m.lastUpdates[e.pipeline]) {
					// we've stepped the pipeline since this event occurred
					// NOTE: this will ignore updates before 1970
					continue
				}
				if err := m.attemptStep(m.masterCtx, e); err != nil {
					log.Errorf("PPS master: %v", err)
				}
			case deleteEv:
				delete(m.lastUpdates, e.pipeline)
				// TODO(msteffen) trace this call
				if err := m.deletePipelineResources(e.pipeline); err != nil {
					log.Errorf("PPS master: could not delete resources for pipeline %q: %v",
						e.pipeline, err)
				}
			}
		case <-m.masterCtx.Done():
			break eventLoop
		}
	}
}

// attemptStep calls step for the given pipeline in a backoff loop.
// When it encounters a stepError, returned by most pipeline controller helper
// functions, it uses the fields to decide whether to retry the step or,
// if the retries have been exhausted, fail the pipeline.
//
// Other errors are simply logged and ignored, assuming that some future polling
// of the pipeline will succeed.
func (m *ppsMaster) attemptStep(ctx context.Context, e *pipelineEvent) error {
	var errCount int
	var stepErr stepError
	var startTime time.Time
	err := backoff.RetryNotify(func() error {
		startTime = time.Now()
		// Create/Modify/Delete pipeline resources as needed per new state
		return m.step(e.pipeline, e.timestamp)
	}, backoff.NewExponentialBackOff(), func(err error, d time.Duration) error {
		errCount++
		if errors.As(err, &stepErr) {
			if stepErr.retry && errCount < maxErrCount {
				log.Errorf("PPS master: error updating resources for pipeline %q: %v; retrying in %v",
					e.pipeline, err, d)
				return nil
			}
		}
		return errors.Wrapf(err, "could not update resource for pipeline %q", e.pipeline)
	})
	if err == nil && !startTime.IsZero() {
		// can write to the map as pps master does not use concurrency

		// add a little grace period for clock disagreements
		m.lastUpdates[e.pipeline] = startTime.Add(-2 * time.Second)
	}

	// we've given up on the step, check if the error indicated that the pipeline should fail
	if err != nil && errors.As(err, &stepErr) && stepErr.failPipeline {
		specCommit, specErr := m.a.findPipelineSpecCommit(ctx, e.pipeline)
		if specErr != nil {
			return errors.Wrapf(specErr, "error failing pipeline %q (%v)", e.pipeline, err)
		}
		failError := m.a.setPipelineFailure(ctx, specCommit, fmt.Sprintf(
			"could not update resources after %d attempts: %v", errCount, err))
		if failError != nil {
			return errors.Wrapf(failError, "error failing pipeline %q (%v)", e.pipeline, err)
		}
		return errors.Wrapf(err, "failing pipeline %q", e.pipeline)
	}
	return err
}

func (m *ppsMaster) deletePipelineResources(pipelineName string) (retErr error) {
	log.Infof("PPS master: deleting resources for pipeline %q", pipelineName)
	span, ctx := tracing.AddSpanToAnyExisting(m.masterCtx,
		"/pps.Master/DeletePipelineResources", "pipeline", pipelineName)
	defer func() {
		tracing.TagAnySpan(ctx, "err", retErr)
		tracing.FinishAnySpan(span)
	}()

	// Cancel any running monitorPipeline call
	m.cancelMonitor(pipelineName)
	// Same for cancelCrashingMonitor
	m.cancelCrashingMonitor(pipelineName)

	kubeClient := m.a.env.GetKubeClient()
	namespace := m.a.namespace

	// Delete any services associated with op.pipeline
	selector := fmt.Sprintf("%s=%s", pipelineNameLabel, pipelineName)
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

// setPipelineState is a PPS-master-specific helper that wraps
// ppsutil.SetPipelineState in a trace
func (a *apiServer) setPipelineState(ctx context.Context, specCommit *pfs.Commit, state pps.PipelineState, reason string) (retErr error) {
	span, ctx := tracing.AddSpanToAnyExisting(ctx,
		"/pps.Master/SetPipelineState", "pipeline", specCommit.Branch.Repo.Name, "new-state", state)
	defer func() {
		tracing.TagAnySpan(span, "err", retErr)
		tracing.FinishAnySpan(span)
	}()
	return ppsutil.SetPipelineState(ctx, a.env.GetDBClient(), a.pipelines,
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
	return ppsutil.SetPipelineState(ctx, a.env.GetDBClient(), a.pipelines,
		specCommit, from, to, reason)
}
