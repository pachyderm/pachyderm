package worker

import (
	"context"
	"path"
	"time"

	"go.pedge.io/lion/proto"

	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/dlock"
)

const (
	masterLockPath = "_master_worker_lock"
)

func (a *APIServer) master() {
	backoff.RetryNotify(func() error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		masterLock, err := dlock.NewDLock(ctx, a.etcdClient, path.Join(a.etcdPrefix, masterLockPath, a.pipelineInfo.Pipeline.Name))
		if err != nil {
			return err
		}
		defer masterLock.Unlock()
		ctx = masterLock.Context()

		return nil
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		protolion.Errorf("master: error running the master process: %v; retrying in %v", err, d)
		return nil
	})
}
