package server

import (
	"context"
	"path"
	"time"

	"go.pedge.io/lion/proto"
	"k8s.io/kubernetes/pkg/api"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/dlock"
	"github.com/pachyderm/pachyderm/src/server/pkg/watch"
	ppsserver "github.com/pachyderm/pachyderm/src/server/pps"
)

const (
	masterLockPath = "_master_lock"
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

		protolion.Infof("Launching PPS master process")

		pipelineWatcher, err := a.pipelines.ReadOnly(ctx).WatchWithPrev()
		if err != nil {
			return err
		}
		defer pipelineWatcher.Close()

		for {
			event := <-pipelineWatcher.Watch()
			if event.Err != nil {
				return event.Err
			}
			switch event.Type {
			case watch.EventPut:
				var pipelineName string
				var pipelineInfo pps.PipelineInfo
				if err := event.Unmarshal(&pipelineName, &pipelineInfo); err != nil {
					return err
				}

				var prevPipelineInfo pps.PipelineInfo
				if event.PrevKey != nil {
					if err := event.UnmarshalPrev(&pipelineName, &prevPipelineInfo); err != nil {
						return err
					}
				}

				// If the pipeline has been stopped, delete workers
				if pipelineStateToStopped(pipelineInfo.State) {
					protolion.Infof("master: deleting workers for pipeline %s", pipelineInfo.Pipeline.Name)
					if err := a.deleteWorkersForPipeline(&pipelineInfo); err != nil {
						return err
					}
				}

				// If the pipeline has been restarted, create workers
				if !pipelineStateToStopped(pipelineInfo.State) && event.PrevKey != nil && pipelineStateToStopped(prevPipelineInfo.State) {
					if err := a.upsertWorkersForPipeline(&pipelineInfo); err != nil {
						return err
					}
				}

				// If the pipeline has been updated, create new workers
				if pipelineInfo.Version > prevPipelineInfo.Version {
					protolion.Infof("master: creating/updating workers for pipeline %s", pipelineInfo.Pipeline.Name)
					if event.PrevKey != nil {
						if err := a.deleteWorkersForPipeline(&prevPipelineInfo); err != nil {
							return err
						}
					}
					if err := a.upsertWorkersForPipeline(&pipelineInfo); err != nil {
						return err
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
		}
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		protolion.Errorf("master: error running the master process: %v; retrying in %v", err, d)
		return nil
	})
}

func (a *apiServer) upsertWorkersForPipeline(pipelineInfo *pps.PipelineInfo) error {
	parallelism, err := ppsserver.GetExpectedNumWorkers(a.kubeClient, pipelineInfo.ParallelismSpec)
	if err != nil {
		return err
	}
	var resources *api.ResourceList
	if pipelineInfo.ResourceSpec != nil {
		resources, err = parseResourceList(pipelineInfo.ResourceSpec, pipelineInfo.CacheSize)
		if err != nil {
			return err
		}
	}
	options := a.getWorkerOptions(
		ppsserver.PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version),
		int32(parallelism),
		resources,
		pipelineInfo.Transform,
		pipelineInfo.CacheSize,
	)
	// Set the pipeline name env
	options.workerEnv = append(options.workerEnv, api.EnvVar{
		Name:  client.PPSPipelineNameEnv,
		Value: pipelineInfo.Pipeline.Name,
	})
	return a.createWorkerRc(options)
}

func (a *apiServer) deleteWorkersForPipeline(pipelineInfo *pps.PipelineInfo) error {
	rcName := ppsserver.PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version)
	if err := a.kubeClient.Services(a.namespace).Delete(rcName); err != nil {
		if !isNotFoundErr(err) {
			return err
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
