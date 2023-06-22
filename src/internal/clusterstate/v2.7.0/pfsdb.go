package v2_7_0

import (
	"context"

	proto "github.com/gogo/protobuf/proto"
	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

func ListReposFromCollection(ctx context.Context, q sqlx.QueryerContext) ([]Repo, error) {
	// First collect all repos from collections.repos
	var repoColRows []CollectionRecord
	if err := sqlx.SelectContext(ctx, q, &repoColRows, `SELECT key, proto, createdat, updatedat FROM collections.repos ORDER BY createdat, key ASC`); err != nil {
		return nil, errors.Wrap(err, "listing repos from collections.repos")
	}
	// List all projects from core.projects
	var projects []Project
	if err := sqlx.SelectContext(ctx, q, &projects, `SELECT id, name FROM core.projects`); err != nil {
		return nil, errors.Wrap(err, "listing projects from core.projects")
	}
	// Create a map of project name to project id
	projectNameToID := make(map[string]uint64)
	for _, project := range projects {
		projectNameToID[project.Name] = project.ID
	}

	var repos []Repo
	for i, row := range repoColRows {
		var repoInfo pfs.RepoInfo
		if err := proto.Unmarshal(row.Proto, &repoInfo); err != nil {
			return nil, errors.Wrap(err, "unmarshaling repo")
		}
		projectID, ok := projectNameToID[repoInfo.Repo.GetProject().Name]
		if !ok {
			return nil, errors.Errorf("project not found for repo %s", repoInfo.Repo)
		}
		repos = append(repos, Repo{
			ID:          uint64(i + 1),
			Name:        repoInfo.Repo.Name,
			Description: repoInfo.Description,
			RepoType:    repoInfo.Repo.Type,
			ProjectID:   projectID,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		})
	}
	return repos, nil
}

func createPFSSchema(ctx context.Context, tx *pachsql.Tx) error {
	// pfs schema already exists, but this SQL is idempotent
	if _, err := tx.ExecContext(ctx, `CREATE SCHEMA IF NOT EXISTS pfs;`); err != nil {
		return errors.Wrap(err, "error creating core schema")
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
			description text NOT NULL DEFAULT '',
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

// Migrate repos from collections.repos to pfs.repos
func migrateRepos(ctx context.Context, tx *pachsql.Tx) error {
	insertStmt, err := tx.PrepareContext(ctx, `INSERT INTO pfs.repos(name, type, project_id, description, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`)
	if err != nil {
		return errors.Wrap(err, "preparing insert statement")
	}
	defer insertStmt.Close()

	repos, err := ListReposFromCollection(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "listing repos from collections.repos")
	}
	// Insert all repos into pfs.repos
	for _, repo := range repos {
		if _, err := insertStmt.ExecContext(ctx, repo.Name, repo.RepoType, repo.ProjectID, repo.Description, repo.CreatedAt, repo.UpdatedAt); err != nil {
			return errors.Wrap(err, "inserting repo")
		}
	}
	return nil
}
