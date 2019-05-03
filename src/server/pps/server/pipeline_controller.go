package server

import (
	"context"
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
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsutil"
)

const maxErrCount = 3

type pipelineOp struct {
	apiServer    *apiServer
	pachClient   *client.APIClient
	ptr          *pps.EtcdPipelineInfo
	name         string // also in pipelineInfo, but that may not be set initially
	pipelineInfo *pps.PipelineInfo
	rc           *v1.ReplicationController
}

// step takes 'ptr', a newly-changed pipeline pointer in etcd, and
// 1. converts it into the full pipeline spec and RC
// 2. makes whatever changes are needed to bring the RC in line with the (new) spec
// 3. updates 'ptr', if needed, to reflect the action it just took
func (a *apiServer) step(pachClient *client.APIClient, pipelineName string, ptr *pps.EtcdPipelineInfo) error {
	// Retrieve pipelineInfo from the spec repo
	op := &pipelineOp{
		apiServer:  a,
		pachClient: pachClient,
		ptr:        ptr,
		name:       pipelineName,
	}
	if ptr.SpecCommit == nil || ptr.SpecCommit.ID == "" {
		op.setPipelineFailure("no spec commit")
	}

	// set op.ppelineInfo
	if err := op.getPipelineInfo(); err != nil {
		return err
	}
	// set op.rc
	if err := op.getRC(); err != nil && !isNotFoundErr(err) {
		return fmt.Errorf("error reading RC for %q: %v", op.name, err)
	}

	// Take whatever actions are needed
	switch ptr.State {
	case pps.PipelineState_PIPELINE_STARTING, pps.PipelineState_PIPELINE_RESTARTING:
		if op.rc != nil {
			// overwrite etcdPipelineInfo with itself so that we re-process
			// the pipeline after the RC is gone
			return op.restartPipeline()
		}
		if err := op.createPipelineResources(); err != nil {
			return err
		}
		// trigger another event--once pipeline is RUNNING, step() will scale it up
		if err := op.setPipelineState(pps.PipelineState_PIPELINE_RUNNING); err != nil {
			return err
		}
	case pps.PipelineState_PIPELINE_RUNNING:
		specCommitB64, err := commitIDToB64(ptr.SpecCommit.ID)
		if err != nil {
			return err
		}
		if op.rc == nil ||
			op.rc.ObjectMeta.Labels[hashedAuthTokenLabel] != hashAuthToken(ptr.AuthToken) ||
			op.rc.ObjectMeta.Labels[specCommitLabel] != specCommitB64 ||
			op.rc.ObjectMeta.Labels[pachVersionLabel] != version.PrettyVersion() {
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
		// pipeline has been unpaused or taken out of standby--scale up
		if err := op.scaleUpPipeline(); err != nil {
			return err
		}
	case pps.PipelineState_PIPELINE_STANDBY, pps.PipelineState_PIPELINE_PAUSED:
		if op.rc == nil {
			if err := op.restartPipeline(); err != nil {
				return err
			}
			return nil
		}
		op.startPipelineMonitor()

		// scale down is pause/standby hasn't been propagated to etcd yet
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
func (op *pipelineOp) getRC() error {
	kubeClient := op.apiServer.kubeClient
	namespace := op.apiServer.namespace
	span, _ := tracing.AddSpanToAnyExisting(op.pachClient.Ctx(), "/kube.RC/Get",
		"pipeline", op.name)
	defer tracing.FinishAnySpan(span)
	var err error
	op.rc, err = kubeClient.CoreV1().ReplicationControllers(namespace).Get(
		ppsutil.PipelineRcName(op.name, op.pipelineInfo.Version),
		metav1.GetOptions{})
	if err != nil {
		op.rc = nil // unset; even with non-nil error Get() returns default RC
	}
	return err
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

	return a.sudo(pachClient, func(superUserClient *client.APIClient) error {
		commitInfos, err := superUserClient.ListCommit(pipelineName, pipelineInfo.OutputBranch, "", 0)
		if err != nil {
			return fmt.Errorf("could not list output commits of %q to finish them: %v", pipelineName, err)
		}

		var finishCommitErr error
		for _, ci := range commitInfos {
			if ci.Finished != nil {
				continue // nothing needs to be done
			}
			if _, err := superUserClient.PfsAPIClient.FinishCommit(superUserClient.Ctx(),
				&pfs.FinishCommitRequest{
					Commit: client.NewCommit(pipelineName, ci.Commit.ID),
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
			if err := op.getRC(); err != nil && !isNotFoundErr(err) {
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
			if err := op.getRC(); err != nil && !isNotFoundErr(err) {
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
