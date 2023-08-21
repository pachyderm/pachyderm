package server

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/cronutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing/extended"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	workerserver "github.com/pachyderm/pachyderm/v2/src/server/worker/server"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// startMonitor starts a new goroutine running monitorPipeline for
// 'pipelineInfo.Pipeline'.
//
// Every running pipeline with standby == true or a cron input has a
// corresponding goroutine running monitorPipeline() that puts the pipeline in
// and out of standby in response to new output commits appearing in that
// pipeline's output repo.
// returns a cancel()
func (pc *pipelineController) startMonitor(ctx context.Context, pipelineInfo *pps.PipelineInfo) func() {
	return startMonitorThread(
		pctx.Child(ctx, fmt.Sprintf("monitorPipeline(%s)", pipelineInfo.Pipeline),
			pctx.WithFields(zap.Stringer("pipeline", pipelineInfo.Pipeline))),
		func(ctx context.Context) {
			// monitorPipeline needs auth privileges to call subscribeCommit and
			// inspectCommit
			pachClient := pc.env.GetPachClient(ctx)
			pachClient.SetAuthToken(pipelineInfo.AuthToken)
			pc.monitorPipeline(pachClient.Ctx(), pipelineInfo)
		})
}

// startCrashingMonitor starts a new goroutine running monitorCrashingPipeline
// for 'pipelineInfo.Pipeline'
//
// Every crashing pipeline has a corresponding goro running
// monitorCrashingPipeline that checks to see if the issues have resolved
// themselves and moves the pipeline out of crashing if they have.
// returns a cancel for the crashing monitor
func (pc *pipelineController) startCrashingMonitor(ctx context.Context, pipelineInfo *pps.PipelineInfo) func() {
	return startMonitorThread(
		pctx.Child(ctx, fmt.Sprintf("monitorCrashingPipeline(%s)", pipelineInfo.Pipeline),
			pctx.WithFields(zap.Stringer("pipeline", pipelineInfo.Pipeline))),
		func(ctx context.Context) {
			pc.monitorCrashingPipeline(ctx, pipelineInfo)
		})
}

// startMonitorThread is a helper used by startMonitor, startCrashingMonitor,
// and startPipelinePoller (in poller.go). It doesn't manipulate any of
// APIServer's fields, just wrapps the passed function in a goroutine, and
// returns a cancel() fn to cancel it and block until it returns.
func startMonitorThread(ctx context.Context, f func(ctx context.Context)) func() {
	ctx, cancel := pctx.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		f(ctx)
		close(done)
	}()
	return func() {
		cancel()
		select {
		case <-done:
			return
		case <-time.After(time.Minute):
			// restart pod rather than permanently locking up the PPS master (which
			// would break the PPS API)
			log.Error(ctx, "monitorThread blocked for over a minute after cancellation; exiting")
			panic("blocked for over a minute after cancellation; restarting container")
		}
	}
}

