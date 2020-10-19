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
)

// Every running pipeline with standby == true has a corresponding goroutine
// running monitorPipeline() that puts the pipeline in and out of standby in
// response to new output commits appearing in that pipeline's output repo
func (a *apiServer) cancelMonitor(pipeline string) {
	a.monitorCancelsMu.Lock()
	defer a.monitorCancelsMu.Unlock()
	if cancel, ok := a.monitorCancels[pipeline]; ok {
		cancel()
		delete(a.monitorCancels, pipeline)
	}
}

// Every crashing pipeline has a corresponding goro running
// monitorCrashingPipeline that checks to see if the issues have resolved
// themselves and moves the pipeline out of crashing if they have.
func (a *apiServer) cancelCrashingMonitor(pipeline string) {
	a.monitorCancelsMu.Lock()
	defer a.monitorCancelsMu.Unlock()
	if cancel, ok := a.crashingMonitorCancels[pipeline]; ok {
		cancel()
		delete(a.crashingMonitorCancels, pipeline)
	}
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

func (a *apiServer) monitorCrashingPipeline(ctx context.Context, op *pipelineOp) {
	defer a.cancelMonitor(op.name)
For:
	for {
		select {
		case <-time.After(crashingBackoff):
		case <-ctx.Done():
			break For
		}
		time.Sleep(crashingBackoff)
		workersUp, err := op.allWorkersUp()
		if err != nil {
			log.Printf("error in monitorCrashingPipeline: %v", err)
			continue
		}
		if workersUp {
			if err := a.transitionPipelineState(ctx, op.name,
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
