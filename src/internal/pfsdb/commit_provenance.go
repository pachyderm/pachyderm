package pfsdb

import (
	"context"
	"fmt"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
)

// returns the commit of a certain repo in a commit set.
func ResolveCommitProvenance(ctx context.Context, tx *pachsql.Tx, repo *pfs.Repo, commitSet string) (*pfs.Commit, error) {
	cs, err := CommitSetProvenance(ctx, tx, commitSet)
	if err != nil {
		return nil, err
	}
	matches := make([]*pfs.Commit, 0)
	for _, c := range cs {
		if RepoKey(c.Repo) == RepoKey(repo) {
			matches = append(matches, c)
		}
	}
	if len(matches) == 0 {
		return nil, pfsserver.ErrCommitNotFound{Commit: &pfs.Commit{Repo: repo, ID: commitSet}}
	}
	if len(matches) > 1 {
		return nil, pfsserver.ErrAmbiguousCommit{Commit: &pfs.Commit{Repo: repo, ID: commitSet}, PossibleCommits: matches}
	}
	return matches[0], nil
}

func CommitProvenance(ctx context.Context, tx *pachsql.Tx, repo *pfs.Repo, commitSet string) ([]*pfs.Commit, error) {
	commitKey := CommitKey(&pfs.Commit{
		Repo: repo,
		ID:   commitSet,
	})
	query := `SELECT commit_id FROM pfs.commits 
                  WHERE int_id IN (       
                      SELECT to_id FROM pfs.commits JOIN pfs.commit_provenance 
                        ON int_id = from_id 
                      WHERE commit_id = $1
                  );`
	rows, err := tx.QueryxContext(ctx, query, commitKey)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	commitProvenance := make([]*pfs.Commit, 0)
	for rows.Next() {
		var commitId string
		if err := rows.Scan(&commitId); err != nil {
			return nil, errors.EnsureStack(err)
		}
		commitProvenance = append(commitProvenance, ParseCommit(commitId))
	}
	return commitProvenance, nil
}

// CommitSetProvenance returns all the commit IDs that are in the provenance
// of all the commits in this commit set.
func CommitSetProvenance(ctx context.Context, tx *pachsql.Tx, id string) ([]*pfs.Commit, error) {
	q := `
          WITH RECURSIVE prov(from_id, to_id) AS (
            SELECT from_id, to_id 
            FROM pfs.commit_provenance JOIN pfs.commits ON int_id = from_id 
            WHERE commit_set_id = $1
           UNION ALL
            SELECT cp.from_id, cp.to_id
            FROM prov p, pfs.commit_provenance cp
            WHERE cp.from_id = p.to_id
          )
          SELECT DISTINCT commit_id 
          FROM pfs.commits, prov 
          WHERE int_id = prov.to_id AND commit_set_id != $2;`
	rows, err := tx.QueryxContext(ctx, q, id, id)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	defer rows.Close()
	cs := make([]*pfs.Commit, 0)
	for rows.Next() {
		var commit string
		if err := rows.Scan(&commit); err != nil {
			return nil, errors.EnsureStack(err)
		}
		cs = append(cs, ParseCommit(commit))
	}
	return cs, nil
}

// CommitSetSubvenance returns all the commit IDs that contain commits in this commit set in their
// full provenance
func CommitSetSubvenance(ctx context.Context, tx *pachsql.Tx, id string) ([]*pfs.Commit, error) {
	q := `
          WITH RECURSIVE subv(from_id, to_id) AS (
            SELECT from_id, to_id 
            FROM pfs.commit_provenance JOIN pfs.commits ON int_id = to_id 
            WHERE commit_set_id = $1
           UNION ALL
            SELECT cp.from_id, cp.to_id
            FROM subv s, pfs.commit_provenance cp
            WHERE cp.to_id = s.from_id
          )
          SELECT DISTINCT commit_id 
          FROM pfs.commits, subv 
          WHERE int_id = subv.from_id AND commit_set_id != $2;`
	rows, err := tx.QueryxContext(ctx, q, id, id)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	defer rows.Close()
	cs := make([]*pfs.Commit, 0)
	for rows.Next() {
		var commit string
		if err := rows.Scan(&commit); err != nil {
			return nil, errors.EnsureStack(err)
		}
		cs = append(cs, ParseCommit(commit))
	}
	return cs, nil
}

