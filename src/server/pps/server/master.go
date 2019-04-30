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
	kube "k8s.io/client-go/kubernetes"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/tracing"
	"github.com/pachyderm/pachyderm/src/client/pkg/tracing/extended"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
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

		pipelineWatcher, err := a.pipelines.ReadOnly(ctx).WatchWithPrev()
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

		// Watch for new pipeline creates/updates
		var (
			span          opentracing.Span
			oldCtx        = ctx
			oldPachClient = pachClient
		)
		defer func() {
			// Finish any dangling span
			// Note: must wrap 'tracing.FinishAnySpan(span)' in closure so that
			// 'span' is dereferenced after the "for" loop below runs (it's nil now)
			tracing.FinishAnySpan(span) // finish any dangling span
		}()
		for {
			// finish span from previous pipeline operation
			tracing.FinishAnySpan(span)
			span = nil

			select {
			case event := <-pipelineWatcher.Watch():
				if event.Err != nil {
					return fmt.Errorf("event err: %+v", event.Err)
				}
				switch event.Type {
				case watch.EventPut:
					var pipelineName string
					var prevPipelinePtr pps.EtcdPipelineInfo
					var pipelinePtr pps.EtcdPipelineInfo
					if err := event.Unmarshal(&pipelineName, &pipelinePtr); err != nil {
						return err
					}
					if event.PrevKey != nil {
						if err := event.UnmarshalPrev(&pipelineName, &prevPipelinePtr); err != nil {
							return err
						}
					}
					log.Infof("PPS master: pipeline %q: %s -> %s", pipelineName, prevPipelinePtr.State, pipelinePtr.State)
					var prevSpecCommit string
					if prevPipelinePtr.SpecCommit != nil {
						prevSpecCommit = prevPipelinePtr.SpecCommit.ID
					}
					var curSpecCommit = pipelinePtr.SpecCommit.ID

					// Handle tracing (pipelineRestarted is used to maybe delete trace)
					if span, ctx = extended.AddPipelineSpanToAnyTrace(oldCtx,
						a.etcdClient, pipelineName, "/pps.Master/ProcessPipelineUpdate",
						"key-version", event.Ver,
						"mod-revision", event.Rev,
						"prev-key", string(event.PrevKey),
						"old-state", prevPipelinePtr.State.String(),
						"old-spec-commit", prevSpecCommit,
						"new-state", pipelinePtr.State.String(),
						"new-spec-commit", curSpecCommit,
					); span != nil {
						pachClient = oldPachClient.WithCtx(ctx)
					} else {
						pachClient = oldPachClient
					}

					// Retrieve pipelineInfo (and prev pipeline's pipelineInfo) from the
					// spec repo
					var pipelineInfo, prevPipelineInfo *pps.PipelineInfo
					if err := a.sudo(pachClient, func(superUserClient *client.APIClient) error {
						var err error
						pipelineInfo, err = ppsutil.GetPipelineInfo(superUserClient, &pipelinePtr)
						if err != nil {
							return err
						}

						if prevPipelinePtr.SpecCommit != nil {
							prevPipelineInfo, err = ppsutil.GetPipelineInfo(superUserClient, &prevPipelinePtr)
							if err != nil {
								return err
							}
						}
						return nil
					}); err != nil {
						return fmt.Errorf("watch event had no pipelineInfo: %v", err)
					}

					// True if the pipeline has a git input
					var hasGitInput bool
					pps.VisitInput(pipelineInfo.Input, func(input *pps.Input) {
						if input.Git != nil {
							hasGitInput = true
						}
					})
					// True if the user has called StopPipeline, but the PPS master hasn't
					// processed the update yet (PPS master sets the pipeline state to
					// PAUSED)
					pipelinePartiallyPaused := pipelineInfo.Stopped &&
						pipelineInfo.State != pps.PipelineState_PIPELINE_PAUSED
					// True if the pipeline has been restarted (regardless of any change
					// to the pipeline spec)
					pipelineRestarted := !pipelineInfo.Stopped &&
						event.PrevKey != nil && prevPipelineInfo.Stopped
					// True if auth has been activated or deactivated
					authActivationChanged := (pipelinePtr.AuthToken == "") !=
						(prevPipelinePtr.AuthToken == "")
					// True if the pipeline has been created or updated
					pipelineNewSpecCommit := func() bool {
						var prevSpecCommit string
						if prevPipelinePtr.SpecCommit != nil {
							prevSpecCommit = prevPipelinePtr.SpecCommit.ID
						}
						return pipelinePtr.SpecCommit.ID != prevSpecCommit &&
							!pipelineInfo.Stopped
					}()

					// Handle cases where pipeline isn't running anymore
					if pipelineInfo.State == pps.PipelineState_PIPELINE_FAILURE {
						// pipeline fails if docker image isn't found OR if pipeline RC is
						// missing
						if err := a.finishPipelineOutputCommits(pachClient, pipelineInfo); err != nil {
							return err
						}
						if err := a.deletePipelineResources(ctx, pipelineName); err != nil {
							return err
						}
						continue
					} else if pipelineInfo.State == pps.PipelineState_PIPELINE_PAUSED {
						continue // pipeline has already been paused -- nothing to do
					} else if pipelinePartiallyPaused {
						// Clusters can get into broken state where pipeline exists but
						// RC is deleted by user--causes master() to crashloop trying to
						// scale up/down workers repeatedly. Break the cycle here
						if err := a.scaleDownWorkersForPipeline(ctx, pipelineInfo); err != nil {
							if failErr := a.setPipelineFailure(ctx, pipelineInfo.Pipeline.Name, err.Error()); failErr != nil {
								return failErr
							}
							return err
						}
						// Mark the pipeline paused, thus fully stopping it (Note: this will
						// generate another etcd event, which is ignored below)
						if err := a.setPipelineState(pachClient, pipelineInfo, pps.PipelineState_PIPELINE_PAUSED, ""); err != nil {
							return err
						}
						continue
					}

					// Handle cases where pipeline is still running
					if pipelineRestarted || authActivationChanged || pipelineNewSpecCommit {
						if (pipelineNewSpecCommit || authActivationChanged) && event.PrevKey != nil {
							if err := a.deletePipelineResources(ctx, prevPipelineInfo.Pipeline.Name); err != nil {
								return err
							}
						}
						if (pipelineNewSpecCommit || pipelineRestarted) && hasGitInput {
							if err := a.checkOrDeployGithookService(ctx); err != nil {
								return err
							}
						}
						if err := a.upsertWorkersForPipeline(ctx, pipelineInfo); err != nil {
							log.Errorf("error upserting workers for new/restarted pipeline %q: %v", pipelineName, err)
							if err := a.setPipelineState(pachClient, pipelineInfo, pps.PipelineState_PIPELINE_STARTING, fmt.Sprintf("failed to create workers: %s", err.Error())); err != nil {
								return err
							}
							// We return the error here, this causes us to go
							// into backoff and try again from scratch. This
							// means that we'll try creating this pipeline
							// again and also gives a chance for another node,
							// which might actually be able to talk to k8s, to
							// get a chance at creating the workers.
							return err
						}
					}
					if pipelineInfo.State == pps.PipelineState_PIPELINE_RUNNING {
						if err := a.scaleUpWorkersForPipeline(ctx, pipelineInfo); err != nil {
							if isNotFoundErr(err) {
								// See note above (under "if pipelinePartiallyStopped")
								if failErr := a.setPipelineFailure(ctx, pipelineInfo.Pipeline.Name, err.Error()); failErr != nil {
									return failErr
								}
							}
							return err
						}
					}
					if pipelineInfo.State == pps.PipelineState_PIPELINE_STANDBY {
						if err := a.scaleDownWorkersForPipeline(ctx, pipelineInfo); err != nil {
							// See note above (under "if pipelinePartiallyStopped")
							if failErr := a.setPipelineFailure(ctx, pipelineInfo.Pipeline.Name, err.Error()); failErr != nil {
								return failErr
							}
							return err
						}
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
		for _, c := range a.monitorCancels {
			c()
		}
		a.monitorCancels = make(map[string]func())
		log.Errorf("master: error running the master process: %v; retrying in %v", err, d)
		return nil
	})
	panic("internal error: PPS master has somehow exited. Restarting pod...")
}

func (a *apiServer) setPipelineFailure(ctx context.Context, pipelineName string, reason string) error {
	return ppsutil.FailPipeline(ctx, a.etcdClient, a.pipelines, pipelineName, reason)
}

func (a *apiServer) checkOrDeployGithookService(ctx context.Context) error {
	_, err := getGithookService(a.kubeClient, a.namespace)
	if err != nil {
		if _, ok := err.(*errGithookServiceNotFound); ok {
			svc := assets.GithookService(a.namespace)
			_, err = a.kubeClient.CoreV1().Services(a.namespace).Create(svc)
			return err
		}
		return err
	}
	// service already exists
	return nil
}

func getGithookService(kubeClient *kube.Clientset, namespace string) (*v1.Service, error) {
	labels := map[string]string{
		"app":   "githook",
		"suite": suite,
	}
	serviceList, err := kubeClient.CoreV1().Services(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		},
		LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(labels)),
	})
	if err != nil {
		return nil, err
	}
	if len(serviceList.Items) != 1 {
		return nil, &errGithookServiceNotFound{
			fmt.Errorf("expected 1 githook service but found %v", len(serviceList.Items)),
		}
	}
	return &serviceList.Items[0], nil
}

