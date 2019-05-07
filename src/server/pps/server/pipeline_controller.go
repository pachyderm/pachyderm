package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/pachyderm/pachyderm/src/client"
	"strings"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/tracing"
	"github.com/pachyderm/pachyderm/src/client/pkg/tracing/extended"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsutil"
)

const maxErrCount = 3 // gives all retried operations ~4.5s total to finish

// pipelineOp contains all of the relevent current state for a pipeline. It's
// used by step() to take any necessary actions
type pipelineOp struct {
	apiServer    *apiServer
	pachClient   *client.APIClient
	ptr          *pps.EtcdPipelineInfo
	name         string // also in pipelineInfo, but that may not be set initially
	pipelineInfo *pps.PipelineInfo
	rc           *v1.ReplicationController
}

var (
	errRCNotFound = errors.New("RC not found")
	errTooManyRCs = errors.New("multiple RCs found for pipeline")
	errStaleRC    = errors.New("RC is doesn't match pipeline version")
)

// step takes 'ptr', a newly-changed pipeline pointer in etcd, and
// 1. retrieves its full pipeline spec and RC
// 2. makes whatever changes are needed to bring the RC in line with the (new) spec
// 3. updates 'ptr', if needed, to reflect the action it just took
func (a *apiServer) step(pachClient *client.APIClient, pipelineName string, keyVer, keyRev int64) error {
	// Handle tracing (pipelineRestarted is used to maybe delete trace)
	log.Infof("PPS master: processing event for %q", pipelineName)

	// Retrieve pipelineInfo from the spec repo
	op, err := a.newPipelineOp(pachClient, pipelineName)
	if err != nil {
		return a.setPipelineFailure(pachClient.Ctx(), pipelineName,
			fmt.Sprintf("couldn't initialize pipeline op: %v", err))
	}
	span, ctx := extended.AddPipelineSpanToAnyTrace(pachClient.Ctx(),
		a.etcdClient, pipelineName, "/pps.Master/ProcessPipelineUpdate",
		"key-version", keyVer,
		"mod-revision", keyRev,
		"state", op.ptr.State.String(),
		"spec-commit", op.ptr.SpecCommit,
	)
	defer tracing.FinishAnySpan(span)
	if span != nil {
		pachClient = pachClient.WithCtx(ctx)
	}

	// Take whatever actions are needed
	switch op.ptr.State {
	case pps.PipelineState_PIPELINE_STARTING, pps.PipelineState_PIPELINE_RESTARTING:
		if op.rc != nil && !op.rcIsFresh() {
			// old RC is not down yet
			log.Errorf("PPS master: restarting %q as it has an out-of-date RC", op.name)
			return op.restartPipeline()
		} else if op.rc == nil {
			// default: old RC (if any) is down but new RC is not up yet
			if err := op.createPipelineResources(); err != nil {
				return err
			}
		}
		// trigger another event--once pipeline is RUNNING, step() will scale it up
		if op.pipelineInfo.Stopped {
			if err := op.setPipelineState(pps.PipelineState_PIPELINE_PAUSED); err != nil {
				return err
			}
		} else {
			if err := op.setPipelineState(pps.PipelineState_PIPELINE_RUNNING); err != nil {
				return err
			}
		}
	case pps.PipelineState_PIPELINE_RUNNING:
		if !op.rcIsFresh() {
			return op.restartPipeline()
		}
		op.startPipelineMonitor()

		if op.pipelineInfo.Stopped {
			// StopPipeline has been called, but pipeline hasn't been paused yet
			if err := op.scaleDownPipeline(); err != nil {
				return err
			}
			return op.setPipelineState(pps.PipelineState_PIPELINE_PAUSED)
		}
		// default: scale up if pipeline start hasn't propagated to etcd yet
		// Note: mostly this should do nothing, as this runs several times per job
		if err := op.scaleUpPipeline(); err != nil {
			return err
		}
	case pps.PipelineState_PIPELINE_STANDBY, pps.PipelineState_PIPELINE_PAUSED:
		if !op.rcIsFresh() {
			log.Errorf("PPS master: restarting %q as it has no RC", op.name)
			if err := op.restartPipeline(); err != nil {
				return err
			}
			return nil
		}
		op.startPipelineMonitor()

		if op.ptr.State == pps.PipelineState_PIPELINE_PAUSED && !op.pipelineInfo.Stopped {
			// StartPipeline has been called, but pipeline hasn't been started yet
			if err := op.scaleUpPipeline(); err != nil {
				return err
			}
			return op.setPipelineState(pps.PipelineState_PIPELINE_RUNNING)
		}
		// default: scale down if pause/standby hasn't propagated to etcd yet
		if err := op.scaleDownPipeline(); err != nil {
			return err
		}
	case pps.PipelineState_PIPELINE_FAILURE:
		// pipeline fails if docker image isn't found
		if err := op.finishPipelineOutputCommits(); err != nil {
			return err
		}
		return op.deletePipelineResources()
	}
	return nil
}

