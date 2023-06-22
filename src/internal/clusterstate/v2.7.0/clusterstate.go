package v2_7_0

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
)

func Migrate(state migrations.State) migrations.State {
	return state.
		Apply("Create core schema", func(ctx context.Context, env migrations.Env) error {
			if err := createCoreSchema(ctx, env.Tx); err != nil {
				return errors.Wrap(err, "creating core schema")
			}
			return nil
		}).
		Apply("Create core.projects table", func(ctx context.Context, env migrations.Env) error {
			if err := createProjectsTable(ctx, env.Tx); err != nil {
				return errors.Wrap(err, "creating core.projects table")
			}
			return nil
		}).
		Apply("Migrate collections.projects to core.projects", func(ctx context.Context, env migrations.Env) error {
			if err := migrateProjects(ctx, env.Tx); err != nil {
				return errors.Wrap(err, "error migrating projects")
			}
			return nil
		}).
		Apply("Create pfs schema", func(ctx context.Context, env migrations.Env) error {
			if err := createPFSSchema(ctx, env.Tx); err != nil {
				return errors.Wrap(err, "creating pfs schema")
			}
			return nil
		})
}