func (pc *pipelineController) monitorPipeline(ctx context.Context, pipelineInfo *pps.PipelineInfo) {
	pipelineName := pipelineInfo.Pipeline.Name
	log.Debug(ctx, "monitoring pipeline")
	var eg errgroup.Group
	pps.VisitInput(pipelineInfo.Details.Input, func(in *pps.Input) error { //nolint:errcheck
		if in.Cron != nil {
			eg.Go(func() error {
				cctx := pctx.Child(ctx, "makeCronCommits")
				return backoff.RetryNotify(func() error {
					return makeCronCommits(cctx, pc.env, in)
				}, backoff.NewInfiniteBackOff(),
					backoff.NotifyCtx(cctx, "cron for "+in.Cron.Name))
			})
		}
		return nil
	})
	if pipelineInfo.Details.Autoscaling {
		// Capacity 1 gives us a bit of buffer so we don't needlessly go into
		// standby when SubscribeCommit takes too long to return.
		ciChan := make(chan *pfs.CommitInfo, 1)
		eg.Go(func() error {
			defer close(ciChan)
			return backoff.RetryUntilCancel(ctx, func() error {
				pachClient := pc.env.GetPachClient(ctx)
				return pachClient.SubscribeCommit(client.NewRepo(pipelineInfo.Pipeline.Project.GetName(), pipelineName), "", "", pfs.CommitState_READY, func(ci *pfs.CommitInfo) error {
					ciChan <- ci
					return nil
				})
			}, backoff.NewInfiniteBackOff(),
				backoff.NotifyCtx(ctx, "SubscribeCommit for "+pipelineInfo.Pipeline.String()))
		})
		eg.Go(func() error {
			return backoff.RetryNotify(func() error {
				var (
					oldCtx    = ctx
					childSpan opentracing.Span
					ctx       context.Context
				)
				defer func() {
					// childSpan is overwritten so wrap in a lambda for late binding
					tracing.FinishAnySpan(childSpan)
				}()
				// start span to capture & contextualize etcd state transition
				childSpan, ctx = extended.AddSpanToAnyPipelineTrace(oldCtx,
					pc.env.EtcdClient, pipelineInfo.Pipeline,
					"/pps.Master/MonitorPipeline/Begin")
				if err := pc.psDriver.TransitionState(ctx,
					pipelineInfo.SpecCommit,
					[]pps.PipelineState{
						pps.PipelineState_PIPELINE_RUNNING,
						pps.PipelineState_PIPELINE_CRASHING,
					}, pps.PipelineState_PIPELINE_STANDBY, ""); err != nil {
					pte := &ppsutil.PipelineTransitionError{}
					if errors.As(err, &pte) {
						if pte.Current == pps.PipelineState_PIPELINE_PAUSED {
							// pipeline is stopped, exit monitorPipeline (which pausing the
							// pipeline should also do). monitorPipeline will be called when
							// it transitions back to running
							// TODO(msteffen): this should happen in the pipeline
							// controller
							return nil
						} else if pte.Current != pps.PipelineState_PIPELINE_STANDBY {
							// it's fine if we were already in standby
							return errors.EnsureStack(err)
						}
					} else {
						return errors.EnsureStack(err)
					}
				}
				for {
					// finish span from previous loops
					tracing.FinishAnySpan(childSpan)
					childSpan = nil

					var ci *pfs.CommitInfo
					var ok bool
					select {
					case ci, ok = <-ciChan:
						if !ok {
							return nil // subscribeCommit exited, nothing left to do
						}
						if ci.Finished != nil {
							continue
						}
						childSpan, ctx = extended.AddSpanToAnyPipelineTrace(oldCtx,
							pc.env.EtcdClient, pipelineInfo.Pipeline,
							"/pps.Master/MonitorPipeline/SpinUp",
							"commit", ci.Commit.Id)

						if err := pc.psDriver.TransitionState(ctx,
							pipelineInfo.SpecCommit,
							[]pps.PipelineState{pps.PipelineState_PIPELINE_STANDBY},
							pps.PipelineState_PIPELINE_RUNNING, ""); err != nil {

							pte := &ppsutil.PipelineTransitionError{}
							if errors.As(err, &pte) && pte.Current == pps.PipelineState_PIPELINE_PAUSED {
								// pipeline is stopped, exit monitorPipeline (see above)
								return nil
							}
							return errors.EnsureStack(err)
						}

						// Stay running while commits are available and there's still job-related compaction to do
					running:
						for {
							pachClient := pc.env.GetPachClient(ctx)
							if err := pc.blockStandby(pachClient, ci.Commit); err != nil {
								return err
							}

							tracing.FinishAnySpan(childSpan)
							childSpan = nil
							select {
							case ci, ok = <-ciChan:
								if !ok {
									return nil // subscribeCommit exited, nothing left to do
								}
								childSpan, ctx = extended.AddSpanToAnyPipelineTrace(oldCtx,
									pc.env.EtcdClient, pipelineInfo.Pipeline,
									"/pps.Master/MonitorPipeline/WatchNext",
									"commit", ci.Commit.Id)
							default:
								break running
							}
						}

						if err := pc.psDriver.TransitionState(ctx,
							pipelineInfo.SpecCommit,
							[]pps.PipelineState{
								pps.PipelineState_PIPELINE_RUNNING,
								pps.PipelineState_PIPELINE_CRASHING,
							}, pps.PipelineState_PIPELINE_STANDBY, ""); err != nil {

							pte := &ppsutil.PipelineTransitionError{}
							if errors.As(err, &pte) && pte.Current == pps.PipelineState_PIPELINE_PAUSED {
								// pipeline is stopped; monitorPipeline will be called when it
								// transitions back to running
								// TODO(msteffen): this should happen in the pipeline
								// controller
								return nil
							}
							return errors.EnsureStack(err)
						}
					case <-ctx.Done():
						return errors.EnsureStack(context.Cause(ctx))
					}
				}
			}, backoff.NewInfiniteBackOff(),
				backoff.NotifyCtx(ctx, "monitorPipeline for "+pipelineInfo.Pipeline.String()))
		})
	}
	if err := eg.Wait(); err != nil {
		log.Info(ctx, "error in monitorPipeline", zap.Error(err))
	}
}

