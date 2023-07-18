package v2_7_0

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"go.uber.org/zap"
)

// ListProjectsFromCollection iterates over all projects in the collections.projects table
// and returns a list of Project objects that satisfy the relational model.
func ListProjectsFromCollection(ctx context.Context, q sqlx.QueryerContext) ([]Project, error) {
	var colRows []CollectionRecord
	if err := sqlx.SelectContext(ctx, q, &colRows, "SELECT key, proto, createdat, updatedat FROM collections.projects ORDER BY createdat ASC"); err != nil {
		return nil, errors.Wrap(err, "listing from collections.projects")
	}
	var projects []Project
	for i, row := range colRows {
		var projectInfo pfs.ProjectInfo
		if err := proto.Unmarshal(row.Proto, &projectInfo); err != nil {
			return nil, errors.Wrap(err, "unmarshalling project")
		}
		createdAt := row.CreatedAt
		if projectInfo.CreatedAt != nil {
			if projectInfo.CreatedAt.AsTime().Sub(createdAt) > time.Second {
				return nil, errors.Errorf("project %s's proto created at %s differs from database created at %s", projectInfo.Project.Name, projectInfo.CreatedAt.AsTime(), createdAt)
			}
			createdAt = projectInfo.CreatedAt.AsTime().UTC()
		}
		projects = append(projects, Project{ID: uint64(i + 1), Name: projectInfo.Project.Name, Description: projectInfo.Description, CreatedAt: createdAt, UpdatedAt: row.UpdatedAt})
	}
	return projects, nil
}

func createSetUpdatedAtTrigger(ctx context.Context, tx *pachsql.Tx, tableName string) error {
	_, err := tx.ExecContext(ctx, fmt.Sprintf(`
		CREATE TRIGGER set_updated_at
			BEFORE UPDATE ON %s
			FOR EACH ROW EXECUTE PROCEDURE core.set_updated_at_to_now();
	`, tableName))
	return errors.Wrap(err, "creating set updated_at trigger")
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
	log.Info(ctx, "create core.projects table")
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
	if err := createSetUpdatedAtTrigger(ctx, tx, "core.projects"); err != nil {
		return errors.Wrap(err, "creating set_updated_at trigger for core.projects")
	}
	return nil
}

func migrateProjects(ctx context.Context, tx *pachsql.Tx) error {
	insertStmt, err := tx.PreparexContext(ctx, "INSERT INTO core.projects(name, description, created_at, updated_at) VALUES($1, $2, $3, $4)")
	if err != nil {
		return errors.Wrap(err, "preparing insert projects statement")
	}
	defer insertStmt.Close()
	projects, err := ListProjectsFromCollection(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "listing projects from collection")
	}
	log.Info(ctx, "migrating collections.projects to core.projects", zap.String("completed", fmt.Sprintf("0/%d", len(projects))))
	// Note that although it is more efficient to batch insert multiple rows in a single statement,
	// we don't need it here because this is a one-time migration, and we don't expect users to have a large number of projects.
	for i, project := range projects {
		if _, err := insertStmt.ExecContext(ctx, project.Name, project.Description, project.CreatedAt, project.UpdatedAt); err != nil {
			return errors.Wrap(err, "inserting project")
		}
		log.Info(ctx, "migrating collections.projects to core.projects", zap.String("completed", fmt.Sprintf("%d/%d", i+1, len(projects))))
	}
	return nil
}