func AddCommit(ctx context.Context, tx *pachsql.Tx, commit *pfs.Commit) error {
	stmt := `INSERT INTO pfs.commits(commit_id, commit_set_id) VALUES ($1, $2);`
	_, err := tx.ExecContext(ctx, stmt, CommitKey(commit), commit.ID)
	return errors.Wrapf(err, "insert commit %q into pfs.commits", CommitKey(commit))
}

func DeleteCommit(ctx context.Context, tx *pachsql.Tx, commitKey string) error {
	id, err := getCommitTableID(ctx, tx, commitKey)
	if err != nil {
		return err
	}
	stmt := `DELETE FROM pfs.commits WHERE int_id = $1;`
	_, err = tx.ExecContext(ctx, stmt, id)
	return errors.Wrapf(err, "delete from pfs.commits")
}

func getCommitTableID(ctx context.Context, tx *pachsql.Tx, commitKey string) (int, error) {
	query := `SELECT int_id FROM pfs.commits WHERE commit_id = $1;`
	rows, err := tx.QueryxContext(ctx, query, commitKey)
	if err != nil {
		return 0, errors.EnsureStack(err)
	}
	var id int
	for rows.Next() {
		if err := rows.Scan(&id); err != nil {
			return 0, errors.EnsureStack(err)
		}
	}
	return id, nil
}

func AddCommitProvenance(ctx context.Context, tx *pachsql.Tx, from, to *pfs.Commit) error {
	query := `SELECT int_id, commit_id FROM pfs.commits WHERE commit_id = $1 OR commit_id = $2;`
	rows, err := tx.QueryxContext(ctx, query, CommitKey(from), CommitKey(to))
	if err != nil {
		return errors.EnsureStack(err)
	}
	var fromId, toId int
	var count int
	for rows.Next() {
		count++
		var tmp int
		var commitId string
		if err := rows.Scan(&tmp, &commitId); err != nil {
			return errors.EnsureStack(err)
		}
		if commitId == CommitKey(from) {
			fromId = tmp
		} else {
			toId = tmp
		}
	}
	if count != 2 {
		msg := fmt.Sprintf("expected two existing commits, got %v.", count)
		if count == 1 {
			var found *pfs.Commit
			var missed *pfs.Commit
			if fromId != 0 {
				found = from
				missed = to
			} else {
				found = to
				missed = from
			}
			msg += fmt.Sprintf(" found %q, missed %q", CommitKey(found), CommitKey(missed))
		}
		return errors.Errorf(msg)
	}
	return addCommitProvenance(ctx, tx, fromId, toId)
}

func addCommitProvenance(ctx context.Context, tx *pachsql.Tx, from, to int) error {
	stmt := `INSERT INTO pfs.commit_provenance(from_id, to_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;`
	_, err := tx.ExecContext(ctx, stmt, from, to)
	return errors.EnsureStack(err)
}

func SetupCommitProvenanceV0(ctx context.Context, tx *pachsql.Tx) error {
	_, err := tx.ExecContext(ctx, schema)
	return errors.EnsureStack(err)
}

var schema = `
	CREATE TABLE pfs.commits (
		int_id BIGSERIAL PRIMARY KEY,
		commit_id VARCHAR(4096) UNIQUE,
                commit_set_id VARCHAR(4096),
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE pfs.commit_provenance (
		from_id BIGSERIAL NOT NULL,
		to_id BIGSERIAL NOT NULL,
		PRIMARY KEY (from_id, to_id),
                CONSTRAINT fk_from_commit
                  FOREIGN KEY(from_id) 
	          REFERENCES pfs.commits(int_id)
	          ON DELETE CASCADE,
                CONSTRAINT fk_to_commit
                  FOREIGN KEY(to_id) 
	          REFERENCES pfs.commits(int_id)
	          ON DELETE CASCADE
	);

	CREATE INDEX ON pfs.commit_provenance (
		to_id,
		from_id
	);
`