func (pc *pipelineController) blockStandby(pachClient *client.APIClient, commit *pfs.Commit) error {
	ctx := pachClient.Ctx()
	if pc.env.PachwInSidecar {
		if _, err := pachClient.PfsAPIClient.InspectCommit(ctx, &pfs.InspectCommitRequest{
			Commit: commit,
			Wait:   pfs.CommitState_FINISHED,
		}); err != nil {
			return err
		}
		_, err := pachClient.PfsAPIClient.InspectCommit(ctx, &pfs.InspectCommitRequest{
			Commit: ppsutil.MetaCommit(commit),
			Wait:   pfs.CommitState_FINISHED,
		})
		return err
	}
	if _, err := pachClient.PfsAPIClient.InspectCommit(ctx, &pfs.InspectCommitRequest{
		Commit: commit,
		Wait:   pfs.CommitState_FINISHING,
	}); err != nil {
		return err
	}
	_, err := pachClient.PfsAPIClient.InspectCommit(ctx, &pfs.InspectCommitRequest{
		Commit: ppsutil.MetaCommit(commit),
		Wait:   pfs.CommitState_FINISHING,
	})
	return err
}

func (pc *pipelineController) monitorCrashingPipeline(ctx context.Context, pipelineInfo *pps.PipelineInfo) {
	ctx, cancelInner := pctx.WithCancel(ctx)
	if err := backoff.RetryUntilCancel(ctx, backoff.MustLoop(func() error {
		currRC, _, err := pc.getRC(ctx, pipelineInfo)
		if err != nil {
			return err
		}
		parallelism := int(*currRC.Spec.Replicas)
		workerStatus, err := workerserver.Status(ctx, pipelineInfo, pc.env.EtcdClient, pc.etcdPrefix, pc.env.Config.PPSWorkerPort)
		if err != nil {
			return errors.Wrap(err, "could not check if all workers are up")
		}
		if len(workerStatus) >= parallelism && int(currRC.Status.ReadyReplicas) >= parallelism {
			if err := pc.psDriver.TransitionState(ctx,
				pipelineInfo.SpecCommit,
				[]pps.PipelineState{pps.PipelineState_PIPELINE_CRASHING},
				pps.PipelineState_PIPELINE_RUNNING, ""); err != nil {
				return errors.Wrap(err, "could not transition pipeline to RUNNING")
			}
			cancelInner() // done--pipeline is out of CRASHING
		}
		return nil // loop again to check for new workers
	}), backoff.NewConstantBackOff(pc.crashingBackoff),
		backoff.NotifyContinue(fmt.Sprintf("monitorCrashingPipeline for %s", pipelineInfo.Pipeline)),
	); err != nil && ctx.Err() == nil {
		// retryUntilCancel should exit iff 'ctx' is cancelled, so this should be
		// unreachable (restart master if not)
		log.Error(ctx, "monitorCrashingPipeline is exiting prematurely which should not happen; restarting container...", zap.Error(err))
		os.Exit(10)
	}
}

