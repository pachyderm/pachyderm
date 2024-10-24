// Package setupenv manages creating pachd.*Envs from pachconfig objects.
package setupenv

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachd"
)

func NewPreflightEnv(ctx context.Context, config pachconfig.PachdPreflightConfiguration) (*pachd.PreFlightEnv, error) {
	db, err := openDirectDB(ctx, config.PostgresConfiguration)
	if err != nil {
		return nil, err
	}
	return &pachd.PreFlightEnv{DB: db}, nil
}

func NewRestoreSnapshotEnv(ctx context.Context, config pachconfig.PachdRestoreSnapshotConfiguration) (*pachd.RestoreSnapshotEnv, error) {
	db, err := openDirectDB(ctx, config.PostgresConfiguration)
	if err != nil {
		return nil, err
	}
	return &pachd.RestoreSnapshotEnv{DB: db}, nil
}
