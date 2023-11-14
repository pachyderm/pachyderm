package v2_8_0

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate/migrationutils"
	v2_7_0 "github.com/pachyderm/pachyderm/v2/src/internal/clusterstate/v2.7.0"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pbutil"
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

func ListCommitsFromCollection(ctx context.Context, q sqlx.QueryerContext) ([]CommitInfo, map[string]string, error) {
	var commitsColRows []v2_7_0.CollectionRecord
	childParent := make(map[string]string)
	if err := sqlx.SelectContext(ctx, q, &commitsColRows, `SELECT key, proto, createdat, updatedat FROM collections.commits ORDER BY createdat, key ASC`); err != nil {
		return nil, nil, errors.Wrap(err, "listing commits from collections.commits")
	}
	var commits []CommitInfo
	for _, row := range commitsColRows {
		commitInfo := &pfs.CommitInfo{}
		if err := proto.Unmarshal(row.Proto, commitInfo); err != nil {
			return nil, nil, errors.Wrap(err, "unmarshaling repo")
		}
		commit := InfoToCommit(commitInfo, 0, time.Time{}, time.Time{})
		commitAncestry := InfoToCommitAncestry(commitInfo)
		commits = append(commits, commit.CommitInfo)
		if commitAncestry.ParentCommit != "" {
			childParent[commitInfo.Commit.Key()] = commitAncestry.ParentCommit
		}
		for _, child := range commitAncestry.ChildCommits {
			childParent[child] = commitInfo.Commit.Key()
		}
	}
	return commits, childParent, nil
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

	// Map branch key to branch db object, which contains the branch id.
	keyToBranch := make(map[string]*Branch)
	// Map branch key to its direct provenance.
	keyToDirectProv := make(map[string][]string)
	// Map trigger to the source branch
	triggerToBranchID := make(map[*pfs.Trigger]uint64)

	var (
		branches []*Branch
		edges    []*Edge
		triggers []*BranchTrigger
	)
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
		// Not ideal to make a db call for each branch, but we need the commit id to populate the head field.
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
			triggerToBranchID[branchInfo.Trigger] = branch.ID
		}
		keyToBranch[row.Key] = &branch
		branches = append(branches, &branch)
	}
	for fromKey, prov := range keyToDirectProv {
		for _, toKey := range prov {
			edges = append(edges, &Edge{FromID: keyToBranch[fromKey].ID, ToID: keyToBranch[toKey].ID})
		}
	}
	for trigger, fromBranchID := range triggerToBranchID {
		if _, ok := keyToBranch[trigger.Branch]; !ok {
			// We explicitly ignore triggering branches that don't exist because pfs.branch_triggers enforce foreign key constraints.
			log.Info(ctx, "Skipping branch trigger because branch does not exist", zap.Object("trigger", trigger))
			continue
		}
		bt := BranchTrigger{
			FromBranchID:  fromBranchID,
			ToBranchID:    keyToBranch[trigger.Branch].ID,
			CronSpec:      trigger.CronSpec,
			RateLimitSpec: trigger.RateLimitSpec,
			Size:          trigger.Size,
			NumCommits:    uint64(trigger.Commits),
			All:           trigger.All,
		}
		triggers = append(triggers, &bt)
	}

	return branches, edges, triggers, nil
}

