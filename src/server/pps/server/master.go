package server

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/gogo/protobuf/types"
	"github.com/robfig/cron"
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

		log.Infof("Launching PPS master process")

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
					// Retrieve pipelineInfo (and prev pipeline's pipelineInfo) from the
					// spec repo
					var prevPipelinePtr pps.EtcdPipelineInfo
					var pipelineInfo, prevPipelineInfo *pps.PipelineInfo
					if err := a.sudo(pachClient, func(superUserClient *client.APIClient) error {
						var err error
						pipelineInfo, err = ppsutil.GetPipelineInfo(superUserClient, &pipelinePtr)
						if err != nil {
							return err
						}

						if event.PrevKey != nil {
							if err := event.UnmarshalPrev(&pipelineName, &prevPipelinePtr); err != nil {
								return err
							}
							prevPipelineInfo, err = ppsutil.GetPipelineInfo(superUserClient, &prevPipelinePtr)
							if err != nil {
								return err
							}
						}
						return nil
					}); err != nil {
						return fmt.Errorf("watch event had no pipelineInfo: %v", err)
					}

					// If the pipeline has been stopped, delete workers
					if pipelineInfo.Stopped && pipelineInfo.State != pps.PipelineState_PIPELINE_PAUSED {
						log.Infof("PPS master: deleting workers for pipeline %s (%s)", pipelineName, pipelinePtr.State.String())
						if err := a.deleteWorkersForPipeline(pipelineName); err != nil {
							return err
						}
						if err := a.setPipelineState(pachClient, pipelineInfo, pps.PipelineState_PIPELINE_PAUSED, ""); err != nil {
							return err
						}
					}

					var hasGitInput bool
					pps.VisitInput(pipelineInfo.Input, func(input *pps.Input) {
						if input.Git != nil {
							hasGitInput = true
						}
					})

					// True if the pipeline has been restarted (regardless of any change
					// to the pipeline spec)
					pipelineRestarted := !pipelineInfo.Stopped &&
						event.PrevKey != nil && prevPipelineInfo.Stopped
					// True if auth has been activated or deactivated
					authActivationChanged := (pipelinePtr.AuthToken == "") !=
						(prevPipelinePtr.AuthToken == "")
					// True if the pipeline has been created or updated
					pipelineUpserted := func() bool {
						var prevSpecCommit string
						if prevPipelinePtr.SpecCommit != nil {
							prevSpecCommit = prevPipelinePtr.SpecCommit.ID
						}
						return pipelinePtr.SpecCommit.ID != prevSpecCommit &&
							!pipelineInfo.Stopped
					}()
					if pipelineRestarted || authActivationChanged || pipelineUpserted {
						if (pipelineUpserted || authActivationChanged) && event.PrevKey != nil {
							if err := a.deleteWorkersForPipeline(prevPipelineInfo.Pipeline.Name); err != nil {
								return err
							}
						}
						if (pipelineUpserted || pipelineRestarted) && hasGitInput {
							if err := a.checkOrDeployGithookService(); err != nil {
								return err
							}
						}
						log.Infof("PPS master: creating/updating workers for pipeline %s", pipelineName)
						if err := a.upsertWorkersForPipeline(pipelineInfo); err != nil {
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
						if err := a.scaleUpWorkersForPipeline(pipelineInfo); err != nil {
							return err
						}
					}
					if pipelineInfo.State == pps.PipelineState_PIPELINE_STANDBY {
						if err := a.scaleDownWorkersForPipeline(pipelineInfo); err != nil {
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
		for _, c := range a.monitorCancels {
			c()
		}
		a.monitorCancels = make(map[string]func())
		log.Errorf("master: error running the master process: %v; retrying in %v", err, d)
		return nil
	})
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

func (a *apiServer) upsertWorkersForPipeline(pipelineInfo *pps.PipelineInfo) error {
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
		rc := a.env.GetKubeClient().CoreV1().ReplicationControllers(a.namespace)
		workerRc, err := rc.Get(
			ppsutil.PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version),
			metav1.GetOptions{})
		if err != nil {
			if !isNotFoundErr(err) {
				return err
			}
		}
		if workerRc.ObjectMeta.Labels["version"] != version.PrettyVersion() {
			if err := a.deleteWorkersForPipeline(pipelineInfo.Pipeline.Name); err != nil {
				return err
			}
		}

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
			pipelineInfo.PodSpec,
			pipelineInfo.PodPatch)
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
	}); err != nil {
		return err
	}
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

