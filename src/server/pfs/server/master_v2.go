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
	"golang.org/x/net/context"
)

const (
	masterLockPath = "pfs-master-lock"
)

func (d *driver) master(env *serviceenv.ServiceEnv, objClient obj.Client, db *gorm.DB) {
	masterLock := dlock.NewDLock(d.etcdClient, path.Join(d.prefix, masterLockPath))
	backoff.RetryNotify(func() error {
		ctx, err := masterLock.Lock(context.Background())
		if err != nil {
			return err
		}
		defer masterLock.Unlock(ctx)
		opts, err := gc.ServiceEnvToOptions(env)
		if err != nil {
			return err
		}
		return gc.Run(ctx, objClient, db, opts...)
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		return nil
	})
}
