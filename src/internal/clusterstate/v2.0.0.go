package clusterstate

import (
	"context"

	"google.golang.org/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/internal/authdb"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactiondb"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/server/auth"
	"github.com/pachyderm/pachyderm/v2/src/server/identity"
	"github.com/pachyderm/pachyderm/v2/src/server/license"
	licenseserver "github.com/pachyderm/pachyderm/v2/src/server/license/server"
)

// DO NOT MODIFY THIS STATE
// IT HAS ALREADY SHIPPED IN A RELEASE
var state_2_0_0 migrations.State = migrations.InitialState().
	Apply("create storage schema", func(ctx context.Context, env migrations.Env) error {
		_, err := env.Tx.ExecContext(ctx, `CREATE SCHEMA storage`)
		return errors.EnsureStack(err)
	}, migrations.Squash).
	Apply("storage tracker v0", func(ctx context.Context, env migrations.Env) error {
		return track.SetupPostgresTrackerV0(ctx, env.Tx)
	}, migrations.Squash).
	Apply("storage chunk store v0", func(ctx context.Context, env migrations.Env) error {
		return chunk.SetupPostgresStoreV0(ctx, env.Tx)
	}, migrations.Squash).
	Apply("storage fileset store v0", func(ctx context.Context, env migrations.Env) error {
		return fileset.SetupPostgresStoreV0(ctx, env.Tx)
	}, migrations.Squash).
	Apply("create license schema", func(ctx context.Context, env migrations.Env) error {
		_, err := env.Tx.ExecContext(ctx, `CREATE SCHEMA license`)
		return errors.EnsureStack(err)
	}, migrations.Squash).
	Apply("license clusters v0", func(ctx context.Context, env migrations.Env) error {
		return license.CreateClustersTableV0(ctx, env.Tx)
	}, migrations.Squash).
	Apply("create pfs schema", func(ctx context.Context, env migrations.Env) error {
		_, err := env.Tx.ExecContext(ctx, `CREATE SCHEMA pfs`)
		return errors.EnsureStack(err)
	}, migrations.Squash).
	Apply("pfs commit store v0", func(ctx context.Context, env migrations.Env) error {
		return SetupPostgresCommitStoreV0(ctx, env.Tx)
	}, migrations.Squash).
	Apply("create identity schema", func(ctx context.Context, env migrations.Env) error {
		_, err := env.Tx.ExecContext(ctx, `CREATE SCHEMA identity`)
		return errors.EnsureStack(err)
	}, migrations.Squash).
	Apply("create identity users table v0", func(ctx context.Context, env migrations.Env) error {
		return identity.CreateUsersTable(ctx, env.Tx)
	}, migrations.Squash).
	Apply("create identity config table v0", func(ctx context.Context, env migrations.Env) error {
		return identity.CreateConfigTable(ctx, env.Tx)
	}, migrations.Squash).
	Apply("create auth schema", func(ctx context.Context, env migrations.Env) error {
		_, err := env.Tx.ExecContext(ctx, `CREATE SCHEMA auth`)
		return errors.EnsureStack(err)
	}, migrations.Squash).
	Apply("create auth tokens table v0", func(ctx context.Context, env migrations.Env) error {
		return auth.CreateAuthTokensTable(ctx, env.Tx)
	}, migrations.Squash).
	Apply("license clusters v1", func(ctx context.Context, env migrations.Env) error {
		return license.AddUserContextsToClustersTable(ctx, env.Tx)
	}, migrations.Squash).
	Apply("create collections schema", func(ctx context.Context, env migrations.Env) error {
		return col.CreatePostgresSchema(ctx, env.Tx)
	}, migrations.Squash).
	Apply("create collections trigger functions", func(ctx context.Context, env migrations.Env) error {
		return col.SetupPostgresV0(ctx, env.Tx)
	}, migrations.Squash).
	Apply("create collections", func(ctx context.Context, env migrations.Env) error {
		collections := []col.PostgresCollection{}
		collections = append(collections, collectionsV0()...)
		collections = append(collections, ppsdb.CollectionsV0()...)
		collections = append(collections, transactiondb.CollectionsV0()...)
		collections = append(collections, authdb.CollectionsV0()...)
		return col.SetupPostgresCollections(ctx, env.Tx, collections...)
	}, migrations.Squash).
	Apply("license clusters client_id column", func(ctx context.Context, env migrations.Env) error {
		return license.AddClusterClientIdColumn(ctx, env.Tx)
	}, migrations.Squash).
	Apply("identity config token lifetime", func(ctx context.Context, env migrations.Env) error {
		return identity.AddTokenExpiryConfig(ctx, env.Tx)
	}, migrations.Squash).
	Apply("create license collection", func(ctx context.Context, env migrations.Env) error {
		collections := []col.PostgresCollection{}
		collections = append(collections, licenseserver.CollectionsV0()...)
		return col.SetupPostgresCollections(ctx, env.Tx, collections...)
	}, migrations.Squash).
	Apply("add rotation_token_expiry to identity.config table", func(ctx context.Context, env migrations.Env) error {
		return identity.AddRotationTokenExpiryConfig(ctx, env.Tx)
	}, migrations.Squash)

