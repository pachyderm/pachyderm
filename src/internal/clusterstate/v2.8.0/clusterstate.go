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
		}).
		Apply("Migrate collections.repos to pfs.repos", func(ctx context.Context, env migrations.Env) error {
			if err := migrateRepos(ctx, env.Tx); err != nil {
				return errors.Wrap(err, "migrating repos")
			}
			return nil
		})
}