func (a *apiServer) upsertWorkersForPipeline(ctx context.Context, pipelineInfo *pps.PipelineInfo) (retErr error) {
	log.Infof("PPS master: upserting workers for %q", pipelineInfo.Pipeline.Name)
	span, ctx := tracing.AddSpanToAnyExisting(ctx, "/pps.Master/UpsertWorkersForPipeline", "pipeline", pipelineInfo.Pipeline.Name)
	defer func(span opentracing.Span) {
		if span != nil {
			span.SetTag("err", fmt.Sprintf("%v", retErr))
		}
		tracing.FinishAnySpan(span)
	}(span)
	var errCount int
	if err := backoff.RetryNotify(func() error {
		var resourceRequests *v1.ResourceList
		var resourceLimits *v1.ResourceList
		if pipelineInfo.ResourceRequests != nil {
			var err error
			resourceRequests, err = ppsutil.GetRequestsResourceListFromPipeline(pipelineInfo)
			if err != nil {
				return err
			}
		}
		if pipelineInfo.ResourceLimits != nil {
			var err error
			resourceLimits, err = ppsutil.GetLimitsResourceListFromPipeline(pipelineInfo)
			if err != nil {
				return err
			}
		}

		// Retrieve the current state of the RC.  If the RC is scaled down,
		// we want to ensure that it remains scaled down.
		span, _ = tracing.AddSpanToAnyExisting(ctx, "/kube.RC/Get", "pipeline", pipelineInfo.Pipeline.Name)
		rc := a.kubeClient.CoreV1().ReplicationControllers(a.namespace)
		workerRc, err := rc.Get(
			ppsutil.PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version),
			metav1.GetOptions{})
		tracing.FinishAnySpan(span)
		var rcFound bool
		if err != nil && !isNotFoundErr(err) {
			return err
		} else if err == nil {
			rcFound = true
		}
		newPachVersion := rcFound && workerRc.ObjectMeta.Labels["version"] != version.PrettyVersion()

		// Generate options for new RC
		options := a.getWorkerOptions(
			pipelineInfo.Pipeline.Name,
			pipelineInfo.Version,
			0,
			resourceRequests,
			resourceLimits,
			pipelineInfo.Transform,
			pipelineInfo.CacheSize,
			pipelineInfo.Service,
			pipelineInfo.SpecCommit.ID,
			pipelineInfo.SchedulingSpec,
			pipelineInfo.PodSpec)
		// Set the pipeline name env
		options.workerEnv = append(options.workerEnv, v1.EnvVar{
			Name:  client.PPSPipelineNameEnv,
			Value: pipelineInfo.Pipeline.Name,
		})

		if newPachVersion {
			if err := a.deletePipelineResources(ctx, pipelineInfo.Pipeline.Name); err != nil {
				return err
			}
		}
		span, _ = tracing.AddSpanToAnyExisting(ctx, "/kube.RC/Create", "pipeline", pipelineInfo.Pipeline.Name)
		defer tracing.FinishAnySpan(span)
		return a.createWorkerRc(options)
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		errCount++
		if errCount >= 3 {
			return err
		}
		log.Errorf("error creating workers for pipeline %v: %v; retrying in %v", pipelineInfo.Pipeline.Name, err, d)
		return nil
	}); err != nil {
		return err
	}
	if _, ok := a.monitorCancels[pipelineInfo.Pipeline.Name]; !ok {
		ctx, cancel := context.WithCancel(a.pachClient.Ctx())
		a.monitorCancels[pipelineInfo.Pipeline.Name] = cancel
		pachClient := a.pachClient.WithCtx(ctx)

		go a.sudo(pachClient, func(superUserClient *client.APIClient) error {
			a.monitorPipeline(superUserClient, pipelineInfo)
			return nil
		})
	}
	return nil
}

