package v2_5_0

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/pfs"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
)

func Migrate(state migrations.State) migrations.State {
	return state.
		Apply("Add projects collection", func(ctx context.Context, env migrations.Env) error {
			return setupPostgresCollections(ctx, env.Tx, pfsCollections()...)
		}).
		Apply("Add default project", func(ctx context.Context, env migrations.Env) error {
			var defaultProject = &pfs.ProjectInfo{
				Project: &pfs.Project{
					Name: "default", // hardcoded so that pfs.DefaultProjectName may change in the future
				},
			}
			if err := pfsdb.Projects(nil, nil).ReadWrite(env.Tx).Create("default", defaultProject); err != nil {
				return errors.Wrap(err, "could not create default project")
			}
			return nil
		}).
		Apply("Rename default project to “default”", func(ctx context.Context, env migrations.Env) error {
			if err := migratePFSDB(ctx, env.Tx); err != nil {
				return err
			}
			if err := migratePPSDB(ctx, env.Tx); err != nil {
				return err
			}
			if err := migrateAuth(ctx, env.Tx); err != nil {
				return err
			}
			return nil
		})
	// DO NOT MODIFY THIS STATE
	// IT HAS ALREADY SHIPPED IN A RELEASE
}
