package v2_8_0

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/proto"

	v2_7_0 "github.com/pachyderm/pachyderm/v2/src/internal/clusterstate/v2.7.0"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

func generateTriggerFunctionStatement(schema, table, channel string) string {
	template := `
	CREATE OR REPLACE FUNCTION %s.notify_%s() RETURNS TRIGGER AS $$
	DECLARE
		row record;
		payload text;
	BEGIN
		IF TG_OP = 'DELETE' THEN
			row := OLD;
		ELSE
			row := NEW;
		END IF;
		payload := TG_OP || ' ' || row.id::text;
		PERFORM pg_notify('%s', payload);
		return row;
	END;
	$$ LANGUAGE plpgsql;

	CREATE TRIGGER notify
		AFTER INSERT OR UPDATE OR DELETE ON %s.%s
		FOR EACH ROW EXECUTE PROCEDURE %s.notify_%s();
	`
	return fmt.Sprintf(template, schema, table, channel, schema, table, schema, table)
}

func ListReposFromCollection(ctx context.Context, q sqlx.QueryerContext) ([]Repo, error) {
	// First collect all repos from collections.repos
	var repoColRows []v2_7_0.CollectionRecord
	if err := sqlx.SelectContext(ctx, q, &repoColRows, `SELECT key, proto, createdat, updatedat FROM collections.repos ORDER BY createdat, key ASC`); err != nil {
		return nil, errors.Wrap(err, "listing repos from collections.repos")
	}

	var repos []Repo
	for i, row := range repoColRows {
		var repoInfo pfs.RepoInfo
		if err := proto.Unmarshal(row.Proto, &repoInfo); err != nil {
			return nil, errors.Wrap(err, "unmarshaling repo")
		}
		repos = append(repos, Repo{
			ID:          uint64(i + 1),
			Name:        repoInfo.Repo.Name,
			ProjectName: repoInfo.Repo.Project.Name,
			Description: repoInfo.Description,
			RepoType:    repoInfo.Repo.Type,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		})
	}
	return repos, nil
}

func createPFSSchema(ctx context.Context, tx *pachsql.Tx) error {
	// pfs schema already exists, but this SQL is idempotent
	if _, err := tx.ExecContext(ctx, `CREATE SCHEMA IF NOT EXISTS pfs;`); err != nil {
		return errors.Wrap(err, "creating core schema")
	}

	return nil
}

func createReposTable(ctx context.Context, tx *pachsql.Tx) error {
	if _, err := tx.ExecContext(ctx, `
		DROP TYPE IF EXISTS pfs.repo_type;
		CREATE TYPE pfs.repo_type AS ENUM ('unknown', 'user', 'meta', 'spec');
	`); err != nil {
		return errors.Wrap(err, "creating repo_type enum")
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
			UNIQUE (name, project_id, type)
		);
		CREATE INDEX name_type_idx ON pfs.repos (name, type);
	`); err != nil {
		return errors.Wrap(err, "creating pfs.repos table")
	}
	if _, err := tx.ExecContext(ctx, `
		CREATE TRIGGER set_updated_at
			BEFORE UPDATE ON pfs.repos
			FOR EACH ROW EXECUTE PROCEDURE core.set_updated_at_to_now();
	`); err != nil {
		return errors.Wrap(err, "creating set_updated_at trigger")
	}
	// Create a trigger that notifies on changes to pfs.repos
	// This is used by the PPS API to watch for changes to repos
	if _, err := tx.ExecContext(ctx, generateTriggerFunctionStatement("pfs", "repos", "pfs_repos")); err != nil {
		return errors.Wrap(err, "creating notify trigger on pfs.repos")
	}
	return nil
}

// Migrate repos from collections.repos to pfs.repos
func migrateRepos(ctx context.Context, tx *pachsql.Tx) error {
	insertStmt, err := tx.PrepareContext(ctx, `INSERT INTO pfs.repos(name, type, project_id, description, created_at, updated_at) VALUES ($1, $2, (select id from core.projects where name=$3), $4, $5, $6)`)
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
		if _, err := insertStmt.ExecContext(ctx, repo.Name, repo.RepoType, repo.ProjectName, repo.Description, repo.CreatedAt, repo.UpdatedAt); err != nil {
			return errors.Wrap(err, "inserting repo")
		}
	}
	return nil
}

// alterCommitsTable1 adds useful new columns to pfs.commits table.
// Note that this is not the end all be all. We will need to make more changes after data has been migrated.
// TODO
//   - rename int_id to id? This will requires changing all references as well.
//   - make repo_id not null
//   - make origin not null
//   - make updated_at not null and default to current timestamp
//   - branch_id_str is a metadata reference to the branches table. Today this points to collections, but tomorrow it should
//     point to pfs.branches. Once the PFS master watches branches, we will no longer need this column.
func alterCommitsTable1(ctx context.Context, tx *pachsql.Tx) error {
	query := `
	CREATE TYPE pfs.commit_origin AS ENUM ('UNKNOWN', 'USER', 'AUTO', 'FSCK');

	ALTER TABLE IF EXISTS pfs.commits
	    DROP CONSTRAINT fk_col_commit,
		ADD COLUMN IF NOT EXISTS repo_id bigint REFERENCES pfs.repos(id),
		ADD COLUMN IF NOT EXISTS origin pfs.commit_origin,
		ADD COLUMN IF NOT EXISTS description text DEFAULT '',
		ADD COLUMN IF NOT EXISTS start_time timestamptz,
		ADD COLUMN IF NOT EXISTS finishing_time timestamptz,
		ADD COLUMN IF NOT EXISTS finished_time timestamptz,
		ADD COLUMN IF NOT EXISTS compacting_time bigint,
		ADD COLUMN IF NOT EXISTS validating_time bigint,
		ADD COLUMN IF NOT EXISTS error text,
		ADD COLUMN IF NOT EXISTS size bigint,
		ADD COLUMN IF NOT EXISTS updated_at timestamptz,
		ADD COLUMN IF NOT EXISTS branch_id_str text;

	CREATE TRIGGER set_updated_at
		BEFORE UPDATE ON pfs.commits
		FOR EACH ROW EXECUTE PROCEDURE core.set_updated_at_to_now();
	`
	if _, err := tx.ExecContext(ctx, query); err != nil {
		return errors.Wrap(err, "altering commits table")
	}
	if _, err := tx.ExecContext(ctx, `
	`); err != nil {
		return errors.Wrap(err, "creating set_updated_at trigger")
	}
	return nil
}

// commitAncestry is how we model the commit graph within a single repo.
func createCommitAncestryTable(ctx context.Context, tx *pachsql.Tx) error {
	query := `
	CREATE TABLE IF NOT EXISTS pfs.commit_ancestry (
		from_id bigint REFERENCES pfs.commits(int_id),
		to_id bigint REFERENCES pfs.commits(int_id),
		PRIMARY KEY (from_id, to_id)
	);

	CREATE INDEX ON pfs.commit_ancestry (from_id);
	CREATE INDEX ON pfs.commit_ancestry (to_id);
	`
	if _, err := tx.ExecContext(ctx, query); err != nil {
		return errors.Wrap(err, "creating commit_ancestry table")
	}
	return nil
}

// Migrate commits from collections.commits to pfs.commits
func migrateCommits(ctx context.Context, tx *pachsql.Tx) error {
	if err := alterCommitsTable1(ctx, tx); err != nil {
		return err
	}
	if err := createCommitAncestryTable(ctx, tx); err != nil {
		return err
	}
	return nil
}
