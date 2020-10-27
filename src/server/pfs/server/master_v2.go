package server

import (
	"path"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/dlock"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/gc"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

const (
	masterLockPath = "pfs-master-lock"
)

func (d *driverV2) master(env *serviceenv.ServiceEnv, objClient obj.Client, db *gorm.DB) {
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
			opts, err := env.GarbageCollectionOptions()
			if err != nil {
				return err
			}
			return gc.Run(masterCtx, objClient, db, opts...)
		}, backoff.NewInfiniteBackOff(), func(err error, _ time.Duration) error {
			log.Errorf("error in pfs master: %v", err)
			return err
		})
		// TODO: Always panic, remove this check, see above.
		if err != nil {
			panic(err)
		}
	}
}
