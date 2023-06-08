package v2_7_0

import (
	"context"
	"time"

	proto "github.com/gogo/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
)

// CollectionRecord is a record in a collections table.
type CollectionRecord struct {
	Key       string    `db:"key"`
	Proto     []byte    `db:"proto"`
	CreatedAt time.Time `db:"createdat"`
	UpdatedAt time.Time `db:"updatedat"`
}

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
	insertStmt, err := tx.Preparex("INSERT INTO core.projects(name, description, created_at, updated_at) VALUES($1, $2, $3, $4)")
	if err != nil {
		return errors.Wrap(err, "error preparing insert projects statement")
	}
	defer insertStmt.Close()

	projectColRows := []CollectionRecord{}
	if err := tx.SelectContext(ctx, &projectColRows, "SELECT key, proto, createdat, updatedat FROM collections.projects ORDER BY createdat ASC"); err != nil {
		return errors.Wrap(err, "error listing from collections.projects")
	}
	projectInfos := []ProjectInfo{}
	for _, project := range projectColRows {
		projectInfo := ProjectInfo{}
		if err := proto.Unmarshal(project.Proto, &projectInfo); err != nil {
			return errors.Wrap(err, "error unmarshalling project")
		}
		projectInfos = append(projectInfos, projectInfo)
	}
	for i := 0; i < len(projectColRows); i++ {
		if _, err := insertStmt.ExecContext(ctx, projectInfos[i].Project.Name, projectInfos[i].Description, projectColRows[i].CreatedAt, projectColRows[i].UpdatedAt); err != nil {
			return errors.Wrap(err, "error inserting project")
		}
	}
	return nil
}
