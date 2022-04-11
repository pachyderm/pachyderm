package server

import (
	"context"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	kube_err "k8s.io/apimachinery/pkg/api/errors"
	kube_watch "k8s.io/apimachinery/pkg/watch"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

const pollBackoffTime = 2 * time.Second

// startPipelinePoller starts a new goroutine running pollPipelines
func (m *ppsMaster) startPipelinePoller() {
	m.pollPipelinesMu.Lock()
	defer m.pollPipelinesMu.Unlock()
	m.pollCancel = startMonitorThread(m.masterCtx, "pollPipelines", m.pollPipelines)
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
	m.pollPodsCancel = startMonitorThread(m.masterCtx, "pollPipelinePods", m.pollPipelinePods)
}

func (m *ppsMaster) cancelPipelinePodsPoller() {
	m.pollPipelinesMu.Lock()
	defer m.pollPipelinesMu.Unlock()
	if m.pollPodsCancel != nil {
		m.pollPodsCancel()
		m.pollPodsCancel = nil
	}
}

// startPipelineDBPoller starts a new goroutine running watchPipelines
func (m *ppsMaster) startPipelineWatcher() {
	m.pollPipelinesMu.Lock()
	defer m.pollPipelinesMu.Unlock()
	m.watchCancel = startMonitorThread(m.masterCtx, "watchPipelines", m.watchPipelines)
}

func (m *ppsMaster) cancelPipelineWatcher() {
	m.pollPipelinesMu.Lock()
	defer m.pollPipelinesMu.Unlock()
	if m.watchCancel != nil {
		m.watchCancel()
		m.watchCancel = nil
	}
}

//////////////////////////////////////////////////////////////////////////////
//                     PollPipelines Definition                             //
// - As in monitor.go, functions below should not call functions above, to  //
// avoid reentrancy deadlock.                                               //
//////////////////////////////////////////////////////////////////////////////

// pollPipelines generates regular updateEv and deleteEv events for each
// pipeline and sends them to ppsMaster.Run(). By scanning the database and k8s
// regularly and generating events for them, it prevents pipelines from getting
// orphaned.
func (m *ppsMaster) pollPipelines(ctx context.Context) {
	dbPipelines := map[string]bool{}
	if err := backoff.RetryUntilCancel(ctx, backoff.MustLoop(func() error {
		if len(dbPipelines) == 0 {
			// 1. Get the current set of pipeline RCs, as a base set for stale RCs.
			//
			// Pipelines are created in the database before their RC is created in
			// k8s, so to garbage-collect stale RCs, we have to go the other way and
			// query k8s first (if we were to query the database first, and
			// CreatePipeline(foo) were to run between querying the database and
			// querying k8s, then we might delete the RC for brand-new pipeline
			// 'foo'). Though, even if we do delete a live pipeline's RC, it'll be
			// fixed in the next cycle
			rcs, err := m.kd.ListReplicationControllers(ctx)
			if err != nil {
				// No sensible error recovery here (e.g .if we can't reach k8s). We'll
				// keep going, and just won't delete any RCs this round.
				log.Errorf("error polling pipeline RCs: %v", err)
			}

			// 2. Replenish 'dbPipelines' with the set of pipelines currently in the
			// database; it determines both which RCs (from above) are stale and also
			// which pipelines need to be bumped. Note that there may be zero
			// pipelines in the database, and dbPipelines may be empty.
			if err := ppsutil.ListPipelineInfo(ctx, m.pipelines, nil, 0,
				func(ptr *pps.PipelineInfo) error {
					dbPipelines[ptr.Pipeline.Name] = true
					return nil
				}); err != nil {
				// ListPipelineInfo results (dbPipelines) are used by all remaining
				// steps, so if that didn't work, start over and try again
				dbPipelines = map[string]bool{}
				return errors.Wrap(err, "error polling pipelines")
			}

			// 3. Generate a delete event for orphaned RCs
			if rcs != nil {
				for _, rc := range rcs.Items {
					pipeline, ok := rc.Labels["pipelineName"]
					if !ok {
						return errors.New("'pipelineName' label missing from rc " + rc.Name)
					}
					if !dbPipelines[pipeline] {
						m.eventCh <- &pipelineEvent{pipeline: pipeline}
					}
				}
			}

			// 4. Retry if there are no pipelines to read/write
			if len(dbPipelines) == 0 {
				return backoff.ErrContinue
			}
		}

		// Generate one event for a pipeline (to trigger the pipeline controller)
		// and remove this pipeline from dbPipelines. Always choose the
		// lexicographically smallest pipeline so that pipelines are always
		// traversed in the same order and the period between polls is stable across
		// all pipelines.
		var pipeline string
		for p := range dbPipelines {
			if pipeline == "" || p < pipeline {
				pipeline = p
			}
		}

		// always rm 'pipeline', to advance loop
		delete(dbPipelines, pipeline)

		// generate a pipeline event for 'pipeline'
		log.Debugf("PPS master: polling pipeline %q", pipeline)
		select {
		case m.eventCh <- &pipelineEvent{pipeline: pipeline}:
			break
		case <-ctx.Done():
			break
		}

		// 5. move to next pipeline (after 2s sleep)
		return nil
	}), backoff.NewConstantBackOff(pollBackoffTime),
		backoff.NotifyContinue("pollPipelines"),
	); err != nil && ctx.Err() == nil {
		log.Fatalf("pollPipelines is exiting prematurely which should not happen (error: %v); restarting container...", err)
	}
}

// pollPipelinePods creates a kubernetes watch, and for each event:
//   1) Checks if the event concerns a Pod
//   2) Checks if the Pod belongs to a pipeline (pipelineName annotation is set)
//   3) Checks if the Pod is failing
// If all three conditions are met, then the pipline (in 'pipelineName') is set
// to CRASHING
func (m *ppsMaster) pollPipelinePods(ctx context.Context) {
	if err := backoff.RetryUntilCancel(ctx, backoff.MustLoop(func() error {
		watch, cancel, err := m.kd.WatchPipelinePods(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to watch kubernetes pods")
		}
		defer cancel()
	WatchLoop:
		for {
			select {
			case <-ctx.Done():
				return nil
			case event := <-watch:
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
				crashPipeline := func(reason string) error {
					pipelineName := pod.ObjectMeta.Annotations["pipelineName"]
					pipelineVersion, versionErr := strconv.Atoi(pod.ObjectMeta.Annotations["pipelineVersion"])
					if versionErr != nil {
						return errors.Wrapf(err, "couldn't find pipeline rc version")
					}
					var pipelineInfo pps.PipelineInfo
					if err := m.pipelines.ReadOnly(ctx).GetUniqueByIndex(
						ppsdb.PipelinesVersionIndex,
						ppsdb.VersionKey(pipelineName, uint64(pipelineVersion)),
						&pipelineInfo); err != nil {
						return errors.Wrapf(err, "couldn't retrieve pipeline information")
					}
					return m.setPipelineCrashing(ctx, pipelineInfo.SpecCommit, reason)
				}
				for _, status := range pod.Status.ContainerStatuses {
					if status.State.Waiting != nil && failures[status.State.Waiting.Reason] {
						if err := crashPipeline(status.State.Waiting.Message); err != nil {
							return errors.Wrap(err, "error moving pipeline to CRASHING")
						}
						continue WatchLoop
					}
				}
				for _, condition := range pod.Status.Conditions {
					if condition.Type == v1.PodScheduled &&
						condition.Status != v1.ConditionTrue && failures[condition.Reason] {
						if err := crashPipeline(condition.Message); err != nil {
							return errors.Wrap(err, "error moving pipeline to CRASHING")
						}
						continue WatchLoop
					}
				}
			}
		}
	}), backoff.NewInfiniteBackOff(), backoff.NotifyContinue("pollPipelinePods"),
	); err != nil && ctx.Err() == nil {
		log.Fatalf("pollPipelinePods is exiting prematurely which should not happen (error: %v); restarting container...", err)
	}
}

// watchPipelines watches the 'pipelines' collection in the database and sends
// writeEv and deleteEv events to the PPS master when it sees them.
//
// watchPipelines is unlike the other poll and monitor goroutines in that it sees
// the result of other poll/monitor goroutines' writes. For example, when
// pollPipelinePods (above) observes that a pipeline is crashing and updates its
// state in the database, the flow for starting monitorPipelineCrashing is:
//
//  k8s watch ─> pollPipelinePods  ╭───> watchPipelines    ╭──> m.run()
//                      │          │            │          │      │
//                      ↓          │            ↓          │      ↓
//                   db write──────╯       m.eventCh ──────╯   m.step()
//
// Most of the other poll/monitor goroutines actually go through watchPipelines
// (by writing to the database, which is then observed by the watch below)
func (m *ppsMaster) watchPipelines(ctx context.Context) {
	if err := backoff.RetryUntilCancel(ctx, backoff.MustLoop(func() error {
		// TODO(msteffen) request only keys, since pipeline_controller.go reads
		// fresh values for each event anyway
		pipelineWatcher, err := m.pipelines.ReadOnly(ctx).Watch()
		if err != nil {
			return errors.Wrapf(err, "error creating watch")
		}
		defer pipelineWatcher.Close()

		for event := range pipelineWatcher.Watch() {
			if event.Err != nil {
				return errors.Wrapf(event.Err, "event err")
			}
			pipelineName, _, err := ppsdb.ParsePipelineKey(string(event.Key))
			if err != nil {
				return errors.Wrap(err, "bad watch event key")
			}
			switch event.Type {
			case watch.EventPut, watch.EventDelete:
				e := &pipelineEvent{
					pipeline:  pipelineName,
					timestamp: time.Unix(event.Rev, 0),
				}
				select {
				case m.eventCh <- e:
				case <-m.masterCtx.Done():
					return errors.Wrap(err, "pipeline event arrived while master is restarting")
				}
			case watch.EventError:
				log.Errorf("watchPipelines received an errored event from the pipelines watcher, %v", event.Err)
			}
		}
		return nil // reset until ctx is cancelled (RetryUntilCancel)
	}), &backoff.ZeroBackOff{}, backoff.NotifyContinue("watchPipelines"),
	); err != nil && ctx.Err() == nil {
		log.Fatalf("watchPipelines is exiting prematurely which should not happen (error: %v); restarting container...", err)
	}
}
