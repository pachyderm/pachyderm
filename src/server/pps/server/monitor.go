package server

// monitor.go contains methods of s/s/pps/server/api_server.go:APIServer that
// pertain to these fields of APIServer:
//   - monitorCancels,
//   - crashingMonitorCancels,
//   - pollCancel, and
//   - monitorCancelsMu (which protects all of the other fields).
//
// Other functions of APIServer.go should not access any of
// these fields directly (particularly APIServer.monitorCancelsMu, to avoid
// deadlocks) and should instead interact with monitorPipeline and such via
// methods in this file.
//
// Likewise, to avoid reentrancy deadlocks (A -> B -> A), methods in this file
// should avoid calling other methods of APIServer defined outside this file and
// shouldn't call each other.

import (
	"context"
	"path"
	"strings"
	"time"

	"github.com/gogo/protobuf/types"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/tracing"
	"github.com/pachyderm/pachyderm/src/client/pkg/tracing/extended"
	"github.com/pachyderm/pachyderm/src/client/pps"
	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsconsts"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsutil"
	workerserver "github.com/pachyderm/pachyderm/src/server/worker/server"
)

const crashingBackoff = time.Second * 15

// startMonitorThread is a helper used by startMonitor, startCrashingMonitor,
// and startPipelinePoller. It doesn't manipulate any of APIServer's fields,
// just wrapps the passed function in a goroutine, and returns a cancel() fn to
// cancel it and block until it returns.
func startMonitorThread(parentPachClient *client.APIClient, name string, f func(pachClient *client.APIClient)) func() {
	childCtx, cancel := context.WithCancel(parentPachClient.Ctx())
	done := make(chan struct{})
	go func() {
		f(parentPachClient.WithCtx(childCtx))
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
			panic(name + " blocked for over a minute after cancellation; restarting pod")
		}
	}
}

