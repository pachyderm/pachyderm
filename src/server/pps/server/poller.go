package server

import (
	"context"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	kube_err "k8s.io/apimachinery/pkg/api/errors"
	kube_watch "k8s.io/apimachinery/pkg/watch"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

const pollBackoffTime = 2 * time.Second

// startPipelinePoller starts a new goroutine running pollPipelines
func (m *ppsMaster) startPipelinePoller() {
	m.pollPipelinesMu.Lock()
	defer m.pollPipelinesMu.Unlock()
	m.pollCancel = startMonitorThread(pctx.Child(m.masterCtx, "pollPipelines"), m.pollPipelines)
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
	m.pollPodsCancel = startMonitorThread(pctx.Child(m.masterCtx, "pollPipelinePods"), m.pollPipelinePods)
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
	m.watchCancel = startMonitorThread(pctx.Child(m.masterCtx, "watchPipelines"), m.watchPipelines)
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
	dbPipelines := map[pipelineKey]bool{}
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
				log.Info(ctx, "error polling pipeline RCs", zap.Error(err))
			}

			// 2. Replenish 'dbPipelines' with the set of pipelines currently in the
			// database; it determines both which RCs (from above) are stale and also
			// which pipelines need to be bumped. Note that there may be zero
			// pipelines in the database, and dbPipelines may be empty.
			if err := m.sd.ListPipelineInfo(ctx,
				func(ptr *pps.PipelineInfo) error {
					dbPipelines[toKey(ptr.Pipeline)] = true
					return nil
				}); err != nil {
				// ListPipelineInfo results (dbPipelines) are used by all remaining
				// steps, so if that didn't work, start over and try again
				dbPipelines = map[pipelineKey]bool{}
				return errors.Wrap(err, "error polling pipelines")
			}

			// 3. Generate a delete event for orphaned RCs
			if rcs != nil {
				for _, rc := range rcs.Items {
					projectName := rc.Labels[pipelineProjectLabel]
					pipelineName, ok := rc.Labels[pipelineNameLabel]
					if !ok {
						return errors.New("'pipelineName' label missing from rc " + rc.Name)
					}
					pipeline := newPipeline(projectName, pipelineName)
					if !dbPipelines[toKey(pipeline)] {
						log.Debug(ctx, "generating pipelineEvent for orphaned RC", zap.Stringer("pipeline", pipeline))
						m.eventCh <- &pipelineEvent{pipeline: pipeline}
					}
				}
			}

			// 4. Retry if there are no pipelines to read/write
			if len(dbPipelines) == 0 {
				return backoff.ErrContinue
			}
		}

		// Generate one event for a pKey (to trigger the pKey controller)
		// and remove this pKey from dbPipelines. Always choose the
		// lexicographically smallest pKey so that pipelines are always
		// traversed in the same order and the period between polls is stable across
		// all pipelines.
		var pKey pipelineKey
		for p := range dbPipelines {
			if pKey == "" || p < pKey {
				pKey = p
			}
		}

		// always rm 'pipeline', to advance loop
		delete(dbPipelines, pKey)

		// generate a pipeline event for 'pipeline'
		log.Debug(ctx, "polling pipeline", zap.String("pipeline", string(pKey)))
		pipeline, err := fromKey(pKey)
		if err != nil {
			return errors.Wrapf(err, "invalid pipeline key %q in dbPipelines", pKey)
		}
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
		log.Error(ctx, "pollPipelines is exiting prematurely which should not happen; restarting container...", zap.Error(err))
		os.Exit(10)
	}
}

// pollPipelinePods creates a kubernetes watch, and for each event:
//  1. Checks if the event concerns a Pod
//  2. Checks if the Pod belongs to a pipeline (pipelineName annotation is set)
//  3. Checks if the Pod is failing
//
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
			case event, ok := <-watch:
				if !ok {
					log.Debug(ctx, "kubernetes pod watch unexpectedly ended; restarting watch")
					return backoff.ErrContinue
				}
				// if we get an error we restart the watch
				if event.Type == kube_watch.Error {
					kerr := kube_err.FromObject(event.Object)
					log.Debug(ctx, "kubernetes pod watch unexpectedly ended with error; restarting watch", zap.Error(err))
					return errors.Wrap(kerr, "error while watching kubernetes pods")
				}
				pod, ok := event.Object.(*v1.Pod)
				if !ok {
					continue // irrelevant event
				}
				if pod.Status.Phase == v1.PodFailed {
					// This is one of those log messages that is going to be
					// unnecessarily alarming; there are lots of processes in
					// place to restart the pod.  But it's good to have a
					// record.
					log.Info(ctx, "worker pod failed", zap.String("podName", pod.Name), zap.Any("podStatus", pod.Status))
				}
				crashPipeline := func(reason string) error {
					projectName := pod.ObjectMeta.Annotations[pipelineProjectAnnotation]
					pipelineName := pod.ObjectMeta.Annotations[pipelineNameAnnotation]
					pipelineVersion, versionErr := strconv.Atoi(pod.ObjectMeta.Annotations["pipelineVersion"])
					if versionErr != nil {
						return errors.Wrapf(err, "couldn't find pipeline rc version")
					}
					pipeline := &pps.Pipeline{
						Project: &pfs.Project{Name: projectName},
						Name:    pipelineName,
					}
					var pipelineInfo *pps.PipelineInfo
					if pipelineInfo, err = m.sd.GetPipelineInfo(ctx, pipeline, pipelineVersion); err != nil {
						return errors.EnsureStack(err)
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
		log.Error(ctx, "pollPipelinePods is exiting prematurely which should not happen; restarting container...", zap.Error(err))
		os.Exit(11)
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
//	k8s watch ─> pollPipelinePods  ╭───> watchPipelines    ╭──> m.run()
//	                    │          │            │          │      │
//	                    ↓          │            ↓          │      ↓
//	                 db write──────╯       m.eventCh ──────╯   m.step()
//
// Most of the other poll/monitor goroutines actually go through watchPipelines
// (by writing to the database, which is then observed by the watch below)
func (m *ppsMaster) watchPipelines(ctx context.Context) {
	if err := backoff.RetryUntilCancel(ctx, backoff.MustLoop(func() error {
		// TODO(msteffen) request only keys, since pipeline_controller.go reads
		// fresh values for each event anyway
		watcher, close, err := m.sd.Watch(ctx)
		if err != nil {
			return errors.Wrapf(err, "error creating watch")
		}
		defer close()
		for event := range watcher {
			if event.Err != nil {
				return errors.Wrapf(event.Err, "event err")
			}
			projectName, pipelineName, _, err := ppsdb.ParsePipelineKey(string(event.Key))
			if err != nil {
				return errors.Wrap(err, "bad watch event key")
			}
			switch event.Type {
			case watch.EventPut, watch.EventDelete:
				e := &pipelineEvent{
					pipeline:  newPipeline(projectName, pipelineName),
					timestamp: time.Unix(event.Rev, 0),
				}
				select {
				case m.eventCh <- e:
				case <-m.masterCtx.Done():
					return errors.Wrap(err, "pipeline event arrived while master is restarting")
				}
			case watch.EventError:
				log.Error(ctx, "watchPipelines received an errored event from the pipelines watcher", zap.Error(event.Err))
			}
		}
		return nil // reset until ctx is cancelled (RetryUntilCancel)
	}), &backoff.ZeroBackOff{}, backoff.NotifyContinue("watchPipelines"),
	); err != nil && ctx.Err() == nil {
		log.Error(ctx, "watchPipelines is exiting prematurely which should not happen; restarting container...", zap.Error(err))
		os.Exit(11)
	}
}
