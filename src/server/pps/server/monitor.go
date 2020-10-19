// monitor.go contains methods of s/s/pps/server/api_server.go:APIServer that
// pertain to these fields of APIServer:
//   - monitorCancels,
//   - crashingMonitorCancels,
//   - pollCancel, and
//   - monitorCancelsMu (which protects all of the other fields).
package server

import (
	"context"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/tracing"
	"github.com/pachyderm/pachyderm/src/client/pkg/tracing/extended"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsconsts"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsutil"
	workerserver "github.com/pachyderm/pachyderm/src/server/worker/server"
)

// startMonitor starts a new goroutine running monitorPipeline for
// 'pipelineInfo.Pipeline'.
//
// Every running pipeline with standby == true or a cron input has a
// corresponding goroutine running monitorPipeline() that puts the pipeline in
// and out of standby in response to new output commits appearing in that
// pipeline's output repo.
func (a *apiServer) startMonitor(pipelineInfo *pps.PipelineInfo) {
	a.monitorCancelsMu.Lock()
	defer a.monitorCancelsMu.Unlock()
	if _, ok := a.monitorCancels[pipelineInfo.Pipeline.Name]; !ok {
		// use context.Background because we expect this goro to run for the rest of
		// pachd's lifetime
		ctx, cancel := context.WithCancel(context.Background())
		a.monitorCancels[pipelineInfo.Pipeline.Name] = cancel
		go a.sudo(a.env.GetPachClient(ctx),
			func(superUserClient *client.APIClient) error {
				a.monitorPipeline(superUserClient, pipelineInfo)
				return nil
			})
	}
}

// cancelMonitor cancels the monitorPipeline goroutine for 'pipeline'. See
// a.startMonitor().
func (a *apiServer) cancelMonitor(pipeline string) {
	a.monitorCancelsMu.Lock()
	defer a.monitorCancelsMu.Unlock()
	if cancel, ok := a.monitorCancels[pipeline]; ok {
		cancel()
		delete(a.monitorCancels, pipeline)
	}
}

// startCrashingMonitor starts a new goroutine running monitorCrashingPipeline
// for 'pipelineInfo.Pipeline'
//
// Every crashing pipeline has a corresponding goro running
// monitorCrashingPipeline that checks to see if the issues have resolved
// themselves and moves the pipeline out of crashing if they have.
func (a *apiServer) startCrashingMonitor(masterCtx context.Context, parallelism int64, pipelineInfo *pps.PipelineInfo) {
	a.monitorCancelsMu.Lock()
	defer a.monitorCancelsMu.Unlock()
	if _, ok := a.crashingMonitorCancels[pipelineInfo.Pipeline.Name]; !ok {
		ctx, cancel := context.WithCancel(context.Background())
		a.crashingMonitorCancels[pipelineInfo.Pipeline.Name] = cancel
		go a.monitorCrashingPipeline(ctx, parallelism, pipelineInfo)
	}
}

// cancelCrashingMonitor cancels the monitorCrashingPipeline goroutine for
// 'pipeline'. See a.startCrashingMonitor().
func (a *apiServer) cancelCrashingMonitor(pipeline string) {
	a.monitorCancelsMu.Lock()
	defer a.monitorCancelsMu.Unlock()
	if cancel, ok := a.crashingMonitorCancels[pipeline]; ok {
		cancel()
		delete(a.crashingMonitorCancels, pipeline)
	}
}

// cancelAllMonitorsAndCrashingMonitors overlaps with cancelMonitor and
// cancelCrashingMonitor, but also iterates over the existing members of
// a.{crashingM,m}onitorCancels in the critical section, so that all monitors
// can be cancelled without the risk that a new monitor is added between cancels
func (a *apiServer) cancelAllMonitorsAndCrashingMonitors() {
	// cancel all monitorPipeline goroutines
	a.monitorCancelsMu.Lock()
	defer a.monitorCancelsMu.Unlock()
	for _, c := range a.monitorCancels {
		c()
	}
	for _, c := range a.crashingMonitorCancels {
		c()
	}
	a.monitorCancels = make(map[string]func())
	a.crashingMonitorCancels = make(map[string]func())
}