func (a *apiServer) finishPipelineOutputCommits(pachClient *client.APIClient, pipelineInfo *pps.PipelineInfo) (retErr error) {
	pipelineName := pipelineInfo.Pipeline.Name
	log.Infof("PPS master: finishing output commits for pipeline %q", pipelineName)
	span, ctx := tracing.AddSpanToAnyExisting(pachClient.Ctx(), "/pps.Master/FinishPipelineOutputCommits", "pipeline", pipelineName)
	if span != nil {
		pachClient = pachClient.WithCtx(ctx)
	}
	defer func(span opentracing.Span) {
		if span != nil {
			span.SetTag("err", fmt.Sprintf("%v", retErr))
		}
		tracing.FinishAnySpan(span)
	}(span)
	cancel, ok := a.monitorCancels[pipelineName]
	if ok {
		cancel()
		delete(a.monitorCancels, pipelineName)
	}

	commitInfos, err := pachClient.ListCommit(pipelineName, pipelineInfo.OutputBranch, "", 0)
	if err != nil {
		return fmt.Errorf("could not list output commits of %q to finish them: %v", pipelineName, err)
	}

	var finishCommitErr error
	for _, ci := range commitInfos {
		if ci.Finished != nil {
			continue // nothing needs to be done
		}
		if err := pachClient.FinishCommit(pipelineName, ci.Commit.ID); err != nil && finishCommitErr == nil {
			finishCommitErr = err
		}
	}
	return finishCommitErr
}

