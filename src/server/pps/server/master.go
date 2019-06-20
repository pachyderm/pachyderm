package server

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

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
	masterLock := dlock.NewDLock(a.env.GetEtcdClient(), path.Join(a.etcdPrefix, masterLockPath))
	backoff.RetryNotify(func() error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		// Use the PPS token to authenticate requests. Note that all requests
		// performed in this function are performed as a cluster admin, so do not
		// pass any unvalidated user input to any requests
		pachClient := a.env.GetPachClient(ctx)
		ctx, err := masterLock.Lock(ctx)
		if err != nil {
			return err
		}
		defer masterLock.Unlock(ctx)
		kubeClient := a.env.GetKubeClient()

		log.Infof("PPS master: launching master process")

		pipelineWatcher, err := a.pipelines.ReadOnly(ctx).Watch(watch.WithPrevKV())
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
		kubePipelineWatch, err := kubeClient.CoreV1().Pods(a.namespace).Watch(
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
			// Note: cannot do 'defer tracing.FinishAnySpan(span)' b/c that would
			// evaluate 'span' before the "for" loop below runs
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
						a.env.GetEtcdClient(), pipelineName, "/pps.Master/ProcessPipelineUpdate",
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
					// True if the user has called StartPipeline, but the PPS master
					// hasn't processed the update yet
					pipelinePartiallyUnpaused := !pipelineInfo.Stopped &&
						pipelineInfo.State == pps.PipelineState_PIPELINE_PAUSED
					// True if the user has called StopPipeline, and the PPS master
					// has processed the update yet
					pipelineFullyPaused := pipelineInfo.Stopped &&
						pipelineInfo.State == pps.PipelineState_PIPELINE_PAUSED
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
						// generate another etcd event, which is ignored below, under
						// pipelineFullyPaused)
						if err := a.setPipelineState(pachClient, pipelineInfo, pps.PipelineState_PIPELINE_PAUSED, ""); err != nil {
							return err
						}
						continue
					} else if pipelinePartiallyUnpaused {
						// Mark the pipeline RUNNING, which will generate another etcd
						// event, which triggers a.scaleUpWorkersForPipeline, below
						if err := a.setPipelineState(pachClient, pipelineInfo, pps.PipelineState_PIPELINE_RUNNING, ""); err != nil {
							return err
						}
						continue
					} else if pipelineFullyPaused {
						continue // pipeline has already been paused -- nothing to do
					}

					// Handle cases where pipeline is still running
					if pipelineRestarted || authActivationChanged || pipelineNewSpecCommit {
						if (pipelineNewSpecCommit || authActivationChanged) && event.PrevKey != nil {
							if err := a.deletePipelineResources(ctx, prevPipelineInfo.Pipeline.Name); err != nil {
								return err
							}
						}
						if (pipelineNewSpecCommit || pipelineRestarted) && hasGitInput {
							if err := a.checkOrDeployGithookService(); err != nil {
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
							// See note above (under "if pipelinePartiallyStopped")
							if failErr := a.setPipelineFailure(ctx, pipelineInfo.Pipeline.Name, err.Error()); failErr != nil {
								return failErr
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
					kubePipelineWatch, err = kubeClient.CoreV1().Pods(a.namespace).Watch(
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
	return ppsutil.FailPipeline(ctx, a.env.GetEtcdClient(), a.pipelines, pipelineName, reason)
}

func (a *apiServer) checkOrDeployGithookService() error {
	kubeClient := a.env.GetKubeClient()
	_, err := getGithookService(kubeClient, a.namespace)
	if err != nil {
		if _, ok := err.(*errGithookServiceNotFound); ok {
			svc := assets.GithookService(a.namespace)
			_, err = kubeClient.CoreV1().Services(a.namespace).Create(svc)
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
		tracing.TagAnySpan(span, "err", retErr)
		tracing.FinishAnySpan(span)
	}(span) // bind span eagerly, as it's overwritten below
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
		rc := a.env.GetKubeClient().CoreV1().ReplicationControllers(a.namespace)
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
		if rcFound && workerRc.ObjectMeta.Labels["version"] != version.PrettyVersion() {
			if err := a.deletePipelineResources(ctx, pipelineInfo.Pipeline.Name); err != nil {
				return err
			}
		}

		// A service can be present either directly on the pipeline spec
		// or on the spout field of the spec.
		var service *pps.Service
		if pipelineInfo.Spout != nil && pipelineInfo.Service != nil {
			return errors.New("only one of pipeline.service or pipeline.spout can be set")
		} else if pipelineInfo.Spout != nil && pipelineInfo.Spout.Service != nil {
			service = pipelineInfo.Spout.Service
		} else {
			service = pipelineInfo.Service
		}

		// Generate options for new RC
		options := a.getWorkerOptions(
			pipelineInfo.Pipeline.Name,
			pipelineInfo.Version,
			0,
			resourceRequests,
			resourceLimits,
			pipelineInfo.Transform,
			pipelineInfo.CacheSize,
			service,
			pipelineInfo.SpecCommit.ID,
			pipelineInfo.SchedulingSpec,
			pipelineInfo.PodSpec,
			pipelineInfo.PodPatch)
		// Set the pipeline name env
		options.workerEnv = append(options.workerEnv, v1.EnvVar{
			Name:  client.PPSPipelineNameEnv,
			Value: pipelineInfo.Pipeline.Name,
		})
		return a.createWorkerRc(ctx, options)
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

	// Start monitorPipeline() for this pipeline, which controls the new workers
	a.monitorCancelsMu.Lock()
	defer a.monitorCancelsMu.Unlock()
	if _, ok := a.monitorCancels[pipelineInfo.Pipeline.Name]; !ok {
		ctx, cancel := context.WithCancel(context.Background())
		a.monitorCancels[pipelineInfo.Pipeline.Name] = cancel
		pachClient := a.env.GetPachClient(ctx)

		go a.sudo(pachClient, func(superUserClient *client.APIClient) error {
			a.monitorPipeline(superUserClient, pipelineInfo)
			return nil
		})
	}
	return nil
}

// finishPipelineOutputCommits finishes any output commits of
// 'pipelineInfo.Pipeline' with an empty tree.
// TODO(msteffen) Note that if the pipeline has any jobs (which can happen if
// the user manually deletes the pipeline's RC, failing the pipeline, after it
// has created jobs) those will not be updated. TODO is to mark jobs FAILED
func (a *apiServer) finishPipelineOutputCommits(pachClient *client.APIClient, pipelineInfo *pps.PipelineInfo) (retErr error) {
	pipelineName := pipelineInfo.Pipeline.Name
	log.Infof("PPS master: finishing output commits for pipeline %q", pipelineName)
	span, _ctx := tracing.AddSpanToAnyExisting(pachClient.Ctx(), "/pps.Master/FinishPipelineOutputCommits", "pipeline", pipelineName)
	if span != nil {
		pachClient = pachClient.WithCtx(_ctx) // copy auth info from input to output
	}
	defer func() {
		tracing.TagAnySpan(span, "err", retErr)
		tracing.FinishAnySpan(span)
	}()

	return a.sudo(pachClient, func(superUserClient *client.APIClient) error {
		commitInfos, err := superUserClient.ListCommit(pipelineName, pipelineInfo.OutputBranch, "", 0)
		if err != nil {
			return fmt.Errorf("could not list output commits of %q to finish them: %v", pipelineName, err)
		}

		var finishCommitErr error
		for _, ci := range commitInfos {
			if ci.Finished != nil {
				continue // nothing needs to be done
			}
			if _, err := superUserClient.PfsAPIClient.FinishCommit(superUserClient.Ctx(),
				&pfs.FinishCommitRequest{
					Commit: client.NewCommit(pipelineName, ci.Commit.ID),
					Empty:  true,
				}); err != nil && finishCommitErr == nil {
				finishCommitErr = err
			}
		}
		return finishCommitErr
	})
}

func (a *apiServer) deletePipelineResources(ctx context.Context, pipelineName string) (retErr error) {
	log.Infof("PPS master: deleting resources for pipeline %q", pipelineName)
	span, ctx := tracing.AddSpanToAnyExisting(ctx,
		"/pps.Master/DeletePipelineResources", "pipeline", pipelineName)
	defer func() {
		tracing.TagAnySpan(span, "err", retErr)
		tracing.FinishAnySpan(span)
	}()

	// Cancel any running monitorPipeline call
	a.monitorCancelsMu.Lock()
	if cancel, ok := a.monitorCancels[pipelineName]; ok {
		cancel()
		delete(a.monitorCancels, pipelineName)
	}
	a.monitorCancelsMu.Unlock()

	kubeClient := a.env.GetKubeClient()
	// Delete any services associated with op.pipeline
	selector := fmt.Sprintf("%s=%s", pipelineNameLabel, pipelineName)
	falseVal := false
	opts := &metav1.DeleteOptions{
		OrphanDependents: &falseVal,
	}
	services, err := kubeClient.CoreV1().Services(a.namespace).List(metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return fmt.Errorf("could not list services: %v", err)
	}
	for _, service := range services.Items {
		if err := kubeClient.CoreV1().Services(a.namespace).Delete(service.Name, opts); err != nil {
			if !isNotFoundErr(err) {
				return fmt.Errorf("could not delete service %q: %v", service.Name, err)
			}
		}
	}
	rcs, err := kubeClient.CoreV1().ReplicationControllers(a.namespace).List(metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return fmt.Errorf("could not list RCs: %v", err)
	}
	for _, rc := range rcs.Items {
		if err := kubeClient.CoreV1().ReplicationControllers(a.namespace).Delete(rc.Name, opts); err != nil {
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
	defer func() {
		tracing.TagAnySpan(span, "err", retErr)
		tracing.FinishAnySpan(span)
	}()

	return backoff.RetryNotify(func() error {
		rc := a.env.GetKubeClient().CoreV1().ReplicationControllers(a.namespace)
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
	}, backoff.New10sBackOff(), func(err error, _ time.Duration) error {
		if err != nil && strings.Contains(err.Error(), "try again") {
			return nil
		}
		return err
	})
}

func (a *apiServer) scaleUpWorkersForPipeline(ctx context.Context, pipelineInfo *pps.PipelineInfo) (retErr error) {
	log.Infof("scaling up workers for %q", pipelineInfo.Pipeline.Name)
	span, ctx := tracing.AddSpanToAnyExisting(ctx, "/pps.Master/ScaleUpWorkersForPipeline", "pipeline", pipelineInfo.Pipeline.Name)
	defer func() {
		tracing.TagAnySpan(span, "err", retErr)
		tracing.FinishAnySpan(span)
	}()

	return backoff.RetryNotify(func() error {
		rc := a.env.GetKubeClient().CoreV1().ReplicationControllers(a.namespace)
		workerRc, err := rc.Get(
			ppsutil.PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version),
			metav1.GetOptions{})
		if err != nil {
			return err
		}
		parallelism, err := ppsutil.GetExpectedNumWorkers(a.env.GetKubeClient(), pipelineInfo.ParallelismSpec)
		if err != nil {
			log.Errorf("error getting number of workers, default to 1 worker: %v", err)
			parallelism = 1
		}
		*workerRc.Spec.Replicas = int32(parallelism)
		_, err = rc.Update(workerRc)
		return err
	}, backoff.New10sBackOff(), func(err error, _ time.Duration) error {
		if err != nil && strings.Contains(err.Error(), "try again") {
			return nil
		}
		return err
	})
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
	defer func() {
		tracing.TagAnySpan(span, "err", retErr)
		tracing.FinishAnySpan(span)
	}()
	log.Infof("moving pipeline %s to %s", pipelineInfo.Pipeline.Name, state.String())
	_, err := col.NewSTM(pachClient.Ctx(), a.env.GetEtcdClient(), func(stm col.STM) error {
		pipelines := a.pipelines.ReadWrite(stm)
		pipelinePtr := &pps.EtcdPipelineInfo{}
		if err := pipelines.Get(pipelineInfo.Pipeline.Name, pipelinePtr); err != nil {
			return err
		}
		tracing.TagAnySpan(span, "old-state", pipelinePtr.State)
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
	if !pipelineInfo.Standby {
		// Standby is false so simply put it in RUNNING and leave it there.  This is
		// only done with eg.Go so that we can handle all the errors in the
		// same way below, it should be a very quick operation so there's no
		// good reason to do it concurrently.
		eg.Go(func() error {
			return backoff.RetryNotify(func() error {
				span, ctx := extended.AddPipelineSpanToAnyTrace(pachClient.Ctx(),
					a.env.GetEtcdClient(), pipelineInfo.Pipeline.Name, "/pps.Master/MonitorPipeline",
					"standby", pipelineInfo.Standby)
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
					a.env.GetEtcdClient(), pipelineInfo.Pipeline.Name, "/pps.Master/MonitorPipeline",
					"standby", pipelineInfo.Standby)
				if span != nil {
					pachClient = pachClient.WithCtx(ctx)
				}
				defer tracing.FinishAnySpan(span)

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
	// make sure there isn't an unfinished commit on the branch
	commitInfo, err := pachClient.InspectCommit(in.Cron.Repo, "master")
	if err != nil && !isNilBranchErr(err) {
		return err
	} else if commitInfo != nil && commitInfo.Finished == nil {
		// and if there is, delete it
		if err = pachClient.DeleteCommit(in.Cron.Repo, "master"); err != nil {
			return err
		}
	}

	var latestTime time.Time
	files, err := pachClient.ListFile(in.Cron.Repo, "master", "")
	if err != nil && !isNilBranchErr(err) {
		return err
	} else if err != nil || len(files) == 0 {
		// File not found, this happens the first time the pipeline is run
		latestTime, err = types.TimestampFromProto(in.Cron.Start)
		if err != nil {
			return err
		}
	} else {
		// Take the name of the most recent file as the latest timestamp
		// ListFile returns the files in lexicographical order, and the RFC3339 format goes
		// from largest unit of time to smallest, so the most recent file will be the last one
		latestTime, err = time.Parse(time.RFC3339, path.Base(files[len(files)-1].File.Path))
		if err != nil {
			return err
		}
	}

	for {
		// get the time of the next time from the latest time using the cron schedule
		next := schedule.Next(latestTime)
		// and wait until then to make the next commit
		select {
		case <-time.After(time.Until(next)):
			break
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
			// If we want to "overwrite" the file, we need to delete the file with the previous time
			err := pachClient.DeleteFile(in.Cron.Repo, "master", latestTime.Format(time.RFC3339))
			if err != nil && !isNotFoundErr(err) && !isNilBranchErr(err) {
				return fmt.Errorf("delete error %v", err)
			}
		}

		// Put in an empty file named by the timestamp
		_, err = pachClient.PutFile(in.Cron.Repo, "master", next.Format(time.RFC3339), strings.NewReader(""))
		if err != nil {
			return fmt.Errorf("put error %v", err)
		}

		err = pachClient.FinishCommit(in.Cron.Repo, "master")
		if err != nil {
			return err
		}

		// set latestTime to the next time
		latestTime = next
	}
}

func isNilBranchErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "has no head")
}