func (a *apiServer) deleteWorkersForPipeline(pipelineName string) error {
	cancel, ok := a.monitorCancels[pipelineName]
	if ok {
		cancel()
		delete(a.monitorCancels, pipelineName)
	}
	kubeClient := a.env.GetKubeClient()
	selector := fmt.Sprintf("pipelineName=%s", pipelineName)
	falseVal := false
	opts := &metav1.DeleteOptions{
		OrphanDependents: &falseVal,
	}
	services, err := kubeClient.CoreV1().Services(a.namespace).List(metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return err
	}
	for _, service := range services.Items {
		if err := kubeClient.CoreV1().Services(a.namespace).Delete(service.Name, opts); err != nil {
			if !isNotFoundErr(err) {
				return err
			}
		}
	}
	rcs, err := kubeClient.CoreV1().ReplicationControllers(a.namespace).List(metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return err
	}
	for _, rc := range rcs.Items {
		if err := kubeClient.CoreV1().ReplicationControllers(a.namespace).Delete(rc.Name, opts); err != nil {
			if !isNotFoundErr(err) {
				return err
			}
		}
	}
	return nil
}

func (a *apiServer) scaleDownWorkersForPipeline(pipelineInfo *pps.PipelineInfo) error {
	rc := a.env.GetKubeClient().CoreV1().ReplicationControllers(a.namespace)
	workerRc, err := rc.Get(
		ppsutil.PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version),
		metav1.GetOptions{})
	if err != nil {
		return err
	}
	*workerRc.Spec.Replicas = 0
	_, err = rc.Update(workerRc)
	return err
}

func (a *apiServer) scaleUpWorkersForPipeline(pipelineInfo *pps.PipelineInfo) error {
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
}

func notifyCtx(ctx context.Context, name string) func(error, time.Duration) error {
	return func(err error, d time.Duration) error {
		select {
		case <-ctx.Done():
			return context.DeadlineExceeded
		default:
			log.Errorf("error in %s: %v: retrying in: %v\n", name, err, d)
		}
		return nil
	}
}

func (a *apiServer) setPipelineState(pachClient *client.APIClient, pipelineInfo *pps.PipelineInfo, state pps.PipelineState, reason string) error {
	log.Infof("moving pipeline %s to %s", pipelineInfo.Pipeline.Name, state.String())
	_, err := col.NewSTM(pachClient.Ctx(), a.env.GetEtcdClient(), func(stm col.STM) error {
		pipelines := a.pipelines.ReadWrite(stm)
		pipelinePtr := &pps.EtcdPipelineInfo{}
		return pipelines.Update(pipelineInfo.Pipeline.Name, pipelinePtr, func() error {
			if pipelinePtr.State == pps.PipelineState_PIPELINE_FAILURE {
				return nil
			}
			pipelinePtr.State = state
			pipelinePtr.Reason = reason
			return nil
		})
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
				if err := a.setPipelineState(pachClient, pipelineInfo, pps.PipelineState_PIPELINE_STANDBY, ""); err != nil {
					return err
				}
				for {
					var ci *pfs.CommitInfo
					select {
					case ci = <-ciChan:
						if ci.Finished != nil {
							continue
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
					fmt.Printf("error in monitorPipeline: %v: retrying in: %v\n", err, d)
				}
				return nil
			})
		})
	}
	if err := eg.Wait(); err != nil {
		fmt.Printf("error in monitorPipeline: %v", err)
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
		time.Sleep(time.Until(next))
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
