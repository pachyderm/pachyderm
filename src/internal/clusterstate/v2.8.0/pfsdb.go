package v2_8_0

import (
	"context"
	"fmt"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/proto"

	v2_7_0 "github.com/pachyderm/pachyderm/v2/src/internal/clusterstate/v2.7.0"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

type commitCollection struct {
	IntID     uint64    `db:"int_id"`
	OldKey    string    `db:"key"`
	Pb        []byte    `db:"proto"`
	CreatedAt time.Time `db:"createdat"`
	UpdatedAt time.Time `db:"updatedat"`
}

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

func migrateCommitsFromCollections(ctx context.Context, tx *pachsql.Tx) error {
	count := struct {
		Collections uint64 `db:"col_count"`
		Commits     uint64 `db:"commits_count"`
	}{}
	rowCountQuery := `SELECT count(commit.int_id) AS commits_count, count(col.key) 
    	AS col_count FROM pfs.commits commit LEFT JOIN collections.commits col on commit.commit_id = col.key;`
	if err := tx.QueryRowxContext(ctx, rowCountQuery).StructScan(&count); err != nil {
		return errors.Wrap(err, "counting rows in collections.commits")
	}
	if count.Collections != count.Commits {
		return errors.New(fmt.Sprintf("collections.commits has %d rows while pfs.commits has %d rows",
			count.Collections, count.Commits))
	}
	pageSize := uint64(10)
	totalPages := count.Collections / pageSize
	if pageSize%count.Collections > 0 {
		totalPages++
	}
	for i := uint64(0); i < totalPages; i++ {
		page := make([]commitCollection, 0)
		getCommitsQuery := `SELECT commit.int_id, col.key, col.proto, col.updatedat, col.createdat
		FROM pfs.commits commit JOIN collections.commits AS col ON commit.commit_id = col.key 
		ORDER BY commit.int_id ASC LIMIT %d OFFSET %d`
		if err := tx.SelectContext(ctx, &page, fmt.Sprintf(getCommitsQuery, pageSize, i*pageSize)); err != nil {
			return errors.Wrap(err, "could not read table")
		}
		if len(page) == 0 {
			return nil
		}
		if err := migratePage(ctx, tx, page); err != nil {
			return err
		}
	}
	return nil
}

func migratePage(ctx context.Context, tx *pachsql.Tx, page []commitCollection) error {
	migrateQuery := `UPDATE pfs.commits SET commit_id=:commit_id, commit_set_id=:commit_set_id,
		    repo_id=(SELECT id from pfs.repos WHERE name=:repo_name AND type=:repo_type AND project_id=(SELECT id from core.projects WHERE name=:proj_name)),
		    branch_id_str=:branch_id_str, description=:description, origin=:origin, start_time=:start_time, finishing_time=:finishing_time, finished_time=:finished_time,
		    compacting_time=:compacting_time, validating_time=:validating_time, size=:size, error=:error WHERE int_id=:int_id;`
	for _, col := range page {
		commit, err := protoToCommit(col)
		if err != nil {
			return err
		}
		_, err = tx.NamedExecContext(ctx, migrateQuery, commit)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("migrating commit %d", commit.IntID))
		}
		if commit.ParentCommit == nil && commit.ChildCommits == nil {
			continue
		}
		if err := migrateRelatives(ctx, tx, commit); err != nil {
			return errors.Wrap(err, fmt.Sprintf("migrating relatives of commit %d", commit.IntID))
		}
	}
	return nil
}

func protoToCommit(col commitCollection) (*Commit, error) {
	commitInfo := &pfs.CommitInfo{}
	if err := proto.Unmarshal(col.Pb, commitInfo); err != nil {
		return nil, errors.Wrapf(err, "could not unmarshal proto")
	}
	if commitInfo.Details == nil {
		commitInfo.Details = &pfs.CommitInfo_Details{}
	}
	if commitInfo.Commit.Branch == nil {
		commitInfo.Commit.Branch = &pfs.Branch{Name: "master", Repo: commitInfo.Commit.Repo}
	}
	return &Commit{
		IntID:          col.IntID,
		CommitID:       pfsdb.CommitKey(commitInfo.Commit), //this might have to be pfsdb.CommitKey(commit)
		CommitSetID:    commitInfo.Commit.Id,
		RepoName:       commitInfo.Commit.Repo.Name,
		RepoType:       commitInfo.Commit.Repo.Type,
		ProjectName:    commitInfo.Commit.Repo.Project.Name,
		BranchID:       commitInfo.Commit.Branch.Name,
		Origin:         commitInfo.Origin.Kind.String(),
		StartTime:      sanitizeTimestamppb(commitInfo.Started),
		FinishingTime:  sanitizeTimestamppb(commitInfo.Finishing),
		FinishedTime:   sanitizeTimestamppb(commitInfo.Finished),
		Description:    commitInfo.Description,
		CompactingTime: durationpbToBigInt(commitInfo.Details.CompactingTime),
		ValidatingTime: durationpbToBigInt(commitInfo.Details.ValidatingTime),
		Size:           commitInfo.Details.SizeBytes,
		Error:          commitInfo.Error,
		ChildCommits:   commitInfo.ChildCommits,
		ParentCommit:   commitInfo.ParentCommit,
		CreatedAt:      col.CreatedAt,
		UpdatedAt:      col.UpdatedAt,
	}, nil
}