//////////////////////////////////////////////////////////////////////////////
//                     Monitor Functions                                    //
//////////////////////////////////////////////////////////////////////////////

func (a *apiServer) monitorPipeline(pachClient *client.APIClient, pipelineInfo *pps.PipelineInfo) {
	log.Printf("PPS master: monitoring pipeline %q", pipelineInfo.Pipeline.Name)
	// If this exits (e.g. b/c Standby is false, and pipeline has no cron inputs),
	// remove this fn's cancel() call from a.monitorCancels (if it hasn't already
	// been removed, e.g. by deletePipelineResources cancelling this call), so
	// that it can be called again
	defer a.cancelMonitor(pipelineInfo.Pipeline.Name)
	var eg errgroup.Group
	pps.VisitInput(pipelineInfo.Input, func(in *pps.Input) {
		if in.Cron != nil {
			eg.Go(func() error {
				return backoff.RetryNotify(func() error {
					return a.makeCronCommits(pachClient, in)
				}, backoff.NewInfiniteBackOff(), notifyCtx(pachClient.Ctx(), "cron for "+in.Cron.Name))
			})
		}
	})
	if pipelineInfo.Standby {
		// Capacity 1 gives us a bit of buffer so we don't needlessly go into
		// standby when SubscribeCommit takes too long to return.
		ciChan := make(chan *pfs.CommitInfo, 1)
		eg.Go(func() error {
			return backoff.RetryNotify(func() error {
				return pachClient.SubscribeCommitF(pipelineInfo.Pipeline.Name, "",
					client.NewCommitProvenance(ppsconsts.SpecRepo, pipelineInfo.Pipeline.Name, pipelineInfo.SpecCommit.ID),
					"", pfs.CommitState_READY, func(ci *pfs.CommitInfo) error {
						ciChan <- ci
						return nil
					})
			}, backoff.NewInfiniteBackOff(), notifyCtx(pachClient.Ctx(), "SubscribeCommit"))
		})
		eg.Go(func() error {
			return backoff.RetryNotify(func() error {
				span, ctx := extended.AddPipelineSpanToAnyTrace(pachClient.Ctx(),
					a.env.GetEtcdClient(), pipelineInfo.Pipeline.Name, "/pps.Master/MonitorPipeline",
					"standby", pipelineInfo.Standby)
				if span != nil {
					pachClient = pachClient.WithCtx(ctx)
				}
				defer tracing.FinishAnySpan(span)

				if err := a.transitionPipelineState(pachClient.Ctx(),
					pipelineInfo.Pipeline.Name,
					pps.PipelineState_PIPELINE_RUNNING,
					pps.PipelineState_PIPELINE_STANDBY, ""); err != nil {

					pte := &ppsutil.PipelineTransitionError{}
					if errors.As(err, &pte) && pte.Current == pps.PipelineState_PIPELINE_PAUSED {
						// pipeline is stopped, exit monitorPipeline (which pausing the
						// pipeline should also do). monitorPipeline will be called when
						// it transitions back to running
						// TODO(msteffen): this should happen in the pipeline
						// controller
						return nil
					}
					return err
				}
				var (
					childSpan     opentracing.Span
					oldCtx        = ctx
					oldPachClient = pachClient
				)
				defer func() {
					tracing.FinishAnySpan(childSpan) // Finish any dangling children of 'span'
				}()
				for {
					// finish span from previous loops
					tracing.FinishAnySpan(childSpan)
					childSpan = nil

					var ci *pfs.CommitInfo
					select {
					case ci = <-ciChan:
						if ci.Finished != nil {
							continue
						}
						childSpan, ctx = tracing.AddSpanToAnyExisting(
							oldCtx, "/pps.Master/MonitorPipeline_SpinUp",
							"pipeline", pipelineInfo.Pipeline.Name, "commit", ci.Commit.ID)
						if childSpan != nil {
							pachClient = oldPachClient.WithCtx(ctx)
						}

						if err := a.transitionPipelineState(pachClient.Ctx(),
							pipelineInfo.Pipeline.Name,
							pps.PipelineState_PIPELINE_STANDBY,
							pps.PipelineState_PIPELINE_RUNNING, ""); err != nil {

							pte := &ppsutil.PipelineTransitionError{}
							if errors.As(err, &pte) && pte.Current == pps.PipelineState_PIPELINE_PAUSED {
								// pipeline is stopped, exit monitorPipeline (see above)
								return nil
							}
							return err
						}

						// Stay running while commits are available
					running:
						for {
							// Wait for the commit to be finished before blocking on the
							// job because the job may not exist yet.
							if _, err := pachClient.BlockCommit(ci.Commit.Repo.Name, ci.Commit.ID); err != nil {
								return err
							}
							if _, err := pachClient.InspectJobOutputCommit(ci.Commit.Repo.Name, ci.Commit.ID, true); err != nil {
								return err
							}

							select {
							case ci = <-ciChan:
							default:
								break running
							}
						}

						if err := a.transitionPipelineState(pachClient.Ctx(),
							pipelineInfo.Pipeline.Name,
							pps.PipelineState_PIPELINE_RUNNING,
							pps.PipelineState_PIPELINE_STANDBY, ""); err != nil {

							pte := &ppsutil.PipelineTransitionError{}
							if errors.As(err, &pte) && pte.Current == pps.PipelineState_PIPELINE_PAUSED {
								// pipeline is stopped; monitorPipeline will be called when it
								// transitions back to running
								// TODO(msteffen): this should happen in the pipeline
								// controller
								return nil
							}
							return err
						}
					case <-pachClient.Ctx().Done():
						return context.DeadlineExceeded
					}
				}
			}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
				select {
				case <-pachClient.Ctx().Done():
					return context.DeadlineExceeded
				default:
					log.Printf("error in monitorPipeline: %v: retrying in: %v", err, d)
				}
				return nil
			})
		})
	}
	if err := eg.Wait(); err != nil {
		log.Printf("error in monitorPipeline: %v", err)
	}
}