func createPFSSchema(ctx context.Context, env migrations.Env) error {
	tx := env.Tx
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
func migrateRepos(ctx context.Context, env migrations.Env) error {
	tx := env.Tx
	if err := createReposTable(ctx, tx); err != nil {
		return errors.Wrap(err, "creating pfs.repos table")
	}
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

// alterCommitsTable adds useful new columns to pfs.commits table.
// Note: that this is not the end all be all. We will need to make more changes after data has been migrated.
// Note: we don't cascade deleting on branch_id because we need to delete the commit_totals and commit_diffs.
// TODO
//   - rename int_id to id? This will requires changing all references as well.
//   - make repo_id not null
//   - make origin not null
func alterCommitsTable(ctx context.Context, tx *pachsql.Tx) error {
	query := `
	CREATE TYPE pfs.commit_origin AS ENUM ('ORIGIN_KIND_UNKNOWN', 'USER', 'AUTO', 'FSCK');

	ALTER TABLE pfs.commits
		ADD COLUMN repo_id bigint REFERENCES pfs.repos(id) ON DELETE CASCADE,
		ADD COLUMN origin pfs.commit_origin,
		ADD COLUMN description text NOT NULL DEFAULT '',
		ADD COLUMN start_time timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
		ADD COLUMN finishing_time timestamptz,
		ADD COLUMN finished_time timestamptz,
		ADD COLUMN compacting_time_s bigint,
		ADD COLUMN validating_time_s bigint,
		ADD COLUMN error text,
		ADD COLUMN size bigint,
		ADD COLUMN updated_at timestamptz DEFAULT CURRENT_TIMESTAMP,
		ADD COLUMN branch_id bigint;

	CREATE TRIGGER set_updated_at
		BEFORE UPDATE ON pfs.commits
		FOR EACH ROW EXECUTE PROCEDURE core.set_updated_at_to_now();
	`
	if _, err := tx.ExecContext(ctx, query); err != nil {
		return errors.Wrap(err, "altering commits table")
	}
	return nil
}

func alterCommitsTablePostDataMigration(ctx context.Context, env migrations.Env) error {
	query := `
	ALTER TABLE pfs.commits
	    DROP CONSTRAINT fk_col_commit;
	`
	if _, err := env.Tx.ExecContext(ctx, query); err != nil {
		return errors.Wrap(err, "altering commits table")
	}
	return nil
}

// commitAncestry is how we model the commit graph within a single repo.
func createCommitAncestryTable(ctx context.Context, tx *pachsql.Tx) error {
	query := `
	CREATE TABLE pfs.commit_ancestry (
		parent bigint REFERENCES pfs.commits(int_id) ON DELETE CASCADE,
		child bigint REFERENCES pfs.commits(int_id) ON DELETE CASCADE,
		PRIMARY KEY (parent, child)
	);
	CREATE INDEX ON pfs.commit_ancestry (child, parent);
	`
	if _, err := tx.ExecContext(ctx, query); err != nil {
		return errors.Wrap(err, "creating commit_ancestry table")
	}
	return nil
}

func createNotifyCommitsTrigger(ctx context.Context, tx *pachsql.Tx) error {
	query := `
	CREATE FUNCTION pfs.notify_commits() RETURNS TRIGGER AS $$
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
		PERFORM pg_notify('pfs_commits', payload);
		PERFORM pg_notify('pfs_commits_repo_' || row.repo_id::text, payload);
		PERFORM pg_notify('pfs_commits_' || row.int_id::text, payload);
		return row;
	END;
	$$ LANGUAGE plpgsql;
	CREATE TRIGGER notify
		AFTER INSERT OR UPDATE OR DELETE ON pfs.commits
		FOR EACH ROW EXECUTE PROCEDURE pfs.notify_commits();
	`
	if _, err := tx.ExecContext(ctx, query); err != nil {
		return errors.Wrap(err, "creating notify trigger on pfs.commits")
	}
	return nil

}

func migrateCommits(ctx context.Context, env migrations.Env) error {
	log.Info(ctx, "updating database schema for commit tables")
	if err := alterCommitsTable(ctx, env.Tx); err != nil {
		return err
	}
	if err := createCommitAncestryTable(ctx, env.Tx); err != nil {
		return err
	}
	log.Info(ctx, "migrating commits")
	if err := migrateCommitsFromCollections(ctx, env.Tx); err != nil {
		return err
	}
	log.Info(ctx, "finalizing database schema for commit tables")
	if err := alterCommitsTablePostDataMigration(ctx, env); err != nil {
		return err
	}
	if err := createNotifyCommitsTrigger(ctx, env.Tx); err != nil {
		return err
	}
	return nil
}

type commitCollection struct {
	v2_7_0.CollectionRecord
	IntID uint64 `db:"int_id"`
}

func migrateCommitsFromCollections(ctx context.Context, tx *pachsql.Tx) error {
	count := struct {
		Collections uint64 `db:"col_count"`
		Commits     uint64 `db:"commits_count"`
	}{}
	if err := tx.GetContext(ctx, &count, `SELECT count(commit.int_id) AS commits_count, count(col.key) 
    	AS col_count FROM pfs.commits commit LEFT JOIN collections.commits col on commit.commit_id = col.key;`); err != nil {
		return errors.Wrap(err, "counting rows in collections.commits")
	}
	if count.Collections != count.Commits {
		return errors.Errorf("collections.commits has %d rows while pfs.commits has %d rows", count.Collections, count.Commits)
	}
	if count.Collections == 0 {
		return nil
	}
	pageSize := uint64(1000)
	totalPages := count.Collections / pageSize
	if count.Collections%pageSize > 0 {
		totalPages++
	}
	batcher := migrationutils.NewSimplePostgresBatcher(tx)
	for i := uint64(0); i < totalPages; i++ {
		log.Info(ctx, "migrating commits", zap.Uint64("current", i*pageSize), zap.Uint64("total", count.Collections))
		var page []commitCollection
		if err := tx.SelectContext(ctx, &page, fmt.Sprintf(`
		SELECT commit.int_id, col.key, col.proto, col.updatedat, col.createdat
		FROM pfs.commits commit JOIN collections.commits AS col ON commit.commit_id = col.key
		ORDER BY commit.int_id ASC LIMIT %d OFFSET %d`, pageSize, i*pageSize)); err != nil {
			return errors.Wrap(err, "could not read table")
		}
		if err := migratePage(ctx, page, batcher); err != nil {
			return err
		}
	}
	if err := batcher.Flush(ctx); err != nil {
		return err
	}
	return nil
}

func migratePage(ctx context.Context, page []commitCollection, batcher *migrationutils.SimplePostgresBatcher) error {
	for _, col := range page {
		commit, ancestry, err := protoToCommit(col)
		if err != nil {
			return err
		}
		if !commit.StartTime.Valid {
			return errors.Errorf("commit %s has a nil start time", commit.IntID)
		}
		if err := migrateCommit(ctx, commit, batcher); err != nil {
			return errors.Wrap(err, fmt.Sprintf("migrating commit: commit int id %v", commit.IntID))
		}
		if ancestry.ParentCommit == "" && ancestry.ChildCommits == nil {
			continue
		}
		if err := migrateRelatives(ctx, commit.IntID, ancestry, batcher); err != nil {
			return errors.Wrap(err, fmt.Sprintf("migrating commit relatives: commit int id: %v, ancestry %+v",
				commit.IntID, ancestry))
		}
	}
	return nil
}

func migrateCommit(ctx context.Context, commit *Commit, batcher *migrationutils.SimplePostgresBatcher) error {
	query := `WITH repo_row_id AS (SELECT id from pfs.repos WHERE name=$1 AND type=$2 
                                                AND project_id=(SELECT id from core.projects WHERE name= $3))
		UPDATE pfs.commits SET 
			commit_id=$4, 
			commit_set_id=$5,
		    repo_id=(SELECT id from repo_row_id), 
		    branch_id=(SELECT id from pfs.branches WHERE name=$6 AND repo_id=(SELECT id from repo_row_id)), 
			description=$7, 
			origin=$8, 
			start_time=$9,`

	if commit.FinishedTime.Valid {
		query += fmt.Sprintf("\nfinished_time='%v',", commit.FinishedTime.Time.Format(time.RFC3339))
	}
	if commit.FinishedTime.Valid {
		query += fmt.Sprintf("\nfinishing_time='%v',", commit.FinishingTime.Time.Format(time.RFC3339))
	}
	if commit.CompactingTime.Valid {
		query += fmt.Sprintf("\ncompacting_time_s=%v,", commit.CompactingTime.Int64)
	}
	if commit.ValidatingTime.Valid {
		query += fmt.Sprintf("\nvalidating_time_s=%v,", commit.ValidatingTime.Int64)
	}

	query += `
			size=$10, 
			error=$11 
		WHERE int_id=$12;`
	if !commit.StartTime.Valid {
		return errors.Errorf("commit %s has a nil start time", commit.IntID)
	}
	return batcher.Add(ctx, query, commit.RepoName, commit.RepoType, commit.ProjectName, commit.CommitID,
		commit.CommitSetID, commit.BranchName.String, commit.Description, commit.Origin, commit.StartTime.Time,
		commit.Size, commit.Error, commit.IntID)
}

func protoToCommit(col commitCollection) (*Commit, *CommitAncestry, error) {
	commitInfo := &pfs.CommitInfo{}
	if err := proto.Unmarshal(col.Proto, commitInfo); err != nil {
		return nil, nil, errors.Wrapf(err, "could not unmarshal proto")
	}
	return InfoToCommit(commitInfo, col.IntID, col.CreatedAt, col.UpdatedAt),
		InfoToCommitAncestry(commitInfo), nil
}
func InfoToCommitAncestry(commitInfo *pfs.CommitInfo) *CommitAncestry {
	var children []string
	for _, child := range commitInfo.ChildCommits {
		children = append(children, child.Key())
	}
	parentKey := ""
	if commitInfo.ParentCommit != nil {
		parentKey = commitInfo.ParentCommit.Key()
	}
	return &CommitAncestry{
		ChildCommits: children,
		ParentCommit: parentKey,
	}
}

func InfoToCommit(commitInfo *pfs.CommitInfo, id uint64, createdAt, updatedAt time.Time) *Commit {
	if commitInfo.Details == nil {
		commitInfo.Details = &pfs.CommitInfo_Details{}
	}
	branchName := sql.NullString{String: "", Valid: false}
	if commitInfo.Commit.Branch != nil && commitInfo.Commit.Branch.Name != "" {
		branchName = sql.NullString{String: commitInfo.Commit.Branch.Name, Valid: true}
	}
	return &Commit{
		IntID: id,
		CommitInfo: CommitInfo{
			CommitID:       commitInfo.Commit.Key(),
			CommitSetID:    commitInfo.Commit.Id,
			RepoName:       commitInfo.Commit.Repo.Name,
			RepoType:       commitInfo.Commit.Repo.Type,
			ProjectName:    commitInfo.Commit.Repo.Project.Name,
			BranchName:     branchName,
			Origin:         commitInfo.Origin.Kind.String(),
			StartTime:      pbutil.SanitizeTimestampPb(commitInfo.Started),
			FinishingTime:  pbutil.SanitizeTimestampPb(commitInfo.Finishing),
			FinishedTime:   pbutil.SanitizeTimestampPb(commitInfo.Finished),
			Description:    commitInfo.Description,
			CompactingTime: pbutil.DurationPbToBigInt(commitInfo.Details.CompactingTime),
			ValidatingTime: pbutil.DurationPbToBigInt(commitInfo.Details.ValidatingTime),
			Size:           commitInfo.Details.SizeBytes,
			Error:          commitInfo.Error,
		},
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

func migrateRelatives(ctx context.Context, id uint64, commitAncestry *CommitAncestry, batcher *migrationutils.SimplePostgresBatcher) error {
	ancestryQueryTemplate := `
		INSERT INTO pfs.commit_ancestry
		(parent, child)
		VALUES %s
		ON CONFLICT DO NOTHING;
		`
	valuesTemplate := `($1, (SELECT int_id FROM pfs.commits WHERE commit_id=$%d))`
	params := []any{id}
	queryVarNum := 2
	values := make([]string, 0)
	for _, child := range commitAncestry.ChildCommits {
		values = append(values, fmt.Sprintf(valuesTemplate, queryVarNum))
		params = append(params, child)
		queryVarNum++
	}
	if commitAncestry.ParentCommit != "" {
		values = append(values, fmt.Sprintf(`((SELECT int_id FROM pfs.commits WHERE commit_id=$%d), $1)`, queryVarNum))
		params = append(params, commitAncestry.ParentCommit)
	}
	query := fmt.Sprintf(ancestryQueryTemplate, strings.Join(values, ","))
	return batcher.Add(ctx, query, params...)
}

func migrateBranches(ctx context.Context, env migrations.Env) error {
	tx := env.Tx
	if _, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS pfs.branches (
			id bigserial PRIMARY KEY,
			name text NOT NULL,
			head bigint REFERENCES pfs.commits(int_id) NOT NULL,
			repo_id bigint REFERENCES pfs.repos(id) ON DELETE CASCADE NOT NULL,
			created_at timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
			updated_at timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
			UNIQUE (repo_id, name)
		);
	`); err != nil {
		return errors.Wrap(err, "creating branches table")
	}
	if _, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS pfs.branch_provenance (
			from_id bigint REFERENCES pfs.branches(id) NOT NULL,
			to_id bigint REFERENCES pfs.branches(id) NOT NULL,
			PRIMARY KEY (from_id, to_id),
			CHECK (from_id <> to_id)
		);
	`); err != nil {
		return errors.Wrap(err, "creating branch_provenance table")
	}
	if _, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS pfs.branch_triggers (
			from_branch_id bigint REFERENCES pfs.branches(id) PRIMARY KEY,
			to_branch_id bigint REFERENCES pfs.branches(id) NOT NULL,
			cron_spec text,
			rate_limit_spec text,
			size text,
			num_commits bigint,
			all_conditions bool
		);
		CREATE INDEX ON pfs.branch_triggers (to_branch_id);
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
	insertTriggerStmt, err := tx.PrepareContext(ctx, `INSERT INTO pfs.branch_triggers(from_branch_id, to_branch_id, cron_spec, rate_limit_spec, size, num_commits, all_conditions) VALUES ($1, $2, $3, $4, $5, $6, $7)`)
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
		if _, err := insertTriggerStmt.ExecContext(ctx, trigger.FromBranchID, trigger.ToBranchID, trigger.CronSpec, trigger.RateLimitSpec, trigger.Size, trigger.NumCommits, trigger.All); err != nil {
			return errors.Wrap(err, "inserting trigger")
		}
	}

	// Create indices at the end to speed up the inserts
	if _, err := tx.ExecContext(ctx, `CREATE INDEX ON pfs.branch_provenance (to_id);`); err != nil {
		return errors.Wrap(err, "creating index on pfs.branch_provenance")
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
			payload := TG_OP || ' ' || row.id::text;
			PERFORM pg_notify('pfs_branches', payload);
			PERFORM pg_notify('pfs_branches_repo_' || row.repo_id::text, payload);
			return row;
		END;
		$$ LANGUAGE plpgsql;
		
		CREATE TRIGGER notify
			AFTER INSERT OR UPDATE OR DELETE ON pfs.branches
			FOR EACH ROW EXECUTE PROCEDURE pfs.notify_branches();
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
