package v2_7_0

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

func Migrate(state migrations.State) migrations.State {
	return state.
		Apply("Setup core.* tables", func(ctx context.Context, env migrations.Env) error {
			if err := setupAll(ctx, env.Tx); err != nil {
				return errors.Wrap(err, "error setting up core.* tables")
			}
			return nil
		}).
		Apply("Migrate data from collections.projects to core.projects", func(ctx context.Context, env migrations.Env) error {
			selectTimestampsStmt, err := env.Tx.Preparex("SELECT createdat, updatedat FROM collections.projects WHERE key = $1")
			if err != nil {
				return errors.Wrap(err, "error preparing select timestamps statement")
			}
			insertStmt, err := env.Tx.Preparex("INSERT INTO core.projects(name, description, created_at, updated_at) VALUES($1, $2, $3, $4)")
			if err != nil {
				return errors.Wrap(err, "error preparing insert projects statement")
			}
			defer insertStmt.Close()

			projects := pfsdb.Projects(nil, nil).ReadWrite(env.Tx)
			projectInfo := &pfs.ProjectInfo{}
			if err := projects.List(projectInfo, &collection.Options{Target: collection.SortByCreateRevision, Order: collection.SortAscend}, func(string) error {
				var createdAt, updatedAt time.Time
				if err := selectTimestampsStmt.QueryRowContext(ctx, projectInfo.Project.Name).Scan(&createdAt, &updatedAt); err != nil {
					return errors.Wrap(err, "error scanning timestamps")
				}
				if _, err := insertStmt.ExecContext(ctx, projectInfo.Project.Name, projectInfo.Description, createdAt, updatedAt); err != nil {
					return errors.Wrap(err, "error inserting project")
				}
				return nil
			}); err != nil {
				return errors.Wrap(err, "error listing projects")
			}
			return nil
		})
}