func (a *apiServer) newPipelineOp(pachClient *client.APIClient, pipelineName string) (*pipelineOp, error) {
	op := &pipelineOp{
		apiServer:  a,
		pachClient: pachClient,
		ptr:        &pps.EtcdPipelineInfo{},
		name:       pipelineName,
	}
	// get latest EtcdPipelineInfo (events can pile up, so that the current state
	// doesn't match the event being processed)
	if err := a.pipelines.ReadOnly(pachClient.Ctx()).Get(pipelineName, op.ptr); err != nil {
		return nil, fmt.Errorf("could not retrieve etcd pipeline info for %q: %v", pipelineName, err)
	}
	// set op.pipelineInfo
	if err := op.getPipelineInfo(); err != nil {
		return nil, err
	}
	// set op.rc
	if err := op.getRC(); err != nil && err != errRCNotFound {
		return nil, fmt.Errorf("error reading RC for %q: %v", op.name, err)
	}
	return op, nil
}

// getPipelineInfo reads the pipelineInfo associated with 'op's pipeline. This
// should be one of the first calls made on 'op', as most other methods assume
// that op.pipelineInfo is set
func (op *pipelineOp) getPipelineInfo() error {
	var errCount int
	if err := backoff.RetryNotify(func() error {
		return op.apiServer.sudo(op.pachClient, func(superUserClient *client.APIClient) error {
			var err error
			op.pipelineInfo, err = ppsutil.GetPipelineInfo(superUserClient, op.ptr)
			return err
		})
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		if errCount++; errCount >= maxErrCount {
			return fmt.Errorf("error retrieving spec for %q after %d attempts: %v",
				op.name, maxErrCount, err)
		}
		log.Errorf("PPS master: error retrieving spec for %q: %v; retrying in %v", op.name, err, d)
		return nil
	}); err != nil {
		// don't necessarily restart PPS master, which would not fix the problem
		return op.setPipelineFailure(fmt.Sprintf("pipeline spec commit could not be read: %v", err))
	}
	return nil
}

// getRC reads the RC associated with 'op's pipeline. op.pipelineInfo must be
// set already
func (op *pipelineOp) getRC() (retErr error) {
	defer func() {
		if retErr != nil {
			log.Infof("PPS master: could not retrieve RC for %q: %v", op.name, retErr)
		}
	}()

	selector := fmt.Sprintf("%s=%s", pipelineNameLabel, op.name)
	rcName := ppsutil.PipelineRcName(op.name, op.pipelineInfo.Version)
	kubeClient := op.apiServer.kubeClient
	namespace := op.apiServer.namespace
	span, _ := tracing.AddSpanToAnyExisting(op.pachClient.Ctx(),
		"/pps.Master/GetRC", "pipeline", op.name)
	defer func(span opentracing.Span) {
		if span != nil {
			span.SetTag("err", fmt.Sprintf("%v", retErr))
		}
		tracing.FinishAnySpan(span)
	}(span)

	var errCount int
	backoff.RetryNotify(func() error {
		// List all RCs, so stale RCs from old pipelines are noticed and deleted
		rcs, err := kubeClient.CoreV1().ReplicationControllers(namespace).List(metav1.ListOptions{LabelSelector: selector})
		if err != nil && !isNotFoundErr(err) {
			return err
		}
		if len(rcs.Items) == 0 {
			return errRCNotFound
		} else if len(rcs.Items) > 1 {
			return errTooManyRCs
		} else if rcs.Items[0].ObjectMeta.Name != rcName {
			return errStaleRC
		}
		op.rc = &rcs.Items[0]
		return nil
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		if err != errRCNotFound {
			return err
		} else if errCount++; errCount >= maxErrCount {
			return err
		}
		log.Errorf("PPS master: error retrieving RC for %q: %v; retrying in %v", op.name, err, d)
		return nil
	})
	// definitely don't restart master--sometimes we expect to not find an RC!
	// (e.g. when pipeline is in STARTING)
	return nil
}

