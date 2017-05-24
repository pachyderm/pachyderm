package server

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/dlock"
	"github.com/pachyderm/pachyderm/src/server/pkg/watch"

	"go.pedge.io/lion/proto"
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

		pipelineWatcher, err := a.pipelines.ReadOnly(ctx).WatchByIndex(stoppedIndex, false)
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
			pipelineName := string(event.Key)
			switch event.Type {
			case watch.EventPut:
				var pipelineInfo pps.PipelineInfo
				if err := event.Unmarshal(&pipelineName, &pipelineInfo); err != nil {
					return err
				}
				if pipelineInfo.Input == nil {
					pipelineInfo.Input = translatePipelineInputs(pipelineInfo.Inputs)
				}
				protolion.Infof("creating/updating workers for pipeline %s", pipelineInfo.Pipeline.Name)
				if err != nil {
					return a.upsertWorkersForPipeline(ctx, pipelineInfo)
				}
			case watch.EventDelete:
				if err != nil {
					return a.deleteWorkersForPipeline(ctx, pipelineName)
				}
			}
		}
	}, b, func(err error, d time.Duration) error {
		protolion.Infof("master process failed; retrying in %s", d)
		return nil
	})
}