// DO NOT MODIFY THIS STATE
// IT HAS ALREADY SHIPPED IN A RELEASE

// SetupPostgresCommitStoreV0 runs SQL to setup the commit store.
// DO NOT MODIFY THIS FUNCTION
// IT HAS BEEN USED IN A RELEASED MIGRATION
func SetupPostgresCommitStoreV0(ctx context.Context, tx *pachsql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		CREATE TABLE pfs.commit_diffs (
			commit_id TEXT NOT NULL,
			num BIGSERIAL NOT NULL,
			fileset_id UUID NOT NULL,
			PRIMARY KEY(commit_id, num)
		);

		CREATE TABLE pfs.commit_totals (
			commit_id TEXT NOT NULL,
			fileset_id UUID NOT NULL,
			PRIMARY KEY(commit_id)
		);
	`)
	return errors.EnsureStack(err)
}

var repoTypeIndex = &col.Index{
	Name: "type",
	Extract: func(val proto.Message) string {
		return val.(*pfs.RepoInfo).Repo.Type
	},
}

func ReposNameKey(repo *pfs.Repo) string {
	return repo.Project.Name + "/" + repo.Name
}

var reposNameIndex = &col.Index{
	Name: "name",
	Extract: func(val proto.Message) string {
		return ReposNameKey(val.(*pfs.RepoInfo).Repo)
	},
}

var reposIndexes = []*col.Index{reposNameIndex, repoTypeIndex}

func RepoKey(repo *pfs.Repo) string {
	return repo.Project.Name + "/" + repo.Name + "." + repo.Type
}

var commitsRepoIndex = &col.Index{
	Name: "repo",
	Extract: func(val proto.Message) string {
		return RepoKey(val.(*pfs.CommitInfo).Commit.Repo)
	},
}

var commitsBranchlessIndex = &col.Index{
	Name: "branchless",
	Extract: func(val proto.Message) string {
		return commitKey(val.(*pfs.CommitInfo).Commit)
	},
}

var commitsCommitSetIndex = &col.Index{
	Name: "commitset",
	Extract: func(val proto.Message) string {
		return val.(*pfs.CommitInfo).Commit.Id
	},
}

var commitsIndexes = []*col.Index{commitsRepoIndex, commitsBranchlessIndex, commitsCommitSetIndex}

func commitKey(commit *pfs.Commit) string {
	return RepoKey(commit.Repo) + "@" + commit.Id
}

var branchesRepoIndex = &col.Index{
	Name: "repo",
	Extract: func(val proto.Message) string {
		return RepoKey(val.(*pfs.BranchInfo).Branch.Repo)
	},
}

var branchesIndexes = []*col.Index{branchesRepoIndex}

// collectionsV0 returns a list of all the PFS collections for
// postgres-initialization purposes. These collections are not usable for
// querying.
// DO NOT MODIFY THIS FUNCTION
// IT HAS BEEN USED IN A RELEASED MIGRATION
func collectionsV0() []col.PostgresCollection {
	return []col.PostgresCollection{
		col.NewPostgresCollection("repos", nil, nil, nil, reposIndexes),
		col.NewPostgresCollection("commits", nil, nil, nil, commitsIndexes),
		col.NewPostgresCollection("branches", nil, nil, nil, branchesIndexes),
	}
}
