package server

import (
	"github.com/pachyderm/pachyderm/src/server/pkg/migrations"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/track"
	"golang.org/x/net/context"
)

var desiredClusterState migrations.State = migrations.InitialState().
	Apply(func(ctx context.Context, env migrations.Env) error {
		_, err := env.Tx.ExecContext(ctx, `CREATE SCHEMA storage`)
		return err
	}).
	Apply(func(ctx context.Context, env migrations.Env) error {
		if err := track.SetupPostgresTrackerV0(ctx, env.Tx); err != nil {
			return err
		}
		if err := chunk.SetupPostgresStoreV0(ctx, "storage.chunks", env.Tx); err != nil {
			return err
		}
		if err := fileset.SetupPostgresStoreV0(ctx, env.Tx); err != nil {
			return err
		}
		return nil
	})
