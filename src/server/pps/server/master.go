package server

import (
	"context"
	"fmt"
	"path"
	"time"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_watch "k8s.io/apimachinery/pkg/watch"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/tracing"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/dlock"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/watch"
)

const (
	masterLockPath = "_master_lock"
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

// The master process is responsible for creating/deleting workers as
// pipelines are created/removed.
func (a *apiServer) master() {
	masterLock := dlock.NewDLock(a.env.GetEtcdClient(), path.Join(a.etcdPrefix, masterLockPath))
	backoff.RetryNotify(func() error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		ctx, err := masterLock.Lock(ctx)
		if err != nil {
			return err
		}
		defer masterLock.Unlock(ctx)

		// Note: 'pachClient' is unauthenticated. This will use the PPS token (via
		// a.sudo()) to authenticate requests.
		pachClient := a.env.GetPachClient(ctx)
		kubeClient := a.env.GetKubeClient()

		log.Infof("PPS master: launching master process")

		// start pollPipelines in the background to regularly refresh pipelines
		func() {
			a.monitorCancelsMu.Lock()
			defer a.monitorCancelsMu.Unlock()
			a.pollCancel = startMonitorThread(pachClient, "pollPipelines",
				func(pachClient *client.APIClient) {
					a.pollPipelines(pachClient.Ctx())
				})
		}()

		// TODO(msteffen) request only keys, since pipeline_controller.go reads
		// fresh values for each event anyway
		pipelineWatcher, err := a.pipelines.ReadOnly(ctx).Watch()
		if err != nil {
			return errors.Wrapf(err, "error creating watch")
		}
		defer pipelineWatcher.Close()

		// watchChan will be nil if the Watch call below errors, this means
		// that we won't receive events from k8s and won't be able to detect
		// errors in pods. We could just return that error and retry but that
		// prevents pachyderm from creating pipelines when there's an issue
		// talking to k8s.
		var watchChan <-chan kube_watch.Event
		kubePipelineWatch, err := kubeClient.CoreV1().Pods(a.namespace).Watch(
			metav1.ListOptions{
				LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(
					map[string]string{
						"component": "worker",
					})),
				Watch: true,
			})
		if err != nil {
			log.Errorf("failed to watch kuburnetes pods: %v", err)
		} else {
			watchChan = kubePipelineWatch.ResultChan()
			defer kubePipelineWatch.Stop()
		}

		for {
			select {
			case event := <-pipelineWatcher.Watch():
				if event.Err != nil {
					return errors.Wrapf(event.Err, "event err")
				}
				switch event.Type {
				case watch.EventPut:
					pipeline := string(event.Key)
					// Create/Modify/Delete pipeline resources as needed per new state
					if err := a.step(pachClient, pipeline, event.Ver, event.Rev); err != nil {
						log.Errorf("PPS master: %v", err)
					}
				case watch.EventDelete:
					// TODO(msteffen) trace this call
					// This is also called by pollPipelines below, if it discovers
					// dangling monitorPipeline goroutines
					if err := a.deletePipelineResources(pachClient.Ctx(), string(event.Key)); err != nil {
						log.Errorf("PPS master: could not delete pipeline resources for %q: %v", err)
					}
				}
			case event := <-watchChan:
				// if we get an error we restart the watch, k8s watches seem to
				// sometimes get stuck in a loop returning events with Type =
				// "" we treat these as errors since otherwise we get an
				// endless stream of them and can't do anything.
				if event.Type == kube_watch.Error || event.Type == "" {
					if kubePipelineWatch != nil {
						kubePipelineWatch.Stop()
					}
					kubePipelineWatch, err = kubeClient.CoreV1().Pods(a.namespace).Watch(
						metav1.ListOptions{
							LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(
								map[string]string{
									"component": "worker",
								})),
							Watch: true,
						})
					if err != nil {
						log.Errorf("failed to watch kuburnetes pods: %v", err)
						watchChan = nil
					} else {
						watchChan = kubePipelineWatch.ResultChan()
						defer kubePipelineWatch.Stop()
					}
				}
				pod, ok := event.Object.(*v1.Pod)
				if !ok {
					continue
				}
				if pod.Status.Phase == v1.PodFailed {
					log.Errorf("pod failed because: %s", pod.Status.Message)
				}
				pipelineName := pod.ObjectMeta.Annotations["pipelineName"]
				for _, status := range pod.Status.ContainerStatuses {
					if status.State.Waiting != nil && failures[status.State.Waiting.Reason] {
						if err := a.setPipelineCrashing(pachClient.Ctx(), pipelineName, status.State.Waiting.Message); err != nil {
							return err
						}
					}
				}
				for _, condition := range pod.Status.Conditions {
					if condition.Type == v1.PodScheduled &&
						condition.Status != v1.ConditionTrue && failures[condition.Reason] {
						if err := a.setPipelineCrashing(pachClient.Ctx(), pipelineName, condition.Message); err != nil {
							return err
						}
					}
				}
			}
		}
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		// cancel all monitorPipeline and monitorCrashingPipeline goroutines.
		// Strictly speaking, this should be unnecessary, as the base context for
		// all monitor goros is cancelled by 'defer cancel()' at the beginning of
		// 'RetryNotify' above. However, these cancel calls also block until the
		// monitor goros exit, ensuring that a leftover goro won't interfere with a
		// subsequent iteration
		a.cancelAllMonitorsAndCrashingMonitors(nil)
		a.monitorCancelsMu.Lock()
		defer a.monitorCancelsMu.Unlock()
		if a.pollCancel != nil {
			a.pollCancel()
			a.pollCancel = nil
		}
		log.Errorf("PPS master: error running the master process: %v; retrying in %v", err, d)
		return nil
	})
	panic("internal error: PPS master has somehow exited. Restarting pod...")
}

