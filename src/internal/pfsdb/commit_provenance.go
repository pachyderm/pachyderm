package pfsdb

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
)

// returns the commit of a certain repo in a commit set.
func ResolveCommitProvenance(tx *pachsql.Tx, repo *pfs.Repo, commitSet string) (*pfs.Commit, error) {
	cs, err := CommitSetProvenance(tx, commitSet)
	if err != nil {
		return nil, err
	}
	for _, c := range cs {
		if RepoKey(c.Repo) == RepoKey(repo) {
			return c, nil
		}
	}
	return nil, pfsserver.ErrCommitNotFound{Commit: &pfs.Commit{Repo: repo, Id: commitSet}}
}

// CommitSetProvenance returns all the commit IDs that are in the provenance
// of all the commits in this commit set.
//
// TODO(provenance): is 'SELECT DISTINCT commit_id' a performance concern?
func CommitSetProvenance(tx *pachsql.Tx, id string) (_ []*pfs.Commit, retErr error) {
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
          WHERE int_id = prov.to_id AND commit_set_id != $1;`
	rows, err := tx.Queryx(q, id)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	defer errors.Close(&retErr, rows, "close rows")
	cs := make([]*pfs.Commit, 0)
	for rows.Next() {
		var commit string
		if err := rows.Scan(&commit); err != nil {
			return nil, errors.EnsureStack(err)
		}
		cs = append(cs, ParseCommit(commit))
	}
	if err = rows.Err(); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return cs, nil
}

func CommitDirectProvenance(tx *pachsql.Tx, id CommitID) ([]*pfs.Commit, error) {
	var commitKeys []string
	query := `
	SELECT DISTINCT commit.commit_id
	FROM pfs.commit_provenance prov
		JOIN pfs.commits commit ON prov.to_id = commit.int_id
	WHERE prov.from_id = $1	
	`
	if err := tx.Select(&commitKeys, query, id); err != nil {
		return nil, err
	}
	var commits []*pfs.Commit
	for _, commitID := range commitKeys {
		commits = append(commits, ParseCommit(commitID))
	}
	return commits, nil
}

// CommitSetSubvenance returns all the commit IDs that contain commits in this commit set in their
// full (transitive) provenance
func CommitSetSubvenance(tx *pachsql.Tx, id string) (_ []*pfs.Commit, retErr error) {
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
          WHERE int_id = subv.from_id AND commit_set_id != $1;`
	rows, err := tx.Queryx(q, id)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	defer errors.Close(&retErr, rows, "close rows")
	cs := make([]*pfs.Commit, 0)
	for rows.Next() {
		var commit string
		if err := rows.Scan(&commit); err != nil {
			return nil, errors.EnsureStack(err)
		}
		cs = append(cs, ParseCommit(commit))
	}
	if err = rows.Err(); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return cs, nil
}

func AddCommit(tx *pachsql.Tx, commit *pfs.Commit) error {
	stmt := `INSERT INTO pfs.commits(commit_id, commit_set_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;`
	_, err := tx.Exec(stmt, CommitKey(commit), commit.Id)
	return errors.Wrapf(err, "insert commit %q into pfs.commits", CommitKey(commit))
}

func getCommitTableID(tx *pachsql.Tx, commit *pfs.Commit) (_ int, retErr error) {
	commitKey := CommitKey(commit)
	query := `SELECT int_id FROM pfs.commits WHERE commit_id = $1;`
	rows, err := tx.Queryx(query, commitKey)
	if err != nil {
		return 0, errors.EnsureStack(err)
	}
	defer errors.Close(&retErr, rows, "close rows")
	var id int
	for rows.Next() {
		if err := rows.Scan(&id); err != nil {
			return 0, errors.EnsureStack(err)
		}
	}
	if err = rows.Err(); err != nil {
		return 0, errors.EnsureStack(err)
	}
	return id, nil
}

func AddCommitProvenance(tx *pachsql.Tx, from, to *pfs.Commit) error {
	fromId, err := getCommitTableID(tx, from)
	if err != nil {
		return errors.Wrapf(err, "get int id for 'from' commit, %q", CommitKey(from))
	}
	toId, err := getCommitTableID(tx, to)
	if err != nil {
		return errors.Wrapf(err, "get int id for 'to' commit, %q", CommitKey(to))
	}
	return addCommitProvenance(tx, fromId, toId)
}

func addCommitProvenance(tx *pachsql.Tx, from, to int) error {
	stmt := `INSERT INTO pfs.commit_provenance(from_id, to_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;`
	_, err := tx.Exec(stmt, from, to)
	return errors.Wrapf(err, "add commit provenance")
}
