package server

import (
	"time"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	kube_err "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_watch "k8s.io/apimachinery/pkg/watch"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
)

const pollBackoffTime = 2 * time.Second

// startPipelinePoller starts a new goroutine running pollPipelines
func (m *ppsMaster) startPipelinePoller() {
	m.pollPipelinesMu.Lock()
	defer m.pollPipelinesMu.Unlock()
	m.pollCancel = m.startMonitorThread("pollPipelines", m.pollPipelines)
}

func (m *ppsMaster) cancelPipelinePoller() {
	m.pollPipelinesMu.Lock()
	defer m.pollPipelinesMu.Unlock()
	if m.pollCancel != nil {
		m.pollCancel()
		m.pollCancel = nil
	}
}

// startPipelinePodsPoller starts a new goroutine running pollPipelinePods
func (m *ppsMaster) startPipelinePodsPoller() {
	m.pollPipelinesMu.Lock()
	defer m.pollPipelinesMu.Unlock()
	m.pollPodsCancel = m.startMonitorThread("pollPipelinePods", m.pollPipelinePods)
}

func (m *ppsMaster) cancelPipelinePodsPoller() {
	m.pollPipelinesMu.Lock()
	defer m.pollPipelinesMu.Unlock()
	if m.pollPodsCancel != nil {
		m.pollPodsCancel()
		m.pollPodsCancel = nil
	}
}

//////////////////////////////////////////////////////////////////////////////
//                     PollPipelines Definition                             //
// - As in monitor.go, functions below should not call functions above, to  //
// avoid reentrancy deadlock.                                               //
//////////////////////////////////////////////////////////////////////////////

func (m *ppsMaster) pollPipelines(pollClient *client.APIClient) {
	ctx := pollClient.Ctx()
	etcdPipelines := map[string]bool{}
	if err := backoff.RetryUntilCancel(ctx, func() error {
		if len(etcdPipelines) == 0 {
			// 1. Get the current set of pipeline RCs.
			//
			// We'll delete any RCs that don't correspond to a live pipeline after
			// querying etcd to determine the set of live pipelines, but we query k8s
			// first to avoid a race (if we were to query etcd first, and
			// CreatePipeline(foo) were to run between querying etcd and querying k8s,
			// then we might delete the RC for brand-new pipeline 'foo'). Even if we
			// do delete a live pipeline's RC, it'll be fixed in the next cycle)
			kc := m.a.env.GetKubeClient().CoreV1().ReplicationControllers(m.a.env.Namespace)
			rcs, err := kc.List(metav1.ListOptions{
				LabelSelector: "suite=pachyderm,pipelineName",
			})
			if err != nil {
				// No sensible error recovery here (e.g .if we can't reach k8s). We'll
				// keep going, and just won't delete any RCs this round.
				log.Errorf("error polling pipeline RCs: %v", err)
			}

			// 2. Replenish 'etcdPipelines' with the set of pipelines currently in
			// etcd. Note that there may be zero, and etcdPipelines may be empty
			if err := m.a.listPipelinePtr(pollClient, nil, 0,
				func(pipeline string, _ *pps.EtcdPipelineInfo) error {
					etcdPipelines[pipeline] = true
					return nil
				}); err != nil {
				// listPipelinePtr results (etcdPipelines) are used by all remaining
				// steps, so if that didn't work, start over and try again
				etcdPipelines = map[string]bool{}
				return errors.Wrap(err, "error polling pipelines")
			}

			// 3. Clean up any orphaned RCs (because we can't generate etcd delete
			// events for the master to process)
			if rcs != nil {
				for _, rc := range rcs.Items {
					pipeline, ok := rc.Labels["pipelineName"]
					if !ok {
						return errors.New("'pipelineName' label missing from rc " + rc.Name)
					}
					if !etcdPipelines[pipeline] {
						if err := m.deletePipelineResources(pipeline); err != nil {
							// log the error but don't return it, so that one broken RC doesn't
							// block pollPipelines
							log.Errorf("could not delete resources for %q: %v", pipeline, err)
						}
					}
				}
			}

			// 4. Likewise, clean up any orphaned monitorPipeline and
			// monitorCrashingPipeline goros. Note that this may delete a new
			// pipeline's monitorPipeline goro (if CreatePipeline(foo) runs between
			// 'listPipelinePtr' above, and here, then this may delete brand-new
			// pipeline 'foo's monitorPipeline goro).  However, the next run through
			// this loop will restore it, by generating an etcd event for 'foo', which
			// will cause the pipeline controller to restart monitorPipeline(foo).
			m.cancelAllMonitorsAndCrashingMonitors(etcdPipelines)

			// 5. Retry if there are no etcd pipelines to read/write
			if len(etcdPipelines) == 0 {
				return backoff.ErrContinue
			}
		}

		// Generate one etcd event for a pipeline (to trigger the pipeline
		// controller) and remove this pipeline from etcdPipelines. Always choose
		// the lexicographically smallest pipeline so that pipelines are always
		// traversed in the same order and the period between polls is stable across
		// all pipelines.
		var pipeline string
		for p := range etcdPipelines {
			if pipeline == "" || p < pipeline {
				pipeline = p
			}
		}

		// always rm 'pipeline', to advance loop
		delete(etcdPipelines, pipeline)

		// generate an etcd event for 'pipeline' by reading it & writing it back
		log.Debugf("PPS master: polling pipeline %q", pipeline)
		var curPI pps.EtcdPipelineInfo
		_, err := col.NewSTM(ctx, m.a.env.GetEtcdClient(), func(stm col.STM) error {
			return m.a.pipelines.ReadWrite(stm).Update(pipeline, &curPI,
				func() error { /* no modification, just r+w */ return nil })
		})
		if col.IsErrNotFound(err) {
			log.Warnf("%q polling conflicted with delete (not found)", pipeline)
		} else if err != nil {
			// No sensible error recovery here, just keep polling...
			log.Errorf("could not poll %q: %v", pipeline, err)
		}

		// move to next pipeline
		return backoff.ErrContinue
	}, backoff.NewConstantBackOff(pollBackoffTime),
		backoff.NotifyContinue("pollPipelines"),
	); err != nil {
		if ctx.Err() == nil {
			panic("pollPipelines is exiting prematurely which should not happen; restarting pod...")
		}
	}
}

