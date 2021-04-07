package server

import (
	"path"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/dlock"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"

	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
)

const (
	masterLockPath = "pfs-master-lock"
)

func (d *driver) master(env *serviceenv.ServiceEnv) {
	ctx := context.Background()
	masterLock := dlock.NewDLock(d.etcdClient, path.Join(d.prefix, masterLockPath))
	err := backoff.RetryNotify(func() error {
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
		if errors.Is(err, context.Canceled) {
			return err
		}
		log.Errorf("error in pfs master: %v", err)
		return nil
	})
	// Never ending backoff should prevent us from getting here, but tests may want this to exit.
	if !errors.Is(err, context.Canceled) {
		panic(err)
	}
}
