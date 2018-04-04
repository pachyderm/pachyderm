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

					// True if the pipeline has been restarted (regardless of any change
					// to the pipeline spec)
					pipelineRestarted := !pipelineStateToStopped(pipelinePtr.State) &&
						event.PrevKey != nil && pipelineStateToStopped(prevPipelinePtr.State)
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
							!pipelineStateToStopped(pipelinePtr.State)
					}()
					if pipelineRestarted || authActivationChanged || pipelineUpserted {
						if (pipelineUpserted || authActivationChanged) && event.PrevKey != nil {
							if err := a.deleteWorkersForPipeline(prevPipelineInfo); err != nil {
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
							if err := a.setPipelineFailure(ctx, pipelineName, fmt.Sprintf("failed to create workers: %s", err.Error())); err != nil {
								return err
							}
							continue
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
		rc := a.kubeClient.CoreV1().ReplicationControllers(a.namespace)
		workerRc, err := rc.Get(
			ppsutil.PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version),
			metav1.GetOptions{})
		if err != nil {
			log.Errorf("error from rc.Get: %v", err)
		}
		// TODO figure out why the statement below runs even if there's an error
		// rc was made by a previous version of pachyderm so we delete it
		if workerRc.ObjectMeta.Labels["version"] != version.PrettyVersion() {
			if err := a.deleteWorkersForPipeline(pipelineInfo); err != nil {
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

func (a *apiServer) deleteWorkersForPipeline(pipelineInfo *pps.PipelineInfo) error {
	cancel, ok := a.monitorCancels[pipelineInfo.Pipeline.Name]
	if ok {
		cancel()
		delete(a.monitorCancels, pipelineInfo.Pipeline.Name)
	}
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

func (a *apiServer) scaleDownWorkersForPipeline(pipelineInfo *pps.PipelineInfo) error {
	rc := a.kubeClient.CoreV1().ReplicationControllers(a.namespace)
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
			log.Errorf("error in %s: %v: retrying in: %v\n", name, err, d)
		}
		return nil
	}
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
	if pipelineInfo.NoStandby {
		// NoStandby so simply put it in RUNNING and leave it there.  This is
		// only done with eg.Go so that we can handle all the errors in the
		// same way below, it should be a very quick operation so there's no
		// good reason to do it concurrently.
		eg.Go(func() error {
			return backoff.RetryNotify(func() error {
				_, err := col.NewSTM(pachClient.Ctx(), a.etcdClient, func(stm col.STM) error {
					pipelines := a.pipelines.ReadWrite(stm)
					pipelinePtr := &pps.EtcdPipelineInfo{}
					return pipelines.Update(pipelineInfo.Pipeline.Name, pipelinePtr, func() error {
						if pipelinePtr.State == pps.PipelineState_PIPELINE_PAUSED {
							return nil
						}
						pipelinePtr.State = pps.PipelineState_PIPELINE_RUNNING
						return nil
					})
				})
				return err
			}, backoff.NewInfiniteBackOff(), notifyCtx(pachClient.Ctx(), "set running (NoStandby)"))
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
			setPipelineState := func(state pps.PipelineState) error {
				log.Infof("moving pipeline %s to %s", pipelineInfo.Pipeline.Name, state.String())
				_, err := col.NewSTM(pachClient.Ctx(), a.etcdClient, func(stm col.STM) error {
					pipelines := a.pipelines.ReadWrite(stm)
					pipelinePtr := &pps.EtcdPipelineInfo{}
					return pipelines.Update(pipelineInfo.Pipeline.Name, pipelinePtr, func() error {
						if pipelinePtr.State == pps.PipelineState_PIPELINE_PAUSED || pipelinePtr.State == pps.PipelineState_PIPELINE_FAILURE {
							return nil
						}
						pipelinePtr.State = state
						return nil
					})
				})
				return err
			}
			return backoff.RetryNotify(func() error {
				standby := false
				if pipelineInfo.NoStandby {
					if err := setPipelineState(pps.PipelineState_PIPELINE_RUNNING); err != nil {
						return err
					}
				} else {
					if err := setPipelineState(pps.PipelineState_PIPELINE_STANDBY); err != nil {
						return err
					}
					standby = true
				}
				for {
					var ci *pfs.CommitInfo
					if pipelineInfo.NoStandby {
						// Standy is disabled, just block for a commit.
						select {
						case ci = <-ciChan:
						case <-pachClient.Ctx().Done():
							return context.DeadlineExceeded
						}
					} else if standby {
						// Pipelines is already in standby, so we wait for a
						// new commit, or for our context deadline to be
						// exceeded.
						select {
						case ci = <-ciChan:
							if ci.Finished != nil {
								continue
							}
							if err := setPipelineState(pps.PipelineState_PIPELINE_RUNNING); err != nil {
								return err
							}
							standby = false
						case <-pachClient.Ctx().Done():
							return context.DeadlineExceeded
						}
					} else {
						// Pipeline is running, if we get a new commit we can
						// proceed to processing it immediately. If we don't
						// well talk the default branch and put the pipeline in
						// standby.
						select {
						case ci = <-ciChan:
						case <-pachClient.Ctx().Done():
							return context.DeadlineExceeded
						default:
							if err := setPipelineState(pps.PipelineState_PIPELINE_STANDBY); err != nil {
								return err
							}
							standby = true
							continue
						}
					}
					// Wait for the commit to be finished before blocking on the
					// job because the job may not exist yet.
					if _, err := pachClient.BlockCommit(ci.Commit.Repo.Name, ci.Commit.ID); err != nil {
						return err
					}
					if _, err := pachClient.InspectJobOutputCommit(ci.Commit.Repo.Name, ci.Commit.ID, true); err != nil {
						return err
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
// a helper function called by jobSpawner
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
		if _, err := pachClient.StartCommit(in.Cron.Repo, "master"); err != nil {
			return err
		}
		timestamp, err := types.TimestampProto(t)
		if err != nil {
			return err
		}
		timeString, err := (&jsonpb.Marshaler{}).MarshalToString(timestamp)
		if err != nil {
			return err
		}
		if err := pachClient.DeleteFile(in.Cron.Repo, "master", "time"); err != nil {
			return err
		}
		if _, err := pachClient.PutFile(in.Cron.Repo, "master", "time", strings.NewReader(timeString)); err != nil {
			return err
		}
		if err := pachClient.FinishCommit(in.Cron.Repo, "master"); err != nil {
			return err
		}
	}
}

func isNilBranchErr(err error) bool {
	return err != nil &&
		strings.HasPrefix(err.Error(), "the branch \"") &&
		strings.HasSuffix(err.Error(), "\" is nil")
}
