package pachd

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/recovery"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/kv"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
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
			Name: "HelloWorld",
			Fn: func(ctx context.Context) error {
				tracker := track.NewPostgresTracker(env.DB)
				storage := fileset.NewStorage(fileset.NewPostgresStore(env.DB), tracker, chunk.NewStorage(kv.NewMemStore(), nil, env.DB, tracker))
				s := &recovery.Snapshotter{DB: env.DB, Storage: storage}
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
