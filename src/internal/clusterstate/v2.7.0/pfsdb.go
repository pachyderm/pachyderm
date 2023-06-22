package v2_7_0

import (
	"context"

	proto "github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

func createPFSSchema(ctx context.Context, tx *pachsql.Tx) error {
	// pfs schema already exists, but this SQL is idempotent
	if _, err := tx.ExecContext(ctx, `CREATE SCHEMA IF NOT EXISTS pfs;`); err != nil {
		return errors.Wrap(err, "error creating core schema")
	}

	if err := createReposTable(ctx, tx); err != nil {
		return errors.Wrap(err, "error creating pfs.repos table")
	}
	return nil
}

func createReposTable(ctx context.Context, tx *pachsql.Tx) error {
	if _, err := tx.ExecContext(ctx, `
		DROP TYPE IF EXISTS pfs.repo_type;
		CREATE TYPE pfs.repo_type AS ENUM ('unknown', 'user', 'meta', 'spec');
	`); err != nil {
		return errors.Wrap(err, "error creating repo_type enum")
	}
	if _, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS pfs.repos (
			id bigserial PRIMARY KEY,
			name text NOT NULL,
			type pfs.repo_type NOT NULL,
			project_id bigint NOT NULL REFERENCES core.projects(id),
			created_at timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
			updated_at timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
			UNIQUE (name, project_id)
		);
	`); err != nil {
		return errors.Wrap(err, "error creating pfs.repos table")
	}
	if _, err := tx.ExecContext(ctx, `
		CREATE TRIGGER set_updated_at
			BEFORE UPDATE ON pfs.repos
			FOR EACH ROW EXECUTE PROCEDURE core.set_updated_at_to_now();
	`); err != nil {
		return errors.Wrap(err, "error creating set_updated_at trigger")
	}
	// Create a trigger that notifies on changes to pfs.repos
	// This is used by the PPS API to watch for changes to repos
	if _, err := tx.ExecContext(ctx, `
		CREATE OR REPLACE FUNCTION pfs.notify_repos() RETURNS TRIGGER AS $$
		DECLARE
			row record;
			base_channel text;
			payload text;
			key text;
		BEGIN
			IF TG_OP = 'DELETE' THEN
				row := OLD;
			ELSE
				row := NEW;
			END IF;

			SELECT project.name || '/' || row.name INTO key
			FROM core.projects project
			WHERE project.id = row.project_id;

			base_channel := 'pfs.repos';
			payload := TG_OP || ' ' || row.id::text || ' ' || key;

			PERFORM pg_notify(base_channel, payload);
			return row;
		END;
		$$ LANGUAGE plpgsql;

	`); err != nil {
		return errors.Wrap(err, "error creating notify trigger on pfs.repos")
	}
	if _, err := tx.ExecContext(ctx, `
		CREATE TRIGGER notify
			AFTER INSERT OR UPDATE OR DELETE ON pfs.repos
			FOR EACH ROW EXECUTE PROCEDURE pfs.notify_repos();
	`); err != nil {
		return errors.Wrap(err, "error creating notify trigger on pfs.repos")
	}
	return nil
}

func migrateRepos(ctx context.Context, tx *pachsql.Tx) error {
	insertStmt, err := tx.PrepareContext(ctx, `INSERT INTO pfs.repos(name, type, project_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`)
	if err != nil {
		return errors.Wrap(err, "preparing insert statement")
	}
	defer insertStmt.Close()

	// Migrate repos from collections.repos to pfs.repos
	// First collect all repos from collections.repos
	var repoColRows []CollectionRecord
	if err := tx.SelectContext(ctx, &repoColRows, `SELECT key, proto, createdat, updatedat FROM collections.repos ORDER BY createdat ASC`); err != nil {
		return errors.Wrap(err, "listing repos from collections.repos")
	}
	var repoInfos []pfs.RepoInfo
	for _, repo := range repoColRows {
		var repoInfo pfs.RepoInfo
		if err := proto.Unmarshal(repo.Proto, &repoInfo); err != nil {
			return errors.Wrap(err, "unmarshaling repo")
		}
		repoInfos = append(repoInfos, repoInfo)
	}
	// List all projects from core.projects
	var projects []Project
	if err := tx.SelectContext(ctx, &projects, `SELECT id, name FROM core.projects`); err != nil {
		return errors.Wrap(err, "listing projects from core.projects")
	}
	// Create a map of project name to project id
	projectsMap := make(map[string]uint64)
	for _, project := range projects {
		projectsMap[project.Name] = project.ID
	}

	for i := range repoInfos {
		if _, err := insertStmt.ExecContext(ctx, repoInfos[i].Name, repoInfos[i].Type, repoInfos[i].Project)
	}

	return nil
}
