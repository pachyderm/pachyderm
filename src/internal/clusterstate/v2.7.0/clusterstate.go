package v2_7_0

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
)

func Migrate(state migrations.State) migrations.State {
	return state.
		Apply("create cluster defaults", func(ctx context.Context, env migrations.Env) error {
			return setupPostgresCollections(ctx, env.Tx, ppsCollections()...)
		}).
		Apply("Create core schema", func(ctx context.Context, env migrations.Env) error {
			if err := createCoreSchema(ctx, env.Tx); err != nil {
				return errors.Wrap(err, "creating core schema")
			}
			return nil
		}, migrations.Squash).
		Apply("Create core.projects table", func(ctx context.Context, env migrations.Env) error {
			if err := createProjectsTable(ctx, env.Tx); err != nil {
				return errors.Wrap(err, "creating core.projects table")
			}
			return nil
		}, migrations.Squash).
		Apply("Migrate collections.projects to core.projects", func(ctx context.Context, env migrations.Env) error {
			if err := env.LockTables(ctx, "collections.projects"); err != nil {
				return errors.Wrap(err, "acquiring exclusive lock on collections.projects table")
			}
			if err := migrateProjects(ctx, env.Tx); err != nil {
				return errors.Wrap(err, "migrating collections.projects to core.projects")
			}
			return nil
		}, migrations.Squash)
}
