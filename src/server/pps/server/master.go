package server

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/types"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_watch "k8s.io/apimachinery/pkg/watch"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/tracing"
	"github.com/pachyderm/pachyderm/src/client/pkg/tracing/extended"
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
	}

	zero int32 // used to turn down RCs in scaleDownWorkersForPipeline
)

// The master process is responsible for creating/deleting workers as
// pipelines are created/removed.
func (a *apiServer) master() {
	masterLock := dlock.NewDLock(a.etcdClient, path.Join(a.etcdPrefix, masterLockPath))
	backoff.RetryNotify(func() error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		// Use the PPS token to authenticate requests. Note that all requests
		// performed in this function are performed as a cluster admin, so do not
		// pass any unvalidated user input to any requests
		pachClient := a.getPachClient().WithCtx(ctx)
		ctx, err := masterLock.Lock(ctx)
		if err != nil {
			return err
		}
		defer masterLock.Unlock(ctx)

		log.Infof("PPS master: launching master process")

		// TODO(msteffen) requestly only keys, since pipeline_controller.go reads
		// fresh values for each event anyway
		pipelineWatcher, err := a.pipelines.ReadOnly(ctx).Watch()
		if err != nil {
			return fmt.Errorf("error creating watch: %+v", err)
		}
		defer pipelineWatcher.Close()

		// watchChan will be nil if the Watch call below errors, this means
		// that we won't receive events from k8s and won't be able to detect
		// errors in pods. We could just return that error and retry but that
		// prevents pachyderm from creating pipelines when there's an issue
		// talking to k8s.
		var watchChan <-chan kube_watch.Event
		kubePipelineWatch, err := a.kubeClient.CoreV1().Pods(a.namespace).Watch(
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
					return fmt.Errorf("event err: %+v", event.Err)
				}
				switch event.Type {
				case watch.EventPut:
					// Create/Modify/Delete pipeline resources as needed per new state
					if err := a.step(pachClient, string(event.Key), event.Ver, event.Rev); err != nil {
						return err
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
					kubePipelineWatch, err = a.kubeClient.CoreV1().Pods(a.namespace).Watch(
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
				for _, status := range pod.Status.ContainerStatuses {
					if status.Name == "user" && status.State.Waiting != nil && failures[status.State.Waiting.Reason] {
						if err := a.setPipelineFailure(ctx, pod.ObjectMeta.Annotations["pipelineName"], status.State.Waiting.Message); err != nil {
							return err
						}
					}
				}
			}
		}
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		a.monitorCancelsMu.Lock()
		defer a.monitorCancelsMu.Unlock()
		for _, c := range a.monitorCancels {
			c()
		}
		a.monitorCancels = make(map[string]func())
		log.Errorf("PPS master: error running the master process: %v; retrying in %v", err, d)
		return nil
	})
	panic("internal error: PPS master has somehow exited. Restarting pod...")
}

func (a *apiServer) setPipelineFailure(ctx context.Context, pipelineName string, reason string) error {
	return ppsutil.FailPipeline(ctx, a.etcdClient, a.pipelines, pipelineName, reason)
}

func (a *apiServer) deletePipelineResources(ctx context.Context, pipelineName string) (retErr error) {
	log.Infof("PPS master: deleting resources for failed pipeline %q", pipelineName)
	span, ctx := tracing.AddSpanToAnyExisting(ctx,
		"/pps.Master/DeletePipelineResources", "pipeline", pipelineName)
	defer func(span opentracing.Span) {
		if span != nil {
			span.SetTag("err", fmt.Sprintf("%v", retErr))
		}
		tracing.FinishAnySpan(span)
	}(span)

	// Cancel any running monitorPipeline call
	a.monitorCancelsMu.Lock()
	if cancel, ok := a.monitorCancels[pipelineName]; ok {
		cancel()
		delete(a.monitorCancels, pipelineName)
	}
	a.monitorCancelsMu.Unlock()

	// Delete any services associated with op.pipeline
	selector := fmt.Sprintf("%s=%s", pipelineNameLabel, pipelineName)
	falseVal := false
	opts := &metav1.DeleteOptions{
		OrphanDependents: &falseVal,
	}
	services, err := a.kubeClient.CoreV1().Services(a.namespace).List(metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return fmt.Errorf("could not list services: %v", err)
	}
	for _, service := range services.Items {
		if err := a.kubeClient.CoreV1().Services(a.namespace).Delete(service.Name, opts); err != nil {
			if !isNotFoundErr(err) {
				return fmt.Errorf("could not delete service %q: %v", service.Name, err)
			}
		}
	}
	rcs, err := a.kubeClient.CoreV1().ReplicationControllers(a.namespace).List(metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return fmt.Errorf("could not list RCs: %v", err)
	}
	for _, rc := range rcs.Items {
		if err := a.kubeClient.CoreV1().ReplicationControllers(a.namespace).Delete(rc.Name, opts); err != nil {
			if !isNotFoundErr(err) {
				return fmt.Errorf("could not delete RC %q: %v", rc.Name, err)
			}
		}
	}
	return nil
}

func notifyCtx(ctx context.Context, name string) func(error, time.Duration) error {
	return func(err error, d time.Duration) error {
		select {
		case <-ctx.Done():
			return context.DeadlineExceeded
		default:
			log.Errorf("error in %s: %v: retrying in: %v", name, err, d)
		}
		return nil
	}
}

