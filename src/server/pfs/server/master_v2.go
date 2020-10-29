package server

import (
	"path"
	"time"

	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/dlock"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

const (
	masterLockPath = "pfs-master-lock"
)

func (d *driverV2) master(env *serviceenv.ServiceEnv) {
	ctx := context.Background()
	masterLock := dlock.NewDLock(d.etcdClient, path.Join(d.prefix, masterLockPath))
	// TODO: We shouldn't need to wrap this in a for loop, but it looks like gc.Run
	// can return nil. Which then ends the RetryNotify loop.
	for {
		err := backoff.RetryNotify(func() error {
			masterCtx, err := masterLock.Lock(ctx)
			if err != nil {
				return err
			}
			defer masterLock.Unlock(masterCtx)
			return d.storage.GC(masterCtx)
		}, backoff.NewInfiniteBackOff(), func(err error, _ time.Duration) error {
			log.Errorf("error in pfs master: %v", err)
			return nil
		})
		// TODO: Always panic, remove this check, see above.
		if err != nil {
			panic(err)
		}
	}
}