func (a *apiServer) setPipelineFailure(ctx context.Context, pipelineName string, reason string) error {
	return a.setPipelineState(ctx, pipelineName, pps.PipelineState_PIPELINE_FAILURE, reason)
}

func (a *apiServer) setPipelineCrashing(ctx context.Context, pipelineName string, reason string) error {
	return a.setPipelineState(ctx, pipelineName, pps.PipelineState_PIPELINE_CRASHING, reason)
}

func (a *apiServer) deletePipelineResources(ctx context.Context, pipelineName string) (retErr error) {
	log.Infof("PPS master: deleting resources for pipeline %q", pipelineName)
	span, ctx := tracing.AddSpanToAnyExisting(ctx, //lint:ignore SA4006 ctx is unused, but better to have the right ctx in scope so people don't use the wrong one
		"/pps.Master/DeletePipelineResources", "pipeline", pipelineName)
	defer func() {
		tracing.TagAnySpan(span, "err", retErr)
		tracing.FinishAnySpan(span)
	}()

	// Cancel any running monitorPipeline call
	a.cancelMonitor(pipelineName)
	// Same for cancelCrashingMonitor
	a.cancelCrashingMonitor(pipelineName)

	kubeClient := a.env.GetKubeClient()
	// Delete any services associated with op.pipeline
	selector := fmt.Sprintf("%s=%s", pipelineNameLabel, pipelineName)
	opts := &metav1.DeleteOptions{
		OrphanDependents: &falseVal,
	}
	services, err := kubeClient.CoreV1().Services(a.namespace).List(metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return errors.Wrapf(err, "could not list services")
	}
	for _, service := range services.Items {
		if err := kubeClient.CoreV1().Services(a.namespace).Delete(service.Name, opts); err != nil {
			if !isNotFoundErr(err) {
				return errors.Wrapf(err, "could not delete service %q", service.Name)
			}
		}
	}
	rcs, err := kubeClient.CoreV1().ReplicationControllers(a.namespace).List(metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return errors.Wrapf(err, "could not list RCs")
	}
	for _, rc := range rcs.Items {
		if err := kubeClient.CoreV1().ReplicationControllers(a.namespace).Delete(rc.Name, opts); err != nil {
			if !isNotFoundErr(err) {
				return errors.Wrapf(err, "could not delete RC %q: %v", rc.Name)
			}
		}
	}
	return nil
}

