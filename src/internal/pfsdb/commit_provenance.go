package pfsdb

import (
	"context"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
)

func ResolveCommitProvenance(ctx context.Context, tx *pachsql.Tx, repo *pfs.Repo, commitSet string) (string, error) {
	cs, err := CommitSetProvenance(ctx, tx, commitSet)
	if err != nil {
		return "", err
	}
	for _, c := range cs {
		if strings.HasPrefix(c, repo.String()+"@") {
			return c, nil
		}
	}
	// TODO: Commit proto is FUBAR
	return "", pfsserver.ErrCommitNotFound{Commit: &pfs.Commit{ID: repo.String() + "@" + commitSet}}
}

// CommitSetProvenance returns all the commit IDs that are in the provenance
// of all the commits in this commit set.
func CommitSetProvenance(ctx context.Context, tx *pachsql.Tx, id string) ([]string, error) {
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
	cs := make([]string, 0)
	for rows.Next() {
		var commit string
		if err := rows.Scan(&commit); err != nil {
			return nil, err
		}
		cs = append(cs, commit)
	}
	return cs, nil
}

// CommitSetSubvenance returns all the commit IDs that contain commits in this commit set in their
// full provenance
func CommitSetSubvenance(ctx context.Context, tx *pachsql.Tx, id string) ([]string, error) {
	q := `
          WITH RECURSIVE prov(from_id, to_id) AS (
            SELECT from_id, to_id 
            FROM pfs.commit_provenance JOIN pfs.commits ON int_id = to_id 
            WHERE commit_set_id = $1
           UNION ALL
            SELECT cp.from_id, cp.to_id
            FROM prov p, pfs.commit_provenance cp
            WHERE cp.to_id = p.from_id
          )
          SELECT DISTINCT commit_id 
          FROM pfs.commits, prov 
          WHERE int_id = prov.from_id AND commit_set_id != $2;`
	rows, err := tx.QueryxContext(ctx, q, id, id)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	defer rows.Close()
	cs := make([]string, 0)
	for rows.Next() {
		var commit string
		if err := rows.Scan(&commit); err != nil {
			return nil, err
		}
		cs = append(cs, commit)
	}
	return cs, nil
}

func AddCommit(ctx context.Context, tx *pachsql.Tx, commit, commitSet string) error {
	stmt := `INSERT INTO pfs.commits(commit_id, commit_set_id) VALUES ($1, $2)`
	_, err := tx.ExecContext(ctx, stmt, commit, commitSet)
	return errors.EnsureStack(err)
}

func AddCommitProvenance(ctx context.Context, tx *pachsql.Tx, from, to string) error {
	query := `SELECT int_id, commit_id FROM pfs.commits WHERE commit_id = $1 OR commit_id = $2;`
	rows, err := tx.QueryxContext(ctx, query, from, to)
	if err != nil {
		return err
	}
	var fromId, toId int
	for rows.Next() {
		var tmp int
		var commitId string
		rows.Scan(&tmp, &commitId)
		if commitId == from {
			fromId = tmp
		} else {
			toId = tmp
		}
	}
	return addCommitProvenance(ctx, tx, fromId, toId)
}

func addCommitProvenance(ctx context.Context, tx *pachsql.Tx, from, to int) error {
	stmt := `INSERT INTO pfs.commit_provenance(from_id, to_id) VALUES ($1, $2)`
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
		from_id INT8 NOT NULL,
		to_id INT8 NOT NULL,
		PRIMARY KEY (from_id, to_id)
	);

	CREATE INDEX ON pfs.commit_provenance (
		to_id,
		from_id
	);
`