func (a *apiServer) setPipelineState(pachClient *client.APIClient, pipelineInfo *pps.PipelineInfo, state pps.PipelineState, reason string) (retErr error) {
	span, ctx := tracing.AddSpanToAnyExisting(pachClient.Ctx(), "/pps.Master/SetPipelineState",
		"pipeline", pipelineInfo.Pipeline.Name, "new-state", state)
	if span != nil {
		pachClient = pachClient.WithCtx(ctx)
	}
	defer func(span opentracing.Span) {
		if span != nil {
			span.SetTag("err", fmt.Sprintf("%v", retErr))
		}
		tracing.FinishAnySpan(span)
	}(span)
	if span != nil {
		pachClient = pachClient.WithCtx(ctx)
	}
	log.Infof("moving pipeline %s to %s", pipelineInfo.Pipeline.Name, state.String())
	_, err := col.NewSTM(pachClient.Ctx(), a.etcdClient, func(stm col.STM) error {
		pipelines := a.pipelines.ReadWrite(stm)
		pipelinePtr := &pps.EtcdPipelineInfo{}
		if err := pipelines.Get(pipelineInfo.Pipeline.Name, pipelinePtr); err != nil {
			return err
		}
		if span != nil {
			span.SetTag("old-state", pipelinePtr.State)
		}
		if pipelinePtr.State == pps.PipelineState_PIPELINE_FAILURE {
			return nil
		}
		pipelinePtr.State = state
		pipelinePtr.Reason = reason
		return pipelines.Put(pipelineInfo.Pipeline.Name, pipelinePtr)
	})
	return err
}

func (a *apiServer) monitorPipeline(pachClient *client.APIClient, pipelineInfo *pps.PipelineInfo) {
	log.Printf("PPS master: monitoring pipeline %q", pipelineInfo.Pipeline.Name)
	defer func(pipelineName string) {
		// If this exits (e.g. b/c Standby is false, and pipeline has no cron
		// inputs), remove this fn's cancel() call from a.monitorCancels (if it
		// hasn't already been removed, e.g. by deletePipelineResources cancelling
		// this call), so that it can be called again
		a.monitorCancelsMu.Lock()
		if cancel, ok := a.monitorCancels[pipelineName]; ok {
			cancel()
			delete(a.monitorCancels, pipelineName)
		}
		a.monitorCancelsMu.Unlock()
	}(pipelineInfo.Pipeline.Name)
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
				return pachClient.SubscribeCommitF(pipelineInfo.Pipeline.Name, pipelineInfo.OutputBranch, "", pfs.CommitState_READY, func(ci *pfs.CommitInfo) error {
					ciChan <- ci
					return nil
				})
			}, backoff.NewInfiniteBackOff(), notifyCtx(pachClient.Ctx(), "SubscribeCommit"))
		})
		eg.Go(func() error {
			return backoff.RetryNotify(func() error {
				span, ctx := extended.AddPipelineSpanToAnyTrace(pachClient.Ctx(),
					a.etcdClient, pipelineInfo.Pipeline.Name, "/pps.Master/MonitorPipeline",
					"standby", pipelineInfo.Standby)
				if span != nil {
					pachClient = pachClient.WithCtx(ctx)
				}
				defer tracing.FinishAnySpan(span) // finish 'span'

				if err := a.setPipelineState(pachClient, pipelineInfo, pps.PipelineState_PIPELINE_STANDBY, ""); err != nil {
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

						if err := a.setPipelineState(pachClient, pipelineInfo, pps.PipelineState_PIPELINE_RUNNING, ""); err != nil {
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

						if err := a.setPipelineState(pachClient, pipelineInfo, pps.PipelineState_PIPELINE_STANDBY, ""); err != nil {
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

// makeCronCommits makes commits to a single cron input's repo. It's
// a helper function called by monitorPipeline.
func (a *apiServer) makeCronCommits(pachClient *client.APIClient, in *pps.Input) error {
	schedule, err := cron.ParseStandard(in.Cron.Spec)
	if err != nil {
		return err // Shouldn't happen, as the input is validated in CreatePipeline
	}
	var tstamp *types.Timestamp
	var buffer bytes.Buffer
	if err := pachClient.GetFile(in.Cron.Repo, "master", "time", 0, 0, &buffer); err != nil && !isNilBranchErr(err) {
		return err
	} else if err != nil {
		// File not found, this happens the first time the pipeline is run
		tstamp = in.Cron.Start
	} else {
		tstamp = &types.Timestamp{}
		if err := jsonpb.UnmarshalString(buffer.String(), tstamp); err != nil {
			return err
		}
	}
	t, err := types.TimestampFromProto(tstamp)
	if err != nil {
		return err
	}
	for {
		t = schedule.Next(t)
		time.Sleep(time.Until(t))
		timestamp, err := types.TimestampProto(t)
		if err != nil {
			return err
		}
		timeString, err := (&jsonpb.Marshaler{}).MarshalToString(timestamp)
		if err != nil {
			return err
		}
		if _, err := pachClient.PutFileOverwrite(in.Cron.Repo, "master", "time", strings.NewReader(timeString), 0); err != nil {
			return err
		}
	}
}

func isNilBranchErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "has no head")
}