// allWorkersUp is a helper used by monitorCrashingPipelinejkjk
func (a *apiServer) allWorkersUp(ctx context.Context, parallelism64 int64, pipelineInfo *pps.PipelineInfo) (bool, error) {
	parallelism := int(parallelism64)
	if parallelism == 0 {
		parallelism = 1
	}
	workerPoolID := ppsutil.PipelineRcName(pipelineInfo.Pipeline.Name,
		pipelineInfo.Version)
	workerStatus, err := workerserver.Status(ctx, workerPoolID,
		a.env.GetEtcdClient(), a.etcdPrefix, a.workerGrpcPort)
	if err != nil {
		return false, err
	}
	return parallelism == len(workerStatus), nil
}

func (a *apiServer) monitorCrashingPipeline(ctx context.Context, parallelism int64, pipelineInfo *pps.PipelineInfo) {
For:
	for {
		select {
		case <-time.After(crashingBackoff):
		case <-ctx.Done():
			break For
		}
		time.Sleep(crashingBackoff)
		workersUp, err := a.allWorkersUp(ctx, parallelism, pipelineInfo)
		if err != nil {
			log.Printf("error in monitorCrashingPipeline: %v", err)
			continue
		}
		if workersUp {
			if err := a.transitionPipelineState(ctx, pipelineInfo.Pipeline.Name,
				pps.PipelineState_PIPELINE_CRASHING,
				pps.PipelineState_PIPELINE_RUNNING, ""); err != nil {

				pte := &ppsutil.PipelineTransitionError{}
				if errors.As(err, &pte) && pte.Current == pps.PipelineState_PIPELINE_CRASHING {
					log.Print(err) // Pipeline has moved to STOPPED or been updated--give up
					return
				}
				log.Printf("error in monitorCrashingPipeline: %v", err)
				continue
			}
			break
		}
	}
}
