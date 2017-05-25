package server

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/dlock"
	"github.com/pachyderm/pachyderm/src/server/pkg/watch"

	"go.pedge.io/lion/proto"
	"k8s.io/kubernetes/pkg/api"
)

const (
	masterLockPath = "_master_lock"
)

// This function acquires a lock in etcd and carries out the responsibilities
// of a master.  When running in a cluster, only one pachd node becomes the
// master.
//
// The master node watches for pipeline updates and updates kubernetes
// accordingly by adding/removing/modifying replication controllers.
func (a *apiServer) master() {
	b := backoff.NewInfiniteBackOff()
	backoff.RetryNotify(func() error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		masterLock, err := dlock.NewDLock(ctx, a.etcdClient, path.Join(a.etcdPrefix, masterLockPath))
		if err != nil {
			return err
		}
		defer masterLock.Unlock()
		ctx = masterLock.Context()

		pipelineWatcher, err := a.pipelines.ReadOnly(ctx).Watch()
		if err != nil {
			return err
		}
		defer pipelineWatcher.Close()

		for {
			event, ok := <-pipelineWatcher.Watch()
			if !ok {
				return fmt.Errorf("pipelineWatcher closed unexpectedly")
			}
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
					if err := event.Unmarshal(&pipelineName, &prevPipelineInfo); err != nil {
						return err
					}
				}

				if pipelineStateToStopped(pipelineInfo.State) {
					return a.deleteWorkersForPipeline(&pipelineInfo)
				}

				if pipelineInfo.Version > prevPipelineInfo.Version {
					protolion.Infof("creating/updating workers for pipeline %s", pipelineInfo.Pipeline.Name)
					return a.upsertWorkersForPipeline(&pipelineInfo)
				}
			case watch.EventDelete:
				var pipelineName string
				var pipelineInfo pps.PipelineInfo
				if err := event.UnmarshalPrev(&pipelineName, &pipelineInfo); err != nil {
					return err
				}
				return a.deleteWorkersForPipeline(&pipelineInfo)
			}
		}
	}, b, func(err error, d time.Duration) error {
		protolion.Infof("master process failed; retrying in %s", d)
		return nil
	})
}

// pipelineStateToStopped defines what pipeline states are "stopped"
// states, meaning that pipelines in this state should not be managed
// by pipelineManager
func pipelineStateToStopped(state pps.PipelineState) bool {
	switch state {
	case pps.PipelineState_PIPELINE_STARTING:
		return false
	case pps.PipelineState_PIPELINE_RUNNING:
		return false
	case pps.PipelineState_PIPELINE_RESTARTING:
		return false
	case pps.PipelineState_PIPELINE_STOPPED:
		return true
	case pps.PipelineState_PIPELINE_FAILURE:
		return true
	default:
		panic(fmt.Sprintf("unrecognized pipeline state: %s", state))
	}
}

func (a *apiServer) upsertWorkersForPipeline(pipelineInfo *pps.PipelineInfo) error {
	parallelism, err := GetExpectedNumWorkers(a.kubeClient, pipelineInfo.ParallelismSpec)
	if err != nil {
		return err
	}
	var resources *api.ResourceList
	if pipelineInfo.ResourceSpec != nil {
		resources, err = parseResourceList(pipelineInfo.ResourceSpec)
		if err != nil {
			return err
		}
	}
	options := a.getWorkerOptions(
		PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version),
		int32(parallelism),
		resources,
		pipelineInfo.Transform)
	// Set the pipeline name env
	options.workerEnv = append(options.workerEnv, api.EnvVar{
		Name:  client.PPSPipelineNameEnv,
		Value: pipelineInfo.Pipeline.Name,
	})
	return a.createWorkerRc(options)
}

func (a *apiServer) deleteWorkersForPipeline(pipelineInfo *pps.PipelineInfo) error {
	rcName := PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version)
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
