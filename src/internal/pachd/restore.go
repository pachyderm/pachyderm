package pachd

import (
	"context"
	"os"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/pachyderm/pachyderm/v2/src/internal/admindb"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/snapshot"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage"
)

type RestoreSnapshotEnv struct {
	// DB should be a direct database connection.
	DB *sqlx.DB
}

// RestoreSnapshot is a minimal pachd for restoring from a snapshot.
type RestoreSnapshot struct {
	base
	env    RestoreSnapshotEnv
	config pachconfig.PachdRestoreSnapshotConfiguration
}

func NewRestoreSnapshot(env RestoreSnapshotEnv, config pachconfig.PachdRestoreSnapshotConfiguration) *RestoreSnapshot {
	rs := &RestoreSnapshot{env: env, config: config}
	rs.addSetup(
		printVersion(),
		awaitDB(env.DB),
		setupStep{
			Name: "restore",
			Fn: func(ctx context.Context) error {
				bucket, err := obj.NewBucket(ctx, config.StorageBackend, config.StorageRoot, config.StorageURL)
				if err != nil {
					return errors.Wrap(err, "storage env")
				}
				sEnv := storage.Env{
					Bucket: bucket,
					DB:     env.DB,
					Config: config.StorageConfiguration,
				}
				storage, err := storage.New(ctx, sEnv)
				if err != nil {
					return errors.Wrapf(err, "could not configure new storage")
				}
				s := &snapshot.Snapshotter{DB: env.DB, Storage: storage.Filesets}
				snapshotID := snapshot.SnapshotID(config.SnapshotID)
				if err := s.RestoreSnapshot(ctx, snapshotID, snapshot.RestoreSnapshotOptions{}); err != nil {
					return errors.Wrapf(err, "could not restore snapshot %s", snapshotID)
				}
				return errors.Wrap(dbutil.WithTx(ctx, env.DB, func(ctx context.Context, tx *pachsql.Tx) error {
					hostname, err := os.Hostname()
					if err != nil {
						hostname = "un-named pod"
					}
					return errors.Wrap(admindb.ScheduleRestart(ctx, tx, time.Now(), "restored Pachyderm", hostname), "ScheduleRestart")
				}), "WithTx")
			},
		},
	)
	return rs
}
