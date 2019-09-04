package worker

import (
	"context"
	"path"
	"time"

	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/dlock"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsutil"
	"github.com/pachyderm/pachyderm/src/server/worker/logs"
	"github.com/pachyderm/pachyderm/src/server/worker/spawner"
)

const (
	masterLockPath = "_master_worker_lock"
)

func (a *APIServer) master() {
	logger := logs.NewMasterLogger(a.pipelineInfo)

	masterLock := dlock.NewDLock(a.etcdClient, path.Join(a.etcdPrefix, masterLockPath, a.pipelineInfo.Pipeline.Name, a.pipelineInfo.Salt))
	b := backoff.NewInfiniteBackOff()
	// Setting a high backoff so that when this master fails, the other
	// workers are more likely to become the master.
	// Also, we've observed race conditions where StopPipeline would cause
	// a master to restart before it's deleted.  PPS would then get confused
	// by the restart and create the workers again, because the restart would
	// bring the pipeline state from PAUSED to RUNNING.  By setting a high
	// retry interval, the master would be deleted before it gets a chance
	// to restart.
	b.InitialInterval = 10 * time.Second
	backoff.RetryNotify(func() error {
		// We use pachClient.Ctx here because it contains auth information.
		ctx, cancel := context.WithCancel(a.pachClient.Ctx())
		defer cancel() // make sure that everything this loop might spawn gets cleaned up
		ctx, err := masterLock.Lock(ctx)
		pachClient := a.pachClient.WithCtx(ctx)
		if err != nil {
			return err
		}
		defer masterLock.Unlock(ctx)

		// Create a new driver that uses our new cancelable pachClient
		return spawner.Run(pachClient, a.pipelineInfo, logger, a.driver.WithCtx(ctx))
	}, b, func(err error, d time.Duration) error {
		if auth.IsErrNotAuthorized(err) {
			logger.Logf("failing %q due to auth rejection", a.pipelineInfo.Pipeline.Name)
			return ppsutil.FailPipeline(a.pachClient.Ctx(), a.etcdClient, a.driver.Pipelines(),
				a.pipelineInfo.Pipeline.Name, "worker master could not access output "+
					"repo to watch for new commits")
		}
		logger.Logf("master: error running the master process, retrying in %v: %v", d, err)
		return nil
	})
}
