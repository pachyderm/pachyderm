package v2_7_0

import (
	"context"
	"time"

	proto "github.com/gogo/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

// CollectionRecord is a record in a collections table.
type CollectionRecord struct {
	Key       string    `db:"key"`
	Proto     []byte    `db:"proto"`
	CreatedAt time.Time `db:"createdat"`
	UpdatedAt time.Time `db:"updatedat"`
}

type Project struct {
	ID          uint64    `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func createCoreSchema(ctx context.Context, tx *pachsql.Tx) error {
	if _, err := tx.ExecContext(ctx, `CREATE SCHEMA IF NOT EXISTS core;`); err != nil {
		return errors.Wrap(err, "creating core schema")
	}
	if _, err := tx.ExecContext(ctx, `
		CREATE OR REPLACE FUNCTION core.set_updated_at_to_now() RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = now();
			RETURN NEW;
		END;
		$$ language 'plpgsql';
	`); err != nil {
		return errors.Wrap(err, "creating set_updated_at_to_now trigger function")
	}

	return nil
}

func createProjectsTable(ctx context.Context, tx *pachsql.Tx) error {
	if _, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS core.projects (
			id bigserial PRIMARY KEY,
			name varchar(51) UNIQUE NOT NULL,
			description text NOT NULL,
			created_at timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
			updated_at timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL
		);
	`); err != nil {
		return errors.Wrap(err, "creating projects table")
	}
	if _, err := tx.ExecContext(ctx, `
		CREATE TRIGGER set_updated_at
			BEFORE UPDATE ON core.projects
			FOR EACH ROW EXECUTE PROCEDURE core.set_updated_at_to_now();
	`); err != nil {
		return errors.Wrap(err, "creating set_updated_at trigger")
	}
	return nil
}

func migrateProjects(ctx context.Context, tx *pachsql.Tx) error {
	insertStmt, err := tx.PreparexContext(ctx, "INSERT INTO core.projects(name, description, created_at, updated_at) VALUES($1, $2, $3, $4)")
	if err != nil {
		return errors.Wrap(err, "preparing insert projects statement")
	}
	defer insertStmt.Close()

	var projectColRows []CollectionRecord
	if err := tx.SelectContext(ctx, &projectColRows, "SELECT key, proto, createdat, updatedat FROM collections.projects ORDER BY createdat ASC"); err != nil {
		return errors.Wrap(err, "listing from collections.projects")
	}
	var projectInfos []pfs.ProjectInfo
	for _, project := range projectColRows {
		projectInfo := pfs.ProjectInfo{}
		if err := proto.Unmarshal(project.Proto, &projectInfo); err != nil {
			return errors.Wrap(err, "unmarshalling project")
		}
		projectInfos = append(projectInfos, projectInfo)
	}
	// Note that although it is more efficient to batch insert multiple rows in a single statement,
	// we don't need it here because this is a one-time migration, and we don't expect users to have a large number of projects.
	for i := range projectInfos {
		if _, err := insertStmt.ExecContext(ctx, projectInfos[i].Project.Name, projectInfos[i].Description, projectColRows[i].CreatedAt, projectColRows[i].UpdatedAt); err != nil {
			return errors.Wrap(err, "inserting project")
		}
	}
	return nil
}
