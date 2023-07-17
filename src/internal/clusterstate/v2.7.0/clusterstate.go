package v2_7_0

import (
	"context"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsdb"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
)

func Migrate(state migrations.State) migrations.State {
	return state.
		Apply("create cluster defaults", func(ctx context.Context, env migrations.Env) error {
			var cc []col.PostgresCollection
			cc = append(cc, ppsdb.CollectionsV2_7_0()...)
			return col.SetupPostgresCollections(ctx, env.Tx, cc...)
		}).
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
			if err := env.LockTables(ctx, "collections.projects"); err != nil {
				return errors.Wrap(err, "acquiring exclusive lock on collections.projects table")
			}
			if err := migrateProjects(ctx, env.Tx); err != nil {
				return errors.Wrap(err, "migrating collections.projects to core.projects")
			}
			return nil
		}).
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
