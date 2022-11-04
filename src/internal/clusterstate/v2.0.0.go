package clusterstate

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/authdb"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactiondb"
	"github.com/pachyderm/pachyderm/v2/src/server/auth"
	"github.com/pachyderm/pachyderm/v2/src/server/identity"
	"github.com/pachyderm/pachyderm/v2/src/server/license"
	licenseserver "github.com/pachyderm/pachyderm/v2/src/server/license/server"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs/server"
)

// DO NOT MODIFY THIS STATE
// IT HAS ALREADY SHIPPED IN A RELEASE
var state_2_0_0 migrations.State = migrations.InitialState().
	Apply("create storage schema", func(ctx context.Context, env migrations.Env) error {
		_, err := env.Tx.ExecContext(ctx, `CREATE SCHEMA storage`)
		return errors.EnsureStack(err)
	}).
	Apply("storage tracker v0", func(ctx context.Context, env migrations.Env) error {
		return track.SetupPostgresTrackerV0(ctx, env.Tx)
	}).
	Apply("storage chunk store v0", func(ctx context.Context, env migrations.Env) error {
		return chunk.SetupPostgresStoreV0(env.Tx)
	}).
	Apply("storage fileset store v0", func(ctx context.Context, env migrations.Env) error {
		return fileset.SetupPostgresStoreV0(ctx, env.Tx)
	}).
	Apply("create license schema", func(ctx context.Context, env migrations.Env) error {
		_, err := env.Tx.ExecContext(ctx, `CREATE SCHEMA license`)
		return errors.EnsureStack(err)
	}).
	Apply("license clusters v0", func(ctx context.Context, env migrations.Env) error {
		return license.CreateClustersTableV0(ctx, env.Tx)
	}).
	Apply("create pfs schema", func(ctx context.Context, env migrations.Env) error {
		_, err := env.Tx.ExecContext(ctx, `CREATE SCHEMA pfs`)
		return errors.EnsureStack(err)
	}).
	Apply("pfs commit store v0", func(ctx context.Context, env migrations.Env) error {
		return pfsserver.SetupPostgresCommitStoreV0(ctx, env.Tx)
	}).
	Apply("create identity schema", func(ctx context.Context, env migrations.Env) error {
		_, err := env.Tx.ExecContext(ctx, `CREATE SCHEMA identity`)
		return errors.EnsureStack(err)
	}).
	Apply("create identity users table v0", func(ctx context.Context, env migrations.Env) error {
		return identity.CreateUsersTable(ctx, env.Tx)
	}).
	Apply("create identity config table v0", func(ctx context.Context, env migrations.Env) error {
		return identity.CreateConfigTable(ctx, env.Tx)
	}).
	Apply("create auth schema", func(ctx context.Context, env migrations.Env) error {
		_, err := env.Tx.ExecContext(ctx, `CREATE SCHEMA auth`)
		return errors.EnsureStack(err)
	}).
	Apply("create auth tokens table v0", func(ctx context.Context, env migrations.Env) error {
		return auth.CreateAuthTokensTable(ctx, env.Tx)
	}).
	Apply("license clusters v1", func(ctx context.Context, env migrations.Env) error {
		return license.AddUserContextsToClustersTable(ctx, env.Tx)
	}).
	Apply("create collections schema", func(ctx context.Context, env migrations.Env) error {
		return col.CreatePostgresSchema(ctx, env.Tx)
	}).
	Apply("create collections trigger functions", func(ctx context.Context, env migrations.Env) error {
		return col.SetupPostgresV0(ctx, env.Tx)
	}).
	Apply("create collections", func(ctx context.Context, env migrations.Env) error {
		collections := []col.PostgresCollection{}
		collections = append(collections, pfsdb.CollectionsV0()...)
		collections = append(collections, ppsdb.CollectionsV0()...)
		collections = append(collections, transactiondb.CollectionsV0()...)
		collections = append(collections, authdb.CollectionsV0()...)
		return col.SetupPostgresCollections(ctx, env.Tx, collections...)
	}).
	Apply("license clusters client_id column", func(ctx context.Context, env migrations.Env) error {
		return license.AddClusterClientIdColumn(ctx, env.Tx)
	}).
	Apply("identity config token lifetime", func(ctx context.Context, env migrations.Env) error {
		return identity.AddTokenExpiryConfig(ctx, env.Tx)
	}).
	Apply("create license collection", func(ctx context.Context, env migrations.Env) error {
		collections := []col.PostgresCollection{}
		collections = append(collections, licenseserver.CollectionsV0()...)
		return col.SetupPostgresCollections(ctx, env.Tx, collections...)
	}).
	Apply("add rotation_token_expiry to identity.config table", func(ctx context.Context, env migrations.Env) error {
		return identity.AddRotationTokenExpiryConfig(ctx, env.Tx)
	})
	// DO NOT MODIFY THIS STATE
	// IT HAS ALREADY SHIPPED IN A RELEASE
