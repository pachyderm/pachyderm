package clusterstate

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	"github.com/pachyderm/pachyderm/v2/src/server/license"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs/server/"
)

// DesiredClusterState is the set of migrations to apply to run pachd at the current version.
// New migrations should be appended to the end.
var DesiredClusterState migrations.State = migrations.InitialState().
	Apply("create storage schema", func(ctx context.Context, env migrations.Env) error {
		_, err := env.Tx.ExecContext(ctx, `CREATE SCHEMA storage`)
		return err
	}).
	Apply("storage tracker v0", func(ctx context.Context, env migrations.Env) error {
		return track.SetupPostgresTrackerV0(ctx, env.Tx)
	}).
	Apply("storage chunk store v0", func(ctx context.Context, env migrations.Env) error {
		return chunk.SetupPostgresStoreV0(ctx, "storage.chunks", env.Tx)
	}).
	Apply("storage fileset store v0", func(ctx context.Context, env migrations.Env) error {
		return fileset.SetupPostgresStoreV0(ctx, env.Tx)
	}).
	Apply("create license schema", func(ctx context.Context, env migrations.Env) error {
		_, err := env.Tx.ExecContext(ctx, `CREATE SCHEMA license`)
		return err
	}).
	Apply("license clusters v0", func(ctx context.Context, env migrations.Env) error {
		return license.CreateClustersTable(ctx, env.Tx)
	Apply("create pfs schema", func(ctx context.Context, env migrations.Env) error {
		_, err := env.Tx.ExecContext(ctx, `CREATE SCHEMA pfs`)
		return err
	}).
	Apply("pfs commit store v0", func(ctx context.Context, env migrations.Env) error {
		return pfsserver.SetupPostgresCommitStoreV0(ctx, env.Tx)
	})