// rcIsFresh returns a boolean indicating whether op.rc has the right labels
// corresponding to op.ptr. If this returns false, it likely means the current
// RC is using e.g. an old spec commit or something.
func (op *pipelineOp) rcIsFresh() bool {
	if op.rc == nil {
		log.Errorf("PPS master: RC for %q is nil", op.name)
		return false
	}
	rcPachVersion := op.rc.ObjectMeta.Annotations[pachVersionAnnotation]
	rcAuthTokenHash := op.rc.ObjectMeta.Annotations[hashedAuthTokenAnnotation]
	rcSpecCommit := op.rc.ObjectMeta.Annotations[specCommitAnnotation]
	switch {
	case rcAuthTokenHash != hashAuthToken(op.ptr.AuthToken):
		log.Infof("PPS master: auth token in %q is stale %s != %s",
			op.name, rcAuthTokenHash, hashAuthToken(op.ptr.AuthToken))
		return false
	case rcSpecCommit != op.ptr.SpecCommit.ID:
		log.Infof("PPS master: spec commit in %q looks stale %s != %s",
			op.name, rcSpecCommit, op.ptr.SpecCommit.ID)
		return false
	case rcPachVersion != version.PrettyVersion():
		log.Infof("PPS master: %q is using stale pachd v%s != current v%s",
			op.name, rcPachVersion, version.PrettyVersion())
		return false
	}
	return true
}

func (op *pipelineOp) setPipelineState(state pps.PipelineState) error {
	return op.apiServer.setPipelineState(op.pachClient, op.pipelineInfo, state, "")
}

func (op *pipelineOp) restartPipeline() error {
	log.Infof("PPS master: restarting pipeline %q", op.name)
	var errCount int
	if err := backoff.RetryNotify(func() error {
		if err := op.deletePipelineResources(); err != nil {
			return err
		}
		return op.setPipelineState(pps.PipelineState_PIPELINE_RESTARTING)
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		if errCount++; errCount >= maxErrCount {
			return fmt.Errorf("could not restart pipeline %q after %d attempts: %v",
				op.name, maxErrCount, err)
		}
		log.Errorf("PPS master: error restarting pipeline %q: %v; retrying in %v", op.name, err, d)
		return nil
	}); err != nil {
		// don't necessarily restart PPS master, which would not fix the problem
		return op.setPipelineFailure(fmt.Sprintf("error during restart: %v", err))
	}
	return nil
}

func (op *pipelineOp) setPipelineFailure(reason string) error {
	return op.apiServer.setPipelineFailure(op.pachClient.Ctx(),
		op.name, reason)
}

func (op *pipelineOp) createPipelineResources() error {
	log.Infof("PPS master: creating resources for pipeline %q", op.name)
	var errCount int
	if err := backoff.RetryNotify(func() error {
		return op.apiServer.createWorkerSvcAndRc(op.pachClient.Ctx(), op.ptr, op.pipelineInfo)
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		if errCount++; errCount >= maxErrCount {
			return fmt.Errorf("could not create resources for pipeline %q after %d attempts: %v",
				op.name, maxErrCount, err)
		}
		log.Errorf("PPS master: error creating resources for pipeline %q: %v; retrying in %v",
			op.name, err, d)
		return nil
	}); err != nil {
		// don't necessarily restart PPS master, which would not fix the problem
		return op.setPipelineFailure(fmt.Sprintf("error creating resources: %v", err.Error()))
	}
	return nil
}

// startPipelineMonitor spawns a monitorPipeline() goro for this pipeline (if
// one doesn't exist already), which manages standby and cron inputs, and
// updates the the pipeline state.
// Note: this is called by every run through step(), so must be idempotent
func (op *pipelineOp) startPipelineMonitor() {
	op.apiServer.monitorCancelsMu.Lock()
	defer op.apiServer.monitorCancelsMu.Unlock()
	if _, ok := op.apiServer.monitorCancels[op.name]; !ok {
		// use context.Background because we expect this goro to run for the rest of
		// pachd's lifetime
		ctx, cancel := context.WithCancel(context.Background())
		op.apiServer.monitorCancels[op.name] = cancel
		pachClient := op.apiServer.pachClient.WithCtx(ctx)

		go op.apiServer.sudo(pachClient, func(superUserClient *client.APIClient) error {
			op.apiServer.monitorPipeline(superUserClient, op.pipelineInfo)
			return nil
		})
	}
}

// finishPipelineOutputCommits finishes any output commits of
// 'pipelineInfo.Pipeline' with an empty tree.
// TODO(msteffen) Note that if the pipeline has any jobs (which can happen if
// the user manually deletes the pipeline's RC, failing the pipeline, after it
// has created jobs) those will not be updated, but they should be FAILED
func (op *pipelineOp) finishPipelineOutputCommits() (retErr error) {
	log.Infof("PPS master: finishing output commits for pipeline %q", op.name)
	span, _ctx := tracing.AddSpanToAnyExisting(op.pachClient.Ctx(),
		"/pps.Master/FinishPipelineOutputCommits", "pipeline", op.name)
	pachClient := op.pachClient.WithCtx(_ctx) // copy span back into pachClient
	defer func(span opentracing.Span) {
		if span != nil {
			span.SetTag("err", fmt.Sprintf("%v", retErr))
		}
		tracing.FinishAnySpan(span)
	}(span)

	return op.apiServer.sudo(pachClient, func(superUserClient *client.APIClient) error {
		commitInfos, err := superUserClient.ListCommit(op.name, op.pipelineInfo.OutputBranch, "", 0)
		if err != nil {
			return fmt.Errorf("could not list output commits of %q to finish them: %v", op.name, err)
		}

		var finishCommitErr error
		for _, ci := range commitInfos {
			if ci.Finished != nil {
				continue // nothing needs to be done
			}
			if _, err := superUserClient.PfsAPIClient.FinishCommit(superUserClient.Ctx(),
				&pfs.FinishCommitRequest{
					Commit: client.NewCommit(op.name, ci.Commit.ID),
					Empty:  true,
				}); err != nil && finishCommitErr == nil {
				finishCommitErr = err
			}
		}
		return finishCommitErr
	})
}