// startMonitor starts a new goroutine running monitorPipeline for
// 'pipelineInfo.Pipeline'.
//
// Every running pipeline with standby == true or a cron input has a
// corresponding goroutine running monitorPipeline() that puts the pipeline in
// and out of standby in response to new output commits appearing in that
// pipeline's output repo.
func (a *apiServer) startMonitor(ppsMasterClient *client.APIClient, pipelineInfo *pps.PipelineInfo) {
	pipeline := pipelineInfo.Pipeline.Name
	a.monitorCancelsMu.Lock()
	defer a.monitorCancelsMu.Unlock()
	if _, ok := a.monitorCancels[pipeline]; !ok {
		a.monitorCancels[pipeline] = startMonitorThread(ppsMasterClient,
			"monitorPipeline for "+pipeline, func(pachClient *client.APIClient) {
				// monitorPipeline needs auth privileges to call subscribeCommit and
				// blockCommit
				// TODO(msteffen): run the pipeline monitor as the pipeline user, rather
				// than as the PPS superuser
				if err := a.sudo(pachClient, func(superUserClient *client.APIClient) error {
					a.monitorPipeline(superUserClient, pipelineInfo)
					return nil
				}); err != nil {
					log.Errorf("error monitoring %q: %v", pipeline, err)
				}
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
func (a *apiServer) startCrashingMonitor(ppsMasterClient *client.APIClient, parallelism uint64, pipelineInfo *pps.PipelineInfo) {
	pipeline := pipelineInfo.Pipeline.Name
	a.monitorCancelsMu.Lock()
	defer a.monitorCancelsMu.Unlock()
	if _, ok := a.crashingMonitorCancels[pipeline]; !ok {
		a.crashingMonitorCancels[pipeline] = startMonitorThread(ppsMasterClient,
			"monitorCrashingPipeline for "+pipeline,
			func(pachClient *client.APIClient) {
				a.monitorCrashingPipeline(pachClient, parallelism, pipelineInfo)
			})
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
func (a *apiServer) cancelAllMonitorsAndCrashingMonitors(leave map[string]bool) {
	a.monitorCancelsMu.Lock()
	defer a.monitorCancelsMu.Unlock()
	for _, monitorMap := range []map[string]func(){a.monitorCancels, a.crashingMonitorCancels} {
		remove := make([]string, 0, len(monitorMap))
		for p, _ := range monitorMap {
			if !leave[p] {
				remove = append(remove, p)
			}
		}
		for _, p := range remove {
			cancel := monitorMap[p]
			cancel()
			delete(monitorMap, p)
		}
	}
}

// startPipelinePoller starts a new goroutine running pollPipelines
func (a *apiServer) startPipelinePoller(ppsMasterClient *client.APIClient) {
	a.monitorCancelsMu.Lock()
	defer a.monitorCancelsMu.Unlock()
	a.pollCancel = startMonitorThread(ppsMasterClient, "pollPipelines", a.pollPipelines)
}

func (a *apiServer) cancelPipelinePoller() {
	a.monitorCancelsMu.Lock()
	defer a.monitorCancelsMu.Unlock()
	if a.pollCancel != nil {
		a.pollCancel()
		a.pollCancel = nil
	}
}

//////////////////////////////////////////////////////////////////////////////
//                     Monitor Functions                                    //
// - These do not lock monitorCancelsMu, but they are called by the         //
// functions above, which do. They should not be called by functions        //
// outside monitor.go (to avoid data races). They can in turn call each     //
// other but also should not call any of the functions above or any         //
// functions outside this file (or else they will trigger a reentrancy      //
// deadlock):                                                               //
// master.go -> monitor.go(lock succeeds) -> master -> monitor(lock fails,  //
//                                                              deadlock)   //
//////////////////////////////////////////////////////////////////////////////

func (a *apiServer) monitorPipeline(pachClient *client.APIClient, pipelineInfo *pps.PipelineInfo) {
	pipeline := pipelineInfo.Pipeline.Name
	log.Printf("PPS master: monitoring pipeline %q", pipeline)
	var eg errgroup.Group
	pps.VisitInput(pipelineInfo.Input, func(in *pps.Input) {
		if in.Cron != nil {
			eg.Go(func() error {
				return backoff.RetryNotify(func() error {
					return a.makeCronCommits(pachClient, in)
				}, backoff.NewInfiniteBackOff(),
					backoff.NotifyCtx(pachClient.Ctx(), "cron for "+in.Cron.Name))
			})
		}
	})
	if pipelineInfo.Standby {
		// Capacity 1 gives us a bit of buffer so we don't needlessly go into
		// standby when SubscribeCommit takes too long to return.
		ciChan := make(chan *pfs.CommitInfo, 1)
		eg.Go(func() error {
			defer close(ciChan)
			return backoff.RetryNotify(func() error {
				return pachClient.SubscribeCommitF(pipeline, "",
					client.NewCommitProvenance(ppsconsts.SpecRepo, pipeline, pipelineInfo.SpecCommit.ID),
					"", pfs.CommitState_READY, func(ci *pfs.CommitInfo) error {
						ciChan <- ci
						return nil
					})
			}, backoff.NewInfiniteBackOff(),
				backoff.NotifyCtx(pachClient.Ctx(), "SubscribeCommit for "+pipeline))
		})
		eg.Go(func() error {
			return backoff.RetryNotify(func() error {
				span, ctx := extended.AddPipelineSpanToAnyTrace(pachClient.Ctx(),
					a.env.GetEtcdClient(), pipeline, "/pps.Master/MonitorPipeline",
					"standby", pipelineInfo.Standby)
				if span != nil {
					pachClient = pachClient.WithCtx(ctx)
				}
				defer tracing.FinishAnySpan(span)

				if err := a.transitionPipelineState(pachClient.Ctx(),
					pipeline,
					[]pps.PipelineState{
						pps.PipelineState_PIPELINE_RUNNING,
						pps.PipelineState_PIPELINE_CRASHING,
					}, pps.PipelineState_PIPELINE_STANDBY, ""); err != nil {

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
							"pipeline", pipeline, "commit", ci.Commit.ID)
						if childSpan != nil {
							pachClient = oldPachClient.WithCtx(ctx)
						}

						if err := a.transitionPipelineState(pachClient.Ctx(),
							pipeline,
							[]pps.PipelineState{pps.PipelineState_PIPELINE_STANDBY},
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
							pipeline,
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
							return err
						}
					case <-pachClient.Ctx().Done():
						return pachClient.Ctx().Err()
					}
				}
			}, backoff.NewInfiniteBackOff(),
				backoff.NotifyCtx(pachClient.Ctx(), "monitorPipeline for "+pipeline))
		})
	}
	if err := eg.Wait(); err != nil {
		log.Printf("error in monitorPipeline: %v", err)
	}
}

// allWorkersUp is a helper used by monitorCrashingPipelinejkjk
func (a *apiServer) allWorkersUp(ctx context.Context, parallelism64 uint64, pipelineInfo *pps.PipelineInfo) (bool, error) {
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

func (a *apiServer) monitorCrashingPipeline(pachClient *client.APIClient, parallelism uint64, pipelineInfo *pps.PipelineInfo) {
	pipeline := pipelineInfo.Pipeline.Name
	ctx := pachClient.Ctx()
	if err := backoff.RetryUntilCancel(ctx, func() error {
		workersUp, err := a.allWorkersUp(ctx, parallelism, pipelineInfo)
		if err != nil {
			return errors.Wrap(err, "could not check if all workers are up")
		}
		if workersUp {
			if err := a.transitionPipelineState(ctx, pipeline,
				[]pps.PipelineState{pps.PipelineState_PIPELINE_CRASHING},
				pps.PipelineState_PIPELINE_RUNNING, ""); err != nil {
				return errors.Wrap(err, "could not transition pipeline to RUNNING")
			}
			return nil // done--pipeline is out of CRASHING
		}
		return backoff.ErrContinue // loop again to check for new workers
	}, backoff.NewConstantBackOff(crashingBackoff),
		backoff.NotifyContinue("monitorCrashingPipeline for "+pipeline),
	); err != nil && ctx.Err() == nil {
		// retryUntilCancel should exit iff 'ctx' is cancelled, so this should be
		// unreachable (restart master if not)
		panic("monitorCrashingPipeline is exiting early, this should never happen")
	}
}

func (a *apiServer) getLatestCronTime(pachClient *client.APIClient, in *pps.Input) (time.Time, error) {
	var latestTime time.Time
	files, err := pachClient.ListFile(in.Cron.Repo, "master", "")
	if err != nil && !pfsserver.IsNoHeadErr(err) {
		return latestTime, err
	} else if err != nil || len(files) == 0 {
		// File not found, this happens the first time the pipeline is run
		latestTime, err = types.TimestampFromProto(in.Cron.Start)
		if err != nil {
			return latestTime, err
		}
	} else {
		// Take the name of the most recent file as the latest timestamp
		// ListFile returns the files in lexicographical order, and the RFC3339 format goes
		// from largest unit of time to smallest, so the most recent file will be the last one
		latestTime, err = time.Parse(time.RFC3339, path.Base(files[len(files)-1].File.Path))
		if err != nil {
			return latestTime, err
		}
	}
	return latestTime, nil
}

// makeCronCommits makes commits to a single cron input's repo. It's
// a helper function called by monitorPipeline.
func (a *apiServer) makeCronCommits(pachClient *client.APIClient, in *pps.Input) error {
	schedule, err := cron.ParseStandard(in.Cron.Spec)
	if err != nil {
		return err // Shouldn't happen, as the input is validated in CreatePipeline
	}
	// make sure there isn't an unfinished commit on the branch
	commitInfo, err := pachClient.InspectCommit(in.Cron.Repo, "master")
	if err != nil && !pfsserver.IsNoHeadErr(err) {
		return err
	} else if commitInfo != nil && commitInfo.Finished == nil {
		// and if there is, delete it
		if err = pachClient.DeleteCommit(in.Cron.Repo, commitInfo.Commit.ID); err != nil {
			return err
		}
	}

	latestTime, err := a.getLatestCronTime(pachClient, in)
	if err != nil {
		return err
	}

	for {
		// get the time of the next time from the latest time using the cron schedule
		next := schedule.Next(latestTime)
		// and wait until then to make the next commit
		select {
		case <-time.After(time.Until(next)):
		case <-pachClient.Ctx().Done():
			return pachClient.Ctx().Err()
		}
		if err != nil {
			return err
		}

		// We need the DeleteFile and the PutFile to happen in the same commit
		_, err = pachClient.StartCommit(in.Cron.Repo, "master")
		if err != nil {
			return err
		}
		if in.Cron.Overwrite {
			// get rid of any files, so the new file "overwrites" previous runs
			err = pachClient.DeleteFile(in.Cron.Repo, "master", "")
			if err != nil && !isNotFoundErr(err) && !pfsserver.IsNoHeadErr(err) {
				return errors.Wrapf(err, "delete error")
			}
		}

		// Put in an empty file named by the timestamp
		_, err = pachClient.PutFile(in.Cron.Repo, "master", next.Format(time.RFC3339), strings.NewReader(""))
		if err != nil {
			return errors.Wrapf(err, "put error")
		}

		err = pachClient.FinishCommit(in.Cron.Repo, "master")
		if err != nil {
			return err
		}

		// set latestTime to the next time
		latestTime = next
	}
}

func (a *apiServer) pollPipelines(pachClient *client.APIClient) {
	ctx := pachClient.Ctx()
	if err := backoff.RetryUntilCancel(ctx, func() error {
		// 1. Get the set of pipelines currently in etcd, and generate an etcd event
		// for each one (to trigger the pipeline controller)
		etcdPipelines := map[string]bool{}
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
				} else if err != nil {
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
