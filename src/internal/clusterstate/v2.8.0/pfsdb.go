package v2_8_0

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/proto"

	v2_7_0 "github.com/pachyderm/pachyderm/v2/src/internal/clusterstate/v2.7.0"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
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

func ListBranchesEdgesTriggersFromCollections(ctx context.Context, q sqlx.QueryerContext) ([]*Branch, []*Edge, []*BranchTrigger, error) {
	type branchColRow struct {
		v2_7_0.CollectionRecord
		RepoID uint64 `db:"repo_id"`
	}
	var branchColRows []branchColRow
	if err := sqlx.SelectContext(ctx, q, &branchColRows, `
		SELECT col.key, col.proto, col.createdat, col.updatedat, repo.id as repo_id
		FROM pfs.repos repo
			JOIN core.projects project ON repo.project_id = project.id
			JOIN collections.branches col ON col.idx_repo = project.name || '/' || repo.name || '.' || repo.type
		ORDER BY createdat, key ASC;
	`); err != nil {
		return nil, nil, nil, errors.Wrap(err, "listing branches from collections.branches")
	}

	// Build a map from branch key to branch
	keyToBranch := make(map[string]*Branch)
	// Build a map from branch key to its direct provenance
	keyToDirectProv := make(map[string][]string)
	keyToTrigger := make(map[string]*pfs.Trigger)
	var branches []*Branch
	for i, row := range branchColRows {
		var branchInfo pfs.BranchInfo
		if err := proto.Unmarshal(row.Proto, &branchInfo); err != nil {
			return nil, nil, nil, errors.Wrap(err, "unmarshaling branch")
		}
		branch := Branch{
			ID:        uint64(i + 1),
			Name:      branchInfo.Branch.Name,
			RepoID:    row.RepoID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		}
		keyToBranch[row.Key] = &branch
		if err := sqlx.GetContext(ctx, q, &branch.Head, `select int_id from pfs.commits where commit_id = $1`, branchInfo.Head.Key()); err != nil {
			return nil, nil, nil, errors.Wrap(err, "getting commit id")
		}
		for _, prov := range branchInfo.DirectProvenance {
			keyToDirectProv[row.Key] = append(keyToDirectProv[row.Key], prov.Key())
		}
		if branchInfo.Trigger != nil {
			// Note that we use branchInfo.Trigger instead of BranchTrigger here because we don't have the branch id yet.
			// also branchInfo.Trigger only has the branch name, and not the repo name.
			branchInfo.Trigger.Branch = branchInfo.Branch.Repo.Key() + "@" + branchInfo.Trigger.Branch
			keyToTrigger[row.Key] = branchInfo.Trigger
		}
		branches = append(branches, &branch)
	}

	var edges []*Edge
	for fromKey, toKeys := range keyToDirectProv {
		for _, toKey := range toKeys {
			edges = append(edges, &Edge{FromID: keyToBranch[fromKey].ID, ToID: keyToBranch[toKey].ID})
		}
	}

	// Branch triggers
	var triggers []*BranchTrigger
	triggerIDs := make(map[BranchTrigger]uint64)
	for key, trigger := range keyToTrigger {
		if _, ok := keyToBranch[trigger.Branch]; !ok {
			return nil, nil, nil, errors.Errorf("branch not found: %s", trigger.Branch)
		}
		bt := BranchTrigger{
			BranchID:      keyToBranch[trigger.Branch].ID,
			CronSpec:      trigger.CronSpec,
			RateLimitSpec: trigger.RateLimitSpec,
			Size:          trigger.Size,
			NumCommits:    uint64(trigger.Commits),
			All:           trigger.All,
		}
		if _, ok := triggerIDs[bt]; !ok {
			triggers = append(triggers, &bt)
			triggerIDs[bt] = uint64(len(triggers))
		}
		keyToBranch[key].TriggerID = NullUint64{Uint64: triggerIDs[bt], Valid: true}
	}

	return branches, edges, triggers, nil
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

func migrateBranches(ctx context.Context, env migrations.Env) error {
	tx := env.Tx
	if _, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS pfs.branches (
			id bigserial PRIMARY KEY,
			name text NOT NULL,
			head bigint REFERENCES pfs.commits(int_id) NOT NULL,
			repo_id bigint REFERENCES pfs.repos(id) NOT NULL,
			created_at timestamptz,
			updated_at timestamptz,
			UNIQUE (repo_id, name)
		);
	`); err != nil {
		return errors.Wrap(err, "creating branches table")
	}
	if _, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS pfs.branch_provenance (
			from_id bigint REFERENCES pfs.branches(id) NOT NULL,
			to_id bigint REFERENCES pfs.branches(id) NOT NULL
		);
	`); err != nil {
		return errors.Wrap(err, "creating branch_provenance table")
	}
	if _, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS pfs.branch_triggers (
			id bigserial PRIMARY KEY,
			branch_id bigint REFERENCES pfs.branches(id) NOT NULL,
			cron_spec text,
			rate_limit_spec text,
			size text,
			num_commits bigint,
			all_conditions bool,
			UNIQUE (branch_id, cron_spec, rate_limit_spec, size, num_commits, all_conditions)
		);
		ALTER TABLE pfs.branches ADD COLUMN trigger_id bigint REFERENCES pfs.branch_triggers(id);
	`); err != nil {
		return errors.Wrap(err, "creating branch triggers table")
	}

	insertBranchStmt, err := tx.PrepareContext(ctx, `INSERT INTO pfs.branches(name, head, repo_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING id`)
	if err != nil {
		return errors.Wrap(err, "preparing insert statement")
	}
	insertBranchProvStmt, err := tx.PrepareContext(ctx, `INSERT INTO pfs.branch_provenance(from_id, to_id) VALUES ($1, $2)`)
	if err != nil {
		return errors.Wrap(err, "preparing insert branch provenance statement")
	}
	insertTriggerStmt, err := tx.PrepareContext(ctx, `INSERT INTO pfs.branch_triggers(branch_id, cron_spec, rate_limit_spec, size, num_commits, all_conditions) VALUES ($1, $2, $3, $4, $5, $6)`)
	if err != nil {
		return errors.Wrap(err, "preparing insert branch trigger statement")
	}

	branches, edges, triggers, err := ListBranchesEdgesTriggersFromCollections(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "listing branches from collections.branches")
	}

	var id uint64
	for _, branch := range branches {
		if err := insertBranchStmt.QueryRowContext(ctx, branch.Name, branch.Head, branch.RepoID, branch.CreatedAt, branch.UpdatedAt).Scan(&id); err != nil {
			return errors.Wrapf(err, "inserting branch: %s, in repo: %d, head: %d", branch.Name, branch.RepoID, branch.Head)
		}
		if id != branch.ID {
			return errors.Errorf("expected branch id %d, got %d", branch.ID, id)
		}
	}
	for _, edge := range edges {
		if _, err := insertBranchProvStmt.ExecContext(ctx, edge.FromID, edge.ToID); err != nil {
			return errors.Wrap(err, "inserting branch provenance")
		}
	}
	for _, trigger := range triggers {
		if _, err := insertTriggerStmt.ExecContext(ctx, trigger.BranchID, trigger.CronSpec, trigger.RateLimitSpec, trigger.Size, trigger.NumCommits, trigger.All); err != nil {
			return errors.Wrap(err, "inserting trigger")
		}
	}
	for _, branch := range branches {
		if branch.TriggerID.Valid {
			if _, err := tx.ExecContext(ctx, `UPDATE pfs.branches SET trigger_id = $1 WHERE id = $2`, branch.TriggerID.Uint64, branch.ID); err != nil {
				return errors.Wrap(err, "updating branch trigger")
			}
		}
	}

	// Create indices at the end to speed up the inserts
	if _, err := tx.ExecContext(ctx, `CREATE INDEX repo_name_idx ON pfs.branches (repo_id, name);`); err != nil {
		return errors.Wrap(err, "creating index on pfs.branches")
	}

	// Create notify triggers for watchers
	if _, err := tx.ExecContext(ctx, `
		CREATE OR REPLACE FUNCTION pfs.notify_branches() RETURNS TRIGGER AS $$
		DECLARE
			row record;
			payload text;
		BEGIN
			IF TG_OP = 'DELETE' THEN
				row := OLD;
			ELSE
				row := NEW;
			END IF;
			payload := TG_OP || ' ' || row.int_id::text;
			PERFORM pg_notify('pfs_branches', payload);
			PERFORM pg_notify('pfs_branches_repo_' || row.repo_id::text, payload);
			return row;
		END;
		$$ LANGUAGE plpgsql;
	`); err != nil {
		return errors.Wrap(err, "creating notify trigger on pfs.branches")
	}
	// Add this at the end to avoid accidentally updating updated_at field.
	if _, err := tx.ExecContext(ctx, `
		CREATE TRIGGER set_updated_at
			BEFORE UPDATE ON pfs.branches
			FOR EACH ROW EXECUTE PROCEDURE core.set_updated_at_to_now();
	`); err != nil {
		return errors.Wrap(err, "creating set_updated_at trigger")
	}

	return nil
}