// setPipelineState is a PPS-master-specific helper that wraps
// ppsutil.SetPipelineState in a trace
func (a *apiServer) setPipelineState(ctx context.Context, pipeline string, state pps.PipelineState, reason string) (retErr error) {
	span, ctx := tracing.AddSpanToAnyExisting(ctx,
		"/pps.Master/SetPipelineState", "pipeline", pipeline, "new-state", state)
	defer func() {
		tracing.TagAnySpan(span, "err", retErr)
		tracing.FinishAnySpan(span)
	}()
	return ppsutil.SetPipelineState(ctx, a.env.GetEtcdClient(), a.pipelines,
		pipeline, nil, state, reason)
}

// transitionPipelineState is similar to setPipelineState, except that it sets
// 'from' and logs a different trace
func (a *apiServer) transitionPipelineState(ctx context.Context, pipeline string, from []pps.PipelineState, to pps.PipelineState, reason string) (retErr error) {
	span, ctx := tracing.AddSpanToAnyExisting(ctx,
		"/pps.Master/TransitionPipelineState", "pipeline", pipeline,
		"from-state", from, "to-state", to)
	defer func() {
		tracing.TagAnySpan(span, "err", retErr)
		tracing.FinishAnySpan(span)
	}()
	return ppsutil.SetPipelineState(ctx, a.env.GetEtcdClient(), a.pipelines,
		pipeline, from, to, reason)
}

func (a *apiServer) pollPipelines(ctx context.Context) {
	if err := backoff.RetryUntilCancel(ctx, func() error {
		// 1. Get the set of pipelines currently in etcd, and generate an etcd event
		// for each one (to trigger the pipeline controller)
		etcdPipelines := map[string]bool{}
		pachClient := a.env.GetPachClient(ctx)
		// collect all pipelines in etcd
		if err := a.listPipelinePtr(pachClient, nil, 0,
			func(pipeline string, listPI *pps.EtcdPipelineInfo) error {
				log.Debugf("PPS master: polling pipeline %q", pipeline)
				etcdPipelines[pipeline] = true
				var curPI pps.EtcdPipelineInfo
				_, err := col.NewSTM(ctx, a.env.GetEtcdClient(), func(stm col.STM) error {
					// Read & immediately write
					return a.pipelines.ReadWrite(stm).Update(pipeline, &curPI, func() error { return nil })
				})
				if col.IsErrNotFound(err) {
					log.Warnf("%q polling conflicted with delete, will retry", pipeline)
				} else {
					log.Errorf("could not poll %q: %v", pipeline, err)
				}
				return nil // no error recovery to do here, just keep polling...
			}); err != nil {
			// listPipelinePtr results (etcdPipelines) are used by the next two
			// steps (RC cleanup and Monitor cleanup), so if that didn't work, sleep
			// and try again
			return errors.Wrap(err, "error polling pipelines")
		}

		// 2. Clean up any orphaned RCs (because we can't generate etcd delete
		// events for the master to process)
		kc := a.env.GetKubeClient().CoreV1().ReplicationControllers(a.env.Namespace)
		rcs, err := kc.List(metav1.ListOptions{
			LabelSelector: "suite=pachyderm,pipelineName",
		})
		if err != nil {
			return errors.Wrap(err, "error polling pipeline RCs")
		}
		for _, rc := range rcs.Items {
			pipeline, ok := rc.Labels["pipelineName"]
			if !ok {
				return errors.New("'pipelineName' label missing from rc " + rc.Name)
			}
			if !etcdPipelines[pipeline] {
				if err := a.deletePipelineResources(ctx, pipeline); err != nil {
					// log the error but don't return it, so that one broken RC doesn't
					// block pollPipelines
					log.Errorf("could not delete resources for %q: %v", pipeline, err)
				}
			}
		}

		// 3. Clean up any orphaned monitorPipeline and monitorCrashingPipeline
		// goros
		a.cancelAllMonitorsAndCrashingMonitors(etcdPipelines)
		return backoff.ErrContinue
	}, backoff.NewConstantBackOff(30*time.Second),
		backoff.NotifyContinue("pollPipelines"),
	); err != nil {
		if ctx.Err() == nil {
			panic("pollPipelines is exiting prematurely which should not happen; restarting pod...")
		}
	}
}
