package server

import (
	"context"
	"fmt"
	"path"
	"time"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_watch "k8s.io/apimachinery/pkg/watch"
	kube "k8s.io/client-go/kubernetes"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
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
		pachClient.SetAuthToken(a.getPPSToken())
		ctx, err := masterLock.Lock(ctx)
		if err != nil {
			return err
		}
		defer masterLock.Unlock(ctx)

		log.Infof("Launching PPS master process")

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

		for {
			select {
			case event := <-pipelineWatcher.Watch():
				if event.Err != nil {
					return fmt.Errorf("event err: %+v", event.Err)
				}
				switch event.Type {
				case watch.EventPut:
					var pipelineName string
					var pipelinePtr pps.EtcdPipelineInfo
					if err := event.Unmarshal(&pipelineName, &pipelinePtr); err != nil {
						return err
					}
					pipelineInfo, err := ppsutil.GetPipelineInfo(pachClient, pipelineName, &pipelinePtr)
					if err != nil {
						return err
					}

					var prevPipelinePtr pps.EtcdPipelineInfo
					var prevPipelineInfo *pps.PipelineInfo
					if event.PrevKey != nil {
						if err := event.UnmarshalPrev(&pipelineName, &prevPipelinePtr); err != nil {
							return err
						}
						prevPipelineInfo, err = ppsutil.GetPipelineInfo(pachClient, pipelineName, &prevPipelinePtr)
						if err != nil {
							return err
						}
					}

					// If the pipeline has been stopped, delete workers
					if pipelineStateToStopped(pipelinePtr.State) {
						log.Infof("PPS master: deleting workers for pipeline %s (%s)", pipelineName, pipelinePtr.State.String())
						if err := a.deleteWorkersForPipeline(pipelineInfo); err != nil {
							return err
						}
					}

					var hasGitInput bool
					pps.VisitInput(pipelineInfo.Input, func(input *pps.Input) {
						if input.Git != nil {
							hasGitInput = true
						}
					})

					// If the pipeline has been restarted, create workers
					if !pipelineStateToStopped(pipelinePtr.State) && event.PrevKey != nil && pipelineStateToStopped(prevPipelinePtr.State) {
						if hasGitInput {
							if err := a.checkOrDeployGithookService(); err != nil {
								return err
							}
						}
						log.Infof("PPS master: creating/updating workers for restarted pipeline %s", pipelineName)
						if err := a.upsertWorkersForPipeline(pipelineInfo); err != nil {
							if err := a.setPipelineFailure(ctx, pipelineName, fmt.Sprintf("failed to create workers: %s", err.Error())); err != nil {
								return err
							}
							continue
						}
					}

					// If the pipeline has been created or updated, create new workers
					pipelineUpserted := prevPipelinePtr.SpecCommit == nil ||
						pipelinePtr.SpecCommit.ID != prevPipelinePtr.SpecCommit.ID
					if pipelineUpserted && !pipelineStateToStopped(pipelinePtr.State) {
						log.Infof("PPS master: creating/updating workers for new/updated pipeline %s", pipelineName)
						if event.PrevKey != nil {
							if err := a.deleteWorkersForPipeline(prevPipelineInfo); err != nil {
								return err
							}
						}
						if hasGitInput {
							if err := a.checkOrDeployGithookService(); err != nil {
								return err
							}
						}
						if err := a.upsertWorkersForPipeline(pipelineInfo); err != nil {
							if err := a.setPipelineFailure(ctx, pipelineName, fmt.Sprintf("failed to create workers: %s", err.Error())); err != nil {
								return err
							}
							continue
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
		log.Errorf("master: error running the master process: %v; retrying in %v", err, d)
		return nil
	})
}

func (a *apiServer) setPipelineFailure(ctx context.Context, pipelineName string, reason string) error {
	return ppsutil.FailPipeline(ctx, a.etcdClient, a.pipelines, pipelineName, reason)
}

func (a *apiServer) checkOrDeployGithookService() error {
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

func (a *apiServer) upsertWorkersForPipeline(pipelineInfo *pps.PipelineInfo) error {
	var errCount int
	return backoff.RetryNotify(func() error {
		parallelism, err := ppsutil.GetExpectedNumWorkers(a.kubeClient, pipelineInfo.ParallelismSpec)
		if err != nil {
			log.Errorf("error getting number of workers, default to 1 worker: %v", err)
			parallelism = 1
		}
		var resourceRequests *v1.ResourceList
		var resourceLimits *v1.ResourceList
		if pipelineInfo.ResourceRequests != nil {
			resourceRequests, err = ppsutil.GetRequestsResourceListFromPipeline(pipelineInfo)
			if err != nil {
				return err
			}
		}
		if pipelineInfo.ResourceLimits != nil {
			resourceLimits, err = ppsutil.GetLimitsResourceListFromPipeline(pipelineInfo)
			if err != nil {
				return err
			}
		}

		// Retrieve the current state of the RC.  If the RC is scaled down,
		// we want to ensure that it remains scaled down.
		rc := a.kubeClient.CoreV1().ReplicationControllers(a.namespace)
		workerRc, err := rc.Get(
			ppsutil.PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version),
			metav1.GetOptions{})
		if err == nil {
			if (workerRc.Spec.Template.Spec.Containers[0].Resources.Requests == nil) && *workerRc.Spec.Replicas == 1 {
				parallelism = 1
				resourceRequests = nil
				resourceLimits = nil
			}
		}
		// rc was made by a previous version of pachyderm so we delete it
		if workerRc.ObjectMeta.Labels["version"] != version.PrettyVersion() {
			if err := a.deleteWorkersForPipeline(pipelineInfo); err != nil {
				return err
			}
		}

		options := a.getWorkerOptions(
			pipelineInfo.Pipeline.Name,
			ppsutil.PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version),
			int32(parallelism),
			resourceRequests,
			resourceLimits,
			pipelineInfo.Transform,
			pipelineInfo.CacheSize,
			pipelineInfo.Service,
			pipelineInfo.SpecCommit.ID)
		// Set the pipeline name env
		options.workerEnv = append(options.workerEnv, v1.EnvVar{
			Name:  client.PPSPipelineNameEnv,
			Value: pipelineInfo.Pipeline.Name,
		})
		return a.createWorkerRc(options)
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		errCount++
		if errCount >= 3 {
			return err
		}
		log.Errorf("error creating workers for pipeline %v: %v; retrying in %v", pipelineInfo.Pipeline.Name, err, d)
		return nil
	})
}

func (a *apiServer) deleteWorkersForPipeline(pipelineInfo *pps.PipelineInfo) error {
	rcName := ppsutil.PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version)
	if err := a.kubeClient.CoreV1().Services(a.namespace).Delete(
		rcName, &metav1.DeleteOptions{},
	); err != nil {
		if !isNotFoundErr(err) {
			return err
		}
	}
	if pipelineInfo.Service != nil {
		if err := a.kubeClient.CoreV1().Services(a.namespace).Delete(
			rcName+"-user", &metav1.DeleteOptions{},
		); err != nil {
			if !isNotFoundErr(err) {
				return err
			}
		}
	}
	falseVal := false
	deleteOptions := &metav1.DeleteOptions{
		OrphanDependents: &falseVal,
	}
	if err := a.kubeClient.CoreV1().ReplicationControllers(a.namespace).Delete(rcName, deleteOptions); err != nil {
		if !isNotFoundErr(err) {
			return err
		}
	}
	return nil
}

func (a *apiServer) monitorPipeline(pachClient *client.APIClient, pipelineInfo *pps.PipelineInfo) error {
	backoff.RetryNotify(func() error {
		ciChan := make(chan *pfs.CommitInfo)
		go func() {
			if err := pachClient.SubscribeCommitF(pipelineInfo.Pipeline.Name, pipelineInfo.OutputBranch, "", func(ci *pfs.CommitInfo) error {
				ciChan <- ci
				return nil
			}); err != nil {
				fmt.Printf("error from SubscribeCommit in monitorPipeline: %v\n", err)
			}
		}()
		for {
			select {
			case ci := <-ciChan:
				// scale up
				if _, err := pachClient.BlockCommit(ci.Commit.Repo.Name, ci.Commit.ID); err != nil {
					return err
				}
				if _, err := pachClient.InspectJobOutputCommit(ci.Commit.Repo.Name, ci.Commit.ID, true); err != nil {
					return err
				}
			default:
				// scale down
			}
		}
		return nil
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		select {
		case <-pachClient.Ctx().Done():
			return context.DeadlineExceeded
		default:
			fmt.Printf("error in monitorPipeline: %v: retrying in: %v\n", err, d)
		}
		return nil
	})
	return nil
}