func (a *apiServer) deletePipelineResources(ctx context.Context, pipelineName string) (retErr error) {
	log.Infof("PPS master: deleting resources for failed pipeline %q", pipelineName)
	span, ctx := tracing.AddSpanToAnyExisting(ctx, "/pps.Master/DeletePipelineResources", "pipeline", pipelineName)
	defer func(span opentracing.Span) {
		if span != nil {
			span.SetTag("err", fmt.Sprintf("%v", retErr))
		}
		tracing.FinishAnySpan(span)
	}(span)
	cancel, ok := a.monitorCancels[pipelineName]
	if ok {
		cancel()
		delete(a.monitorCancels, pipelineName)
	}
	selector := fmt.Sprintf("pipelineName=%s", pipelineName)
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

func (a *apiServer) scaleDownWorkersForPipeline(ctx context.Context, pipelineInfo *pps.PipelineInfo) (retErr error) {
	log.Infof("scaling down workers for %q", pipelineInfo.Pipeline.Name)
	span, ctx := tracing.AddSpanToAnyExisting(ctx, "/pps.Master/ScaleDownWorkersForPipeline", "pipeline", pipelineInfo.Pipeline.Name)
	defer func(span opentracing.Span) {
		if span != nil {
			span.SetTag("err", fmt.Sprintf("%v", retErr))
		}
		tracing.FinishAnySpan(span)
	}(span)

	rc := a.kubeClient.CoreV1().ReplicationControllers(a.namespace)

	workerRc, err := rc.Get(
		ppsutil.PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version),
		metav1.GetOptions{})
	if err != nil {
		return err
	}
	workerRc.Spec.Replicas = &zero
	_, err = rc.Update(workerRc)
	if err != nil {
		return fmt.Errorf("could not update scaled-down pipeline (%q) RC: %v",
			pipelineInfo.Pipeline.Name, err)
	}
	return nil
}

func (a *apiServer) scaleUpWorkersForPipeline(ctx context.Context, pipelineInfo *pps.PipelineInfo) (retErr error) {
	span, ctx := tracing.AddSpanToAnyExisting(ctx, "/pps.Master/ScaleUpWorkersForPipeline", "pipeline", pipelineInfo.Pipeline.Name)
	defer func(span opentracing.Span) {
		if span != nil {
			span.SetTag("err", fmt.Sprintf("%v", retErr))
		}
		tracing.FinishAnySpan(span)
	}(span)

	rc := a.kubeClient.CoreV1().ReplicationControllers(a.namespace)
	workerRc, err := rc.Get(
		ppsutil.PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version),
		metav1.GetOptions{})
	if err != nil {
		return err
	}
	parallelism, err := ppsutil.GetExpectedNumWorkers(a.kubeClient, pipelineInfo.ParallelismSpec)
	if err != nil {
		log.Errorf("error getting number of workers, default to 1 worker: %v", err)
		parallelism = 1
	}
	*workerRc.Spec.Replicas = int32(parallelism)
	_, err = rc.Update(workerRc)
	return err
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
	if !pipelineInfo.Standby {
		// Standby is false so simply put it in RUNNING and leave it there.  This is
		// only done with eg.Go so that we can handle all the errors in the
		// same way below, it should be a very quick operation so there's no
		// good reason to do it concurrently.
		eg.Go(func() error {
			return backoff.RetryNotify(func() error {
				span, ctx := extended.AddPipelineSpanToAnyTrace(pachClient.Ctx(),
					a.etcdClient, pipelineInfo.Pipeline.Name, "/pps.Master/MonitorPipeline",
					"standby", "false")
				if span != nil {
					pachClient = pachClient.WithCtx(ctx)
				}
				defer tracing.FinishAnySpan(span)
				return a.setPipelineState(pachClient, pipelineInfo, pps.PipelineState_PIPELINE_RUNNING, "")
			}, backoff.NewInfiniteBackOff(), notifyCtx(pachClient.Ctx(), "set running (Standby = false)"))
		})
	} else {
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
					"standby", "false")
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
