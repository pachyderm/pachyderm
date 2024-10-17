package pachd

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
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
				log.Info(ctx, "Hello, you would have just restored a snapshot")
				return nil
			},
		},
	)
	return rs
}
