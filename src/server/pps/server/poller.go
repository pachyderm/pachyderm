package server

import (
	"time"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
)

// startPipelinePoller starts a new goroutine running pollPipelines
func (a *apiServer) startPipelinePoller(ppsMasterClient *client.APIClient) {
	a.pollPipelinesMu.Lock()
	defer a.pollPipelinesMu.Unlock()
	a.pollCancel = startMonitorThread(ppsMasterClient, "pollPipelines", a.pollPipelines)
}

func (a *apiServer) cancelPipelinePoller() {
	a.pollPipelinesMu.Lock()
	defer a.pollPipelinesMu.Unlock()
	if a.pollCancel != nil {
		a.pollCancel()
		a.pollCancel = nil
	}
}

//////////////////////////////////////////////////////////////////////////////
//                     PollPipelines Definition                             //
// - As in monitor.go, functions below should not call functions above, to  //
// avoid reentrancy deadlock.                                               //
//////////////////////////////////////////////////////////////////////////////

func (a *apiServer) pollPipelines(pachClient *client.APIClient) {
	ctx := pachClient.Ctx()
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
			kc := a.env.GetKubeClient().CoreV1().ReplicationControllers(a.env.Namespace)
			rcs, err := kc.List(metav1.ListOptions{
				LabelSelector: "suite=pachyderm,pipelineName",
			})
			if err != nil {
				// No sensible error recovery here (e.g .if we can't reach k8s). We'll
				// keep going, and just won't delete any RCs this round.
				log.Errorf("error polling pipeline RCs: %v", err)
			}

			// 2. Replenish 'etcdPipelines' with the set of pipelines currently in etcd
			if err := a.listPipelinePtr(pachClient, nil, 0,
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
						if err := a.deletePipelineResources(ctx, pipeline); err != nil {
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
			a.cancelAllMonitorsAndCrashingMonitors(etcdPipelines)
		}

		// Generate one etcd event for a pipeline (to trigger the pipeline
		// controller) and remove this pipeline from etcdPipelines
		var pipeline string
		for p := range etcdPipelines {
			pipeline = p
			break // we only want to generate one event
		}

		// always rm 'pipeline', to advance loop
		delete(etcdPipelines, pipeline)

		// generate an etcd event for 'pipeline' by reading it & writing it back
		log.Debugf("PPS master: polling pipeline %q", pipeline)
		var curPI pps.EtcdPipelineInfo
		_, err := col.NewSTM(ctx, a.env.GetEtcdClient(), func(stm col.STM) error {
			return a.pipelines.ReadWrite(stm).Update(pipeline, &curPI,
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
	}, backoff.NewConstantBackOff(2*time.Second),
		backoff.NotifyContinue("pollPipelines"),
	); err != nil {
		if ctx.Err() == nil {
			panic("pollPipelines is exiting prematurely which should not happen; restarting pod...")
		}
	}
}
