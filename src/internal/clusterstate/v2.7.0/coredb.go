package v2_7_0

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

func createCoreSchema(ctx context.Context, tx *pachsql.Tx) error {
	if _, err := tx.ExecContext(ctx, `CREATE SCHEMA IF NOT EXISTS core;`); err != nil {
		return errors.Wrap(err, "error creating core schema")
	}
	if _, err := tx.ExecContext(ctx, `
		CREATE OR REPLACE FUNCTION core.set_updated_at_to_now() RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = now();
			RETURN NEW;
		END;
		$$ language 'plpgsql';
	`); err != nil {
		return errors.Wrap(err, "error creating set_updated_at_to_now trigger function")
	}

	if err := createProjectsTable(ctx, tx); err != nil {
		return errors.Wrap(err, "error creating projects table")
	}
	return nil
}

func createProjectsTable(ctx context.Context, tx *pachsql.Tx) error {
	if _, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS core.projects (
			id bigserial PRIMARY KEY,
			name text UNIQUE NOT NULL,
			description text,
			created_at timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
			updated_at timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL
		);
	`); err != nil {
		return errors.Wrap(err, "error creating projects table")
	}
	if _, err := tx.ExecContext(ctx, `
		CREATE TRIGGER set_updated_at
			BEFORE UPDATE ON core.projects
			FOR EACH ROW EXECUTE PROCEDURE core.set_updated_at_to_now();
	`); err != nil {
		return errors.Wrap(err, "error creating set_updated_at trigger")
	}
	return nil
}

func migrateProjectsTable(ctx context.Context, tx *pachsql.Tx) error {
	selectTimestampsStmt, err := tx.Preparex("SELECT createdat, updatedat FROM collections.projects WHERE key = $1")
	if err != nil {
		return errors.Wrap(err, "error preparing select timestamps statement")
	}
	insertStmt, err := tx.Preparex("INSERT INTO core.projects(name, description, created_at, updated_at) VALUES($1, $2, $3, $4)")
	if err != nil {
		return errors.Wrap(err, "error preparing insert projects statement")
	}
	defer insertStmt.Close()

	projects := pfsdb.Projects(nil, nil).ReadWrite(tx)
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
}
