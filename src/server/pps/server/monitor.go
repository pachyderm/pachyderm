package server

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"time"

	"github.com/gogo/protobuf/types"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing/extended"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	workerserver "github.com/pachyderm/pachyderm/v2/src/server/worker/server"
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
	pipeline := pipelineInfo.Pipeline.Name
	return startMonitorThread(ctx,
		"monitorPipeline for "+pipeline, func(ctx context.Context) {
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
	pipeline := pipelineInfo.Pipeline.Name
	return startMonitorThread(ctx,
		"monitorCrashingPipeline for "+pipeline,
		func(ctx context.Context) {
			pc.monitorCrashingPipeline(ctx, pipelineInfo)
		})
}

// startMonitorThread is a helper used by startMonitor, startCrashingMonitor,
// and startPipelinePoller (in poller.go). It doesn't manipulate any of
// APIServer's fields, just wrapps the passed function in a goroutine, and
// returns a cancel() fn to cancel it and block until it returns.
func startMonitorThread(ctx context.Context, name string, f func(ctx context.Context)) func() {
	ctx, cancel := context.WithCancel(ctx)
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
			panic(name + " blocked for over a minute after cancellation; restarting container")
		}
	}
}

func (pc *pipelineController) monitorPipeline(ctx context.Context, pipelineInfo *pps.PipelineInfo) {
	pipelineName := pipelineInfo.Pipeline.Name
	log.Printf("PPS master: monitoring pipeline %q", pipelineInfo.Pipeline)
	var eg errgroup.Group
	pps.VisitInput(pipelineInfo.Details.Input, func(in *pps.Input) error { //nolint:errcheck
		if in.Cron != nil {
			eg.Go(func() error {
				return backoff.RetryNotify(func() error {
					return makeCronCommits(ctx, pc.env, pipelineInfo.Pipeline.Project, in)
				}, backoff.NewInfiniteBackOff(),
					backoff.NotifyCtx(ctx, "cron for "+in.Cron.Name))
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
				return pachClient.SubscribeCommit(client.NewProjectRepo(pipelineInfo.Pipeline.Project.GetName(), pipelineName), "", "", pfs.CommitState_READY, func(ci *pfs.CommitInfo) error {
					ciChan <- ci
					return nil
				})
			}, backoff.NewInfiniteBackOff(),
				backoff.NotifyCtx(ctx, "SubscribeCommit for "+pipelineName))
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
						if ci.Finishing != nil {
							continue
						}
						childSpan, ctx = extended.AddSpanToAnyPipelineTrace(oldCtx,
							pc.env.EtcdClient, pipelineInfo.Pipeline,
							"/pps.Master/MonitorPipeline/SpinUp",
							"commit", ci.Commit.ID)

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
							// wait for both meta and output commits
							if _, err := pachClient.PfsAPIClient.InspectCommit(ctx, &pfs.InspectCommitRequest{
								Commit: ci.Commit,
								Wait:   pfs.CommitState_FINISHED,
							}); err != nil {
								return grpcutil.ScrubGRPC(err)
							}
							if _, err := pachClient.PfsAPIClient.InspectCommit(ctx, &pfs.InspectCommitRequest{
								Commit: ppsutil.MetaCommit(ci.Commit),
								Wait:   pfs.CommitState_FINISHED,
							}); err != nil {
								return grpcutil.ScrubGRPC(err)
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
									"commit", ci.Commit.ID)
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
						return errors.EnsureStack(ctx.Err())
					}
				}
			}, backoff.NewInfiniteBackOff(),
				backoff.NotifyCtx(ctx, "monitorPipeline for "+pipelineInfo.Pipeline.String()))
		})
	}
	if err := eg.Wait(); err != nil {
		log.Printf("error in monitorPipeline: %v", err)
	}
}

func (pc *pipelineController) monitorCrashingPipeline(ctx context.Context, pipelineInfo *pps.PipelineInfo) {
	projectName := pipelineInfo.Pipeline.Project.GetName()
	pipelineName := pipelineInfo.Pipeline.Name
	ctx, cancelInner := context.WithCancel(ctx)
	if err := backoff.RetryUntilCancel(ctx, backoff.MustLoop(func() error {
		currRC, _, err := pc.getRC(ctx, pipelineInfo)
		if err != nil {
			return err
		}
		parallelism := int(*currRC.Spec.Replicas)
		workerStatus, err := workerserver.Status(ctx, projectName, pipelineName, pipelineInfo.Version,
			pc.env.EtcdClient, pc.etcdPrefix, pc.env.Config.PPSWorkerPort)
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
		log.Fatalf("monitorCrashingPipeline is exiting prematurely which should not happen (error: %v); restarting container...", err)
	}
}

func cronTick(pachClient *client.APIClient, now time.Time, project *pfs.Project, cron *pps.CronInput) error {
	return pachClient.WithModifyFileClient(
		client.NewProjectRepo(project.GetName(), cron.Repo).NewCommit("master", ""),
		func(m client.ModifyFile) error {
			if cron.Overwrite {
				if err := m.DeleteFile("/"); err != nil {
					return errors.EnsureStack(err)
				}
			}
			return errors.EnsureStack(m.PutFile(now.Format(time.RFC3339), bytes.NewReader(nil)))
		})
}

// makeCronCommits makes commits to a single cron input's repo. It's
// a helper function called by monitorPipeline.
func makeCronCommits(ctx context.Context, env Env, project *pfs.Project, in *pps.Input) error {
	schedule, err := cron.ParseStandard(in.Cron.Spec)
	if err != nil {
		return errors.EnsureStack(err) // Shouldn't happen, as the input is validated in CreatePipeline
	}
	pachClient := env.GetPachClient(ctx)
	latestTime, err := getLatestCronTime(ctx, env, project, in)
	if err != nil {
		return err
	}

	for {
		// get the time of the next time from the latest time using the cron schedule
		next := schedule.Next(latestTime)
		if next.IsZero() {
			return nil // zero time indicates there will never be another tick
		}
		// and wait until then to make the next commit
		select {
		case <-time.After(time.Until(next)):
		case <-ctx.Done():
			return errors.EnsureStack(ctx.Err())
		}
		if err := cronTick(pachClient, next, project, in.Cron); err != nil {
			return err
		}
		// set latestTime to the next time
		latestTime = next
	}
}

// getLatestCronTime is a helper used by m.makeCronCommits. It figures out what
// 'in's most recently executed cron tick was and returns it (or, if no cron
// ticks are in 'in's cron repo, it retuns the 'Start' time set in 'in.Cron'
// (typically set by 'pachctl extract')
func getLatestCronTime(ctx context.Context, env Env, project *pfs.Project, in *pps.Input) (time.Time, error) {
	var latestTime time.Time
	pachClient := env.GetPachClient(ctx)
	files, err := pachClient.ListFileAll(client.NewProjectCommit(project.GetName(), in.Cron.Repo, "master", ""), "")
	if err != nil {
		return latestTime, err
	} else if err != nil || len(files) == 0 {
		// File not found, this happens the first time the pipeline is run
		latestTime, err = types.TimestampFromProto(in.Cron.Start)
		if err != nil {
			return latestTime, errors.EnsureStack(err)
		}
	} else {
		// Take the name of the most recent file as the latest timestamp
		// ListFile returns the files in lexicographical order, and the RFC3339 format goes
		// from largest unit of time to smallest, so the most recent file will be the last one
		latestTime, err = time.Parse(time.RFC3339, path.Base(files[len(files)-1].File.Path))
		if err != nil {
			return latestTime, errors.EnsureStack(err)
		}
	}
	return latestTime, nil
}