func migrateRelatives(ctx context.Context, tx *pachsql.Tx, commit *Commit) error {
	ancestryQueryTemplate := `
		INSERT INTO pfs.commit_ancestry
		(from_id, to_id)
		VALUES %s
		ON CONFLICT DO NOTHING;
		`
	valuesTemplate := `($1, (SELECT int_id FROM pfs.commits WHERE commit_id=$%d))`
	params := []any{commit.IntID}
	queryVarNum := 2
	values := make([]string, 0)
	for _, child := range commit.ChildCommits {
		values = append(values, fmt.Sprintf(valuesTemplate, queryVarNum))
		params = append(params, pfsdb.CommitKey(child))
		queryVarNum++
	}
	if commit.ParentCommit != nil {
		values = append(values, fmt.Sprintf(`((SELECT int_id FROM pfs.commits WHERE commit_id=$%d), $1)`, queryVarNum))
		params = append(params, pfsdb.CommitKey(commit.ParentCommit))
	}
	query := fmt.Sprintf(ancestryQueryTemplate, strings.Join(values, ","))
	_, err := tx.ExecContext(ctx, query, params...)
	return errors.Wrap(err, fmt.Sprintf("putting commit relative relationships: query: %s params: %v", query, params))
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
			return nil, nil, nil, errors.Errorf("branch not found: %s", trigger.Branch)
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

// alterCommitsTablePreMigration adds useful new columns to pfs.commits table.
// Note that this is not the end all be all. We will need to make more changes after data has been migrated.
// TODO
//   - rename int_id to id? This will requires changing all references as well.
//   - make repo_id not null
//   - make origin not null
//   - make updated_at not null and default to current timestamp
//   - branch_id_str is a metadata reference to the branches table. Today this points to collections, but tomorrow it should
//     point to pfs.branches. Once the PFS master watches branches, we will no longer need this column.
func alterCommitsTablePreMigration(ctx context.Context, tx *pachsql.Tx) error {
	query := `
	CREATE TYPE pfs.commit_origin AS ENUM ('ORIGIN_KIND_UNKNOWN', 'USER', 'AUTO', 'FSCK');

	ALTER TABLE IF EXISTS pfs.commits
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
		ADD COLUMN IF NOT EXISTS updated_at timestamptz DEFAULT CURRENT_TIMESTAMP,
		ADD COLUMN IF NOT EXISTS branch_id_str text;

	CREATE TRIGGER set_updated_at
		BEFORE UPDATE ON pfs.commits
		FOR EACH ROW EXECUTE PROCEDURE core.set_updated_at_to_now();
	`
	if _, err := tx.ExecContext(ctx, query); err != nil {
		return errors.Wrap(err, "altering commits table")
	}
	return nil
}

// commitAncestry is how we model the commit graph within a single repo.
func createCommitAncestryTable(ctx context.Context, tx *pachsql.Tx) error {
	query := `
	CREATE TABLE IF NOT EXISTS pfs.commit_ancestry (
		from_id bigint REFERENCES pfs.commits(int_id),
		to_id bigint REFERENCES pfs.commits(int_id) UNIQUE,
		PRIMARY KEY (from_id, to_id)
	);
	CREATE INDEX ON pfs.commit_ancestry (to_id, from_id);
	CREATE INDEX ON pfs.commit_ancestry (to_id);
	`
	if _, err := tx.ExecContext(ctx, query); err != nil {
		return errors.Wrap(err, "creating commit_ancestry table")
	}
	return nil
}

func createNotifyCommitsTrigger(ctx context.Context, tx *pachsql.Tx) error {
	query := `
	CREATE OR REPLACE FUNCTION pfs.notify_commits() RETURNS TRIGGER AS $$
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

func updateCommitsSchema(ctx context.Context, tx *pachsql.Tx) error {
	if err := alterCommitsTablePreMigration(ctx, tx); err != nil {
		return err
	}
	if err := createCommitAncestryTable(ctx, tx); err != nil {
		return err
	}
	return nil
}

func migrateCommits(ctx context.Context, tx *pachsql.Tx) error {
	return migrateCommitsFromCollections(ctx, tx)
}

func migrateBranches(ctx context.Context, env migrations.Env) error {
	tx := env.Tx
	if _, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS pfs.branches (
			id bigserial PRIMARY KEY,
			name text NOT NULL,
			head bigint REFERENCES pfs.commits(int_id) NOT NULL,
			repo_id bigint REFERENCES pfs.repos(id) NOT NULL,
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
			PRIMARY KEY (from_id, to_id)
		);
	`); err != nil {
		return errors.Wrap(err, "creating branch_provenance table")
	}
	if _, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS pfs.branch_triggers (
			from_branch_id bigint REFERENCES pfs.branches(id) NOT NULL,
			to_branch_id bigint REFERENCES pfs.branches(id) NOT NULL,
			cron_spec text,
			rate_limit_spec text,
			size text,
			num_commits bigint,
			all_conditions bool,
			PRIMARY KEY (from_branch_id, to_branch_id)
		);
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

func sanitizeTimestamppb(timestamp *timestamppb.Timestamp) time.Time {
	if timestamp != nil {
		return timestamp.AsTime()
	}
	return time.Time{}
}

func durationpbToBigInt(duration *durationpb.Duration) int64 {
	if duration != nil {
		return duration.Seconds
	}
	return 0
}
