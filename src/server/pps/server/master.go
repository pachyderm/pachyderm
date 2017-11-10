package server

import (
	"context"
	"fmt"
	"path"
	"time"

	log "github.com/sirupsen/logrus"
	"k8s.io/kubernetes/pkg/api"
	labelspkg "k8s.io/kubernetes/pkg/labels"
	kube_watch "k8s.io/kubernetes/pkg/watch"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/dlock"
	"github.com/pachyderm/pachyderm/src/server/pkg/util"
	"github.com/pachyderm/pachyderm/src/server/pkg/watch"
	ppsserver "github.com/pachyderm/pachyderm/src/server/pps"
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
		kubePipelineWatch, err := a.kubeClient.Pods(a.namespace).Watch(api.ListOptions{
			LabelSelector: labelspkg.SelectorFromSet(map[string]string{
				"component": "worker",
			}),
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
					var pipelineInfo pps.PipelineInfo
					if err := event.Unmarshal(&pipelineName, &pipelineInfo); err != nil {
						return err
					}

					if pipelineInfo.Salt == "" {
						if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
							pipelines := a.pipelines.ReadWrite(stm)
							newPipelineInfo := new(pps.PipelineInfo)
							if err := pipelines.Get(pipelineInfo.Pipeline.Name, newPipelineInfo); err != nil {
								return fmt.Errorf("error getting pipeline %s: %+v", pipelineName, err)
							}
							if newPipelineInfo.Salt == "" {
								newPipelineInfo.Salt = uuid.NewWithoutDashes()
							}
							pipelines.Put(pipelineInfo.Pipeline.Name, newPipelineInfo)
							return nil
						}); err != nil {
							return err
						}
					}

					var prevPipelineInfo pps.PipelineInfo
					if event.PrevKey != nil {
						if err := event.UnmarshalPrev(&pipelineName, &prevPipelineInfo); err != nil {
							return err
						}
					}

					// If the pipeline has been stopped, delete workers
					if pipelineStateToStopped(pipelineInfo.State) {
						log.Infof("master: deleting workers for pipeline %s", pipelineInfo.Pipeline.Name)
						if err := a.deleteWorkersForPipeline(&pipelineInfo); err != nil {
							return err
						}
					}

					// If the pipeline has been restarted, create workers
					if !pipelineStateToStopped(pipelineInfo.State) && event.PrevKey != nil && pipelineStateToStopped(prevPipelineInfo.State) {
						if err := a.upsertWorkersForPipeline(&pipelineInfo); err != nil {
							if err := a.setPipelineFailure(ctx, pipelineInfo.Pipeline.Name, fmt.Sprintf("failed to create workers: %s", err.Error())); err != nil {
								return err
							}
							continue
						}
					}

					// If the pipeline has been updated, create new workers
					if pipelineInfo.Version > prevPipelineInfo.Version && !pipelineStateToStopped(pipelineInfo.State) {
						log.Infof("master: creating/updating workers for pipeline %s", pipelineInfo.Pipeline.Name)
						if event.PrevKey != nil {
							if err := a.deleteWorkersForPipeline(&prevPipelineInfo); err != nil {
								return err
							}
						}
						if err := a.upsertWorkersForPipeline(&pipelineInfo); err != nil {
							if err := a.setPipelineFailure(ctx, pipelineInfo.Pipeline.Name, fmt.Sprintf("failed to create workers: %s", err.Error())); err != nil {
								return err
							}
							continue
						}
					}
				case watch.EventDelete:
					var pipelineName string
					var pipelineInfo pps.PipelineInfo
					if err := event.UnmarshalPrev(&pipelineName, &pipelineInfo); err != nil {
						return err
					}
					if err := a.deleteWorkersForPipeline(&pipelineInfo); err != nil {
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
					kubePipelineWatch, err = a.kubeClient.Pods(a.namespace).Watch(api.ListOptions{
						LabelSelector: labelspkg.SelectorFromSet(map[string]string{
							"component": "worker",
						}),
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
				pod, ok := event.Object.(*api.Pod)
				if !ok {
					continue
				}
				if pod.Status.Phase == api.PodFailed {
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
	// Set pipeline state to failure
	_, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		pipelines := a.pipelines.ReadWrite(stm)
		pipelineInfo := new(pps.PipelineInfo)
		if err := pipelines.Get(pipelineName, pipelineInfo); err != nil {
			return err
		}
		pipelineInfo.State = pps.PipelineState_PIPELINE_FAILURE
		pipelineInfo.Reason = reason
		pipelines.Put(pipelineName, pipelineInfo)
		return nil
	})
	return err
}

func (a *apiServer) upsertWorkersForPipeline(pipelineInfo *pps.PipelineInfo) error {
	var errCount int
	return backoff.RetryNotify(func() error {
		parallelism, err := ppsserver.GetExpectedNumWorkers(a.kubeClient, pipelineInfo.ParallelismSpec)
		if err != nil {
			log.Errorf("error getting number of workers, default to 1 worker: %v", err)
			parallelism = 1
		}
		var resourceRequests *api.ResourceList
		var resourceLimits *api.ResourceList
		if pipelineInfo.ResourceRequestsSpec != nil {
			resourceRequests, err = util.GetRequestsResourceListFromPipeline(pipelineInfo)
			if err != nil {
				return err
			}
		}
		if pipelineInfo.ResourceLimitsSpec != nil {
			resourceLimits, err = util.GetLimitsResourceListFromPipeline(pipelineInfo)
			if err != nil {
				return err
			}
		}

		// Retrieve the current state of the RC.  If the RC is scaled down,
		// we want to ensure that it remains scaled down.
		rc := a.kubeClient.ReplicationControllers(a.namespace)
		workerRc, err := rc.Get(ppsserver.PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version))
		if err == nil {
			if (workerRc.Spec.Template.Spec.Containers[0].Resources.Requests == nil) && workerRc.Spec.Replicas == 1 {
				parallelism = 1
				resourceRequests = nil
				resourceLimits = nil
			}
		}

		options := a.getWorkerOptions(
			pipelineInfo.Pipeline.Name,
			ppsserver.PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version),
			int32(parallelism),
			resourceRequests,
			resourceLimits,
			pipelineInfo.Transform,
			pipelineInfo.CacheSize,
			pipelineInfo.Service)
		// Set the pipeline name env
		options.workerEnv = append(options.workerEnv, api.EnvVar{
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
	rcName := ppsserver.PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version)
	if err := a.kubeClient.Services(a.namespace).Delete(rcName); err != nil {
		if !isNotFoundErr(err) {
			return err
		}
	}
	if pipelineInfo.Service != nil {
		if err := a.kubeClient.Services(a.namespace).Delete(rcName + "-user"); err != nil {
			if !isNotFoundErr(err) {
				return err
			}
		}
	}
	falseVal := false
	deleteOptions := &api.DeleteOptions{
		OrphanDependents: &falseVal,
	}
	if err := a.kubeClient.ReplicationControllers(a.namespace).Delete(rcName, deleteOptions); err != nil {
		if !isNotFoundErr(err) {
			return err
		}
	}
	return nil
}
