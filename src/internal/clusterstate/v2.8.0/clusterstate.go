package v2_8_0

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
)

func Migrate(state migrations.State) migrations.State {
	return state.
		Apply("Create pfs schema", func(ctx context.Context, env migrations.Env) error {
			if err := createPFSSchema(ctx, env.Tx); err != nil {
				return errors.Wrap(err, "creating pfs schema")
			}
			return nil
		}).
		Apply("Create pfs.repos table", func(ctx context.Context, env migrations.Env) error {
			if err := createReposTable(ctx, env.Tx); err != nil {
				return errors.Wrap(err, "creating pfs.repos table")
			}
			return nil
		}, migrations.Squash).
		Apply("Migrate collections.repos to pfs.repos", func(ctx context.Context, env migrations.Env) error {
			if err := migrateRepos(ctx, env.Tx); err != nil {
				return errors.Wrap(err, "migrating repos")
			}
			return nil
		}, migrations.Squash).
		Apply("Migrate collections.branches to pfs.branches", migrateBranches, migrations.Squash).
		Apply("Synthesize cluster defaults from environment variables", func(ctx context.Context, env migrations.Env) error {
			if err := synthesizeClusterDefaults(ctx, env); err != nil {
				return errors.Wrap(err, "could not synthesize cluster defaults")
			}
			return nil
		}).
		Apply("Synthesize user and effective specs from their pipeline details", synthesizeSpecs, migrations.Squash).
		Apply("Update pfs.commits table and create pfs.commit_ancestry", func(ctx context.Context, env migrations.Env) error {
			if err := updateCommitsSchema(ctx, env.Tx); err != nil {
				return errors.Wrap(err, "migrating commits")
			}
			return nil
		}).
		Apply("Migrate collections.commits to pfs.commits", func(ctx context.Context, env migrations.Env) error {
			if err := migrateCommits(ctx, env.Tx); err != nil {
				return errors.Wrap(err, "migrating commits")
			}
			return nil
		}, migrations.Squash)
}