func (op *pipelineOp) deletePipelineResources() (retErr error) {
	return op.apiServer.deletePipelineResources(op.pachClient.Ctx(), op.name)
}

func (op *pipelineOp) scaleUpPipeline() (retErr error) {
	span, _ := tracing.AddSpanToAnyExisting(op.pachClient.Ctx(),
		"/pps.Master/ScaleUpPipeline", "pipeline", op.name)
	defer func(span opentracing.Span) {
		if span != nil {
			span.SetTag("err", fmt.Sprintf("%v", retErr))
		}
		tracing.FinishAnySpan(span)
	}(span)

	kubeClient := op.apiServer.kubeClient
	namespace := op.apiServer.namespace
	rc := kubeClient.CoreV1().ReplicationControllers(namespace)

	// compute target pipeline parallelism
	parallelism, err := ppsutil.GetExpectedNumWorkers(kubeClient, op.pipelineInfo.ParallelismSpec)
	if err != nil {
		log.Errorf("PPS master: error getting number of workers, default to 1 worker: %v", err)
		parallelism = 1
	}
	parallelism32 := int32(parallelism)

	// don't fail pipeline just because it won't scale up. In this case, prefer
	// to restart PPS master
	var errCount int
	return backoff.RetryNotify(func() error {
		if op.rc.Spec.Replicas != nil && *op.rc.Spec.Replicas == parallelism32 {
			return nil // prior attempt succeeded
		}
		newRC := *op.rc
		newRC.Spec.Replicas = &parallelism32
		if _, err = rc.Update(&newRC); err != nil {
			return fmt.Errorf("could not update scaled-up pipeline %q RC: %v",
				op.name, err)
		}
		return nil
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		if strings.Contains(err.Error(), "try again") {
			// refresh RC--sometimes kubernetes complains that the RC is stale
			if err := op.getRC(); err != nil && err != errRCNotFound {
				return fmt.Errorf("error reading RC for %q: %v", op.name, err)
			}
		} else if errCount++; errCount >= maxErrCount {
			return fmt.Errorf("error scaling up %q after %d attempts: %v",
				op.name, maxErrCount, err)
		}
		log.Errorf("PPS master: error scaling down %q: %v; retrying in %v", op.name, err, d)
		return nil
	})
}

func (op *pipelineOp) scaleDownPipeline() (retErr error) {
	log.Infof("PPS master: scaling down workers for %q", op.name)
	span, _ := tracing.AddSpanToAnyExisting(op.pachClient.Ctx(),
		"/pps.Master/ScaleDownPipeline", "pipeline", op.name)
	defer func(span opentracing.Span) {
		if span != nil {
			span.SetTag("err", fmt.Sprintf("%v", retErr))
		}
		tracing.FinishAnySpan(span)
	}(span)

	kubeClient := op.apiServer.kubeClient
	namespace := op.apiServer.namespace
	rc := kubeClient.CoreV1().ReplicationControllers(namespace)
	var zero int32

	// don't fail pipeline just because it won't scale down. In this case, prefer
	// to restart PPS master
	var errCount int
	return backoff.RetryNotify(func() error {
		if op.rc.Spec.Replicas != nil && *op.rc.Spec.Replicas == 0 {
			return nil // prior attempt succeeded
		}
		newRC := *op.rc
		newRC.Spec.Replicas = &zero
		if _, err := rc.Update(&newRC); err != nil {
			return fmt.Errorf("could not update scaled-down pipeline %q RC: %v", op.name, err)
		}
		return nil
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		if strings.Contains(err.Error(), "try again") {
			// refresh RC--sometimes kubernetes complains that the RC is stale
			if err := op.getRC(); err != nil && err != errRCNotFound {
				return fmt.Errorf("error reading RC for %q: %v", op.name, err)
			}
		} else if errCount++; errCount >= maxErrCount {
			return fmt.Errorf("error scaling down %q after %d attempts: %v",
				op.name, maxErrCount, err)
		}
		log.Errorf("PPS master: error scaling down %q: %v; retrying in %v", op.name, err, d)
		return nil
	})
}
