package server

import (
	"context"
	"path"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/dlock"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

const (
	masterLockPath = "pfs-master-lock"
)

func (d *driver) master(ctx context.Context) {
	masterLock := dlock.NewDLock(d.etcdClient, path.Join(d.prefix, masterLockPath))
	backoff.RetryUntilCancel(ctx, func() error {
		masterCtx, err := masterLock.Lock(ctx)
		if err != nil {
			return err
		}
		defer masterLock.Unlock(masterCtx)
		eg, ctx := errgroup.WithContext(masterCtx)
		eg.Go(func() error {
			return d.storage.GC(ctx)
		})
		eg.Go(func() error {
			gc := chunk.NewGC(d.storage.ChunkStorage())
			return gc.RunForever(ctx)
		})
		return eg.Wait()
	}, backoff.NewInfiniteBackOff(), func(err error, _ time.Duration) error {
		log.Errorf("error in pfs master: %v", err)
		return nil
	})
}