func (m *ppsMaster) pollPipelinePods(pollClient *client.APIClient) {
	ctx := pollClient.Ctx()
	if err := backoff.RetryUntilCancel(ctx, func() error {
		// watchChan will be nil if the Watch call below errors, this means
		// that we won't receive events from k8s and won't be able to detect
		// errors in pods. We could just return that error and retry but that
		// prevents pachyderm from creating pipelines when there's an issue
		// talking to k8s.
		kubePipelineWatch, err := m.a.env.GetKubeClient().CoreV1().Pods(m.a.namespace).Watch(
			metav1.ListOptions{
				LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(
					map[string]string{
						"component": "worker",
					})),
				Watch: true,
			})
		if err != nil {
			return errors.Wrap(err, "failed to watch kubernetes pods")
		}
		defer kubePipelineWatch.Stop()
		for event := range kubePipelineWatch.ResultChan() {
			// if we get an error we restart the watch
			if event.Type == kube_watch.Error {
				return errors.Wrap(kube_err.FromObject(event.Object), "error while watching kubernetes pods")
			} else if event.Type == "" {
				// k8s watches seem to sometimes get stuck in a loop returning events
				// with Type = "". We treat these as errors as otherwise we get an
				// endless stream of them and can't do anything.
				return errors.New("error while watching kubernetes pods: empty event type")
			}
			pod, ok := event.Object.(*v1.Pod)
			if !ok {
				continue // irrelevant event
			}
			if pod.Status.Phase == v1.PodFailed {
				log.Errorf("pod failed because: %s", pod.Status.Message)
			}
			pipelineName := pod.ObjectMeta.Annotations["pipelineName"]
			for _, status := range pod.Status.ContainerStatuses {
				if status.State.Waiting != nil && failures[status.State.Waiting.Reason] {
					if err := m.a.setPipelineCrashing(ctx, pipelineName, status.State.Waiting.Message); err != nil {
						return errors.Wrap(err, "error moving pipeline to CRASHING")
					}
				}
			}
			for _, condition := range pod.Status.Conditions {
				if condition.Type == v1.PodScheduled &&
					condition.Status != v1.ConditionTrue && failures[condition.Reason] {
					if err := m.a.setPipelineCrashing(ctx, pipelineName, condition.Message); err != nil {
						return errors.Wrap(err, "error moving pipeline to CRASHING")
					}
				}
			}
		}
		return backoff.ErrContinue // keep polling until cancelled
	}, backoff.NewConstantBackOff(pollBackoffTime),
		backoff.NotifyContinue("pollPipelinePods")); err != nil {
		if ctx.Err() == nil {
			panic("pollPipelinePods is exiting prematurely which should not happen; restarting pod...")
		}
	}
}
