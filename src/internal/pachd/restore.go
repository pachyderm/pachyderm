package pachd

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/recovery"
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
				//tracker := track.NewPostgresTracker(env.DB)

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
				s := &recovery.Snapshotter{DB: env.DB, Storage: storage.Filesets}
				snapshotID := recovery.SnapshotID(config.SnapshotID)
				if err := s.RestoreSnapshot(ctx, snapshotID, recovery.RestoreSnapshotOptions{}); err != nil {
					return errors.Wrapf(err, "could not restore snapshot %s", snapshotID)
				}
				return nil
			},
		},
	)
	return rs
}