func cronTick(pachClient *client.APIClient, now time.Time, cron *pps.CronInput) error {
	if err := pachClient.WithModifyFileClient(
		client.NewRepo(cron.Project, cron.Repo).NewCommit("master", ""),
		func(m client.ModifyFile) error {
			if cron.Overwrite {
				if err := m.DeleteFile("/"); err != nil {
					return errors.Wrap(err, "DeleteFile(/)")
				}
			}
			file := now.Format(time.RFC3339)
			if err := m.PutFile(file, bytes.NewReader(nil)); err != nil {
				return errors.Wrapf(err, "PutFile(%s)", file)
			}
			return nil
		}); err != nil {
		return errors.Wrap(err, "WithModifyFileClient")
	}
	return nil
}

// makeCronCommits makes commits to a single cron input's repo. It's
// a helper function called by monitorPipeline.
func makeCronCommits(ctx context.Context, env Env, in *pps.Input) error {
	schedule, err := cronutil.ParseCronExpression(in.Cron.Spec)
	if err != nil {
		return errors.EnsureStack(err) // Shouldn't happen, as the input is validated in CreatePipeline
	}
	pachClient := env.GetPachClient(ctx)
	latestTime, err := getLatestCronTime(ctx, env, in)
	if err != nil {
		return errors.Wrap(err, "getLatestCronTime")
	}

	for {
		// get the time of the next time from the latest time using the cron schedule
		next := schedule.Next(latestTime)
		if next.IsZero() {
			log.Debug(ctx, "no more scheduled ticks; exiting loop")
			return nil // zero time indicates there will never be another tick
		}
		log.Info(ctx, "waiting for next cron tick", zap.Time("tick", next), zap.Time("latestTick", latestTime))
		// and wait until then to make the next commit
		select {
		case <-time.After(time.Until(next)):
		case <-ctx.Done():
			return errors.EnsureStack(context.Cause(ctx))
		}
		if err := cronTick(pachClient, next, in.Cron); err != nil {
			return errors.Wrap(err, "cronTick")
		}
		log.Info(ctx, "cron tick committed", zap.Time("tick", next))
		// set latestTime to the next time
		latestTime = next
	}
}

// getLatestCronTime is a helper used by m.makeCronCommits. It figures out what
// 'in's most recently executed cron tick was and returns it (or, if no cron
// ticks are in 'in's cron repo, it retuns the 'Start' time set in 'in.Cron'
// (typically set by 'pachctl extract')
func getLatestCronTime(ctx context.Context, env Env, in *pps.Input) (retTime time.Time, retErr error) {
	var latestTime time.Time
	pachClient := env.GetPachClient(ctx)
	defer log.Span(ctx, "getLatestCronTime")(zap.Timep("latest", &retTime), log.Errorp(&retErr))
	files, err := pachClient.ListFileAll(client.NewCommit(in.Cron.Project, in.Cron.Repo, "master", ""), "")
	// bail if cron repo is not accessible
	if err != nil {
		return latestTime, err
	}
	// otherwise get timestamp from latest filename
	if len(files) > 0 {
		// Take the name of the most recent file as the latest timestamp
		// ListFile returns the files in lexicographical order, and the RFC3339 format goes
		// from largest unit of time to smallest, so the most recent file will be the last one
		latestTime, err = time.Parse(time.RFC3339, path.Base(files[len(files)-1].File.Path))
		// bail if filename format is bad
		if err != nil {
			return latestTime, err //nolint:wrapcheck
		}
		// get cron start time to compare if previous start time was updated
		startTime := in.Cron.Start.AsTime()
		// return latest time from filename if start time cannot be determined

		if latestTime.After(startTime) {
			return latestTime, nil
		} else {
			return startTime, nil
		}
	}
	// otherwise return cron start time since there are no files in cron repo
	startTime := in.Cron.Start.AsTime()
	return startTime, nil
}
