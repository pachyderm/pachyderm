package server

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
)

var errNoTotalFileSet = errors.Errorf("no total fileset")

const commitTrackerPrefix = "commit/"

type commitStore interface {
	// AddFileSet appends a fileset to the diff.
	AddFileSet(ctx context.Context, commit *pfsdb.Commit, filesetID fileset.ID) error
	// AddFileSetTx is identical to AddFileSet except it runs in the provided transaction.
	AddFileSetTx(tx *pachsql.Tx, commit *pfsdb.Commit, filesetID fileset.ID) error
	// SetTotalFileSet sets the total file set for the commit, overwriting whatever is there.
	SetTotalFileSet(ctx context.Context, commit *pfsdb.Commit, id fileset.ID) error
	// SetTotalFileSetTx is like SetTotalFileSet, but in a transaction
	SetTotalFileSetTx(tx *pachsql.Tx, commit *pfsdb.Commit, id fileset.ID) error
	// SetDiffFileSet sets the diff file set for the commit, overwriting whatever is there.
	SetDiffFileSet(ctx context.Context, commit *pfsdb.Commit, id fileset.ID) error
	// SetDiffFileSetTx is like SetDiffFileSet, but in a transaction
	SetDiffFileSetTx(tx *pachsql.Tx, commit *pfsdb.Commit, id fileset.ID) error
	// GetTotalFileSet returns the total file set for a commit.
	GetTotalFileSet(ctx context.Context, commit *pfsdb.Commit) (*fileset.ID, error)
	// GetTotalFileSetTx is like GetTotalFileSet, but in a transaction
	GetTotalFileSetTx(tx *pachsql.Tx, commit *pfsdb.Commit) (*fileset.ID, error)
	// GetDiffFileSet returns the diff file set for a commit
	GetDiffFileSet(ctx context.Context, commit *pfsdb.Commit) (*fileset.ID, error)
	// DropFileSets clears the diff and total file sets for the commit.
	DropFileSets(ctx context.Context, commit *pfsdb.Commit) error
	// DropFileSetsTx is identical to DropFileSets except it runs in the provided transaction.
	DropFileSetsTx(tx *pachsql.Tx, commit *pfsdb.Commit) error
}

var _ commitStore = &postgresCommitStore{}

// TODO: add deleter for the commitStore and stop making permanent filesets, keep the filesets
// around, by referencing them with commit-fileset objects.
type postgresCommitStore struct {
	db *pachsql.DB
	s  *fileset.Storage
	tr track.Tracker
}

func newPostgresCommitStore(db *pachsql.DB, tr track.Tracker, s *fileset.Storage) *postgresCommitStore {
	return &postgresCommitStore{
		db: db,
		s:  s,
		tr: tr,
	}
}

func (cs *postgresCommitStore) AddFileSet(ctx context.Context, commit *pfsdb.Commit, id fileset.ID) error {
	return dbutil.WithTx(ctx, cs.db, func(ctx context.Context, tx *pachsql.Tx) error {
		return cs.AddFileSetTx(tx, commit, id)
	})
}

func (cs *postgresCommitStore) AddFileSetTx(tx *pachsql.Tx, commit *pfsdb.Commit, id fileset.ID) error {
	id2, err := cs.s.CloneTx(tx, id, defaultTTL)
	if err != nil {
		return err
	}
	id = *id2

	oid := commitDiffTrackerID(commit, id)
	pointsTo := []string{id.TrackerID()}
	if _, err := tx.Exec(
		`INSERT INTO pfs.commit_diffs (commit_int_id, fileset_id) VALUES ($1, $2)
`, commit.ID, id); err != nil {
		return errors.Wrap(err, "add file set tx")
	}
	if _, err := tx.Exec(
		`UPDATE pfs.commits SET total_fileset_id = NULL WHERE int_id = $1
		`, commit.ID); err != nil {
		return errors.Wrap(err, "add file set tx")
	}
	return errors.Wrap(cs.tr.CreateTx(tx, oid, pointsTo, track.NoTTL), "add file set tx")
}

func (cs *postgresCommitStore) GetTotalFileSet(ctx context.Context, commit *pfsdb.Commit) (*fileset.ID, error) {
	var id *fileset.ID
	if err := dbutil.WithTx(ctx, cs.db, func(ctx context.Context, tx *pachsql.Tx) error {
		var err error
		id, err = cs.GetTotalFileSetTx(tx, commit)
		return err
	}); err != nil {
		return nil, err
	}
	return id, nil
}

func (cs *postgresCommitStore) GetTotalFileSetTx(tx *pachsql.Tx, commit *pfsdb.Commit) (*fileset.ID, error) {
	id, err := getTotal(tx, commit)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if id == nil {
		return nil, errNoTotalFileSet
	}
	return cs.s.CloneTx(tx, *id, defaultTTL)
}

func (cs *postgresCommitStore) GetDiffFileSet(ctx context.Context, commit *pfsdb.Commit) (*fileset.ID, error) {
	var ids []fileset.ID
	if err := dbutil.WithTx(ctx, cs.db, func(ctx context.Context, tx *pachsql.Tx) error {
		var err error
		ids, err = getDiff(tx, commit)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return cs.s.Compose(ctx, ids, defaultTTL)
}

func (cs *postgresCommitStore) SetTotalFileSet(ctx context.Context, commit *pfsdb.Commit, id fileset.ID) error {
	return dbutil.WithTx(ctx, cs.db, func(ctx context.Context, tx *pachsql.Tx) error {
		return cs.SetTotalFileSetTx(tx, commit, id)
	})
}

func (cs *postgresCommitStore) SetTotalFileSetTx(tx *pachsql.Tx, commit *pfsdb.Commit, id fileset.ID) error {
	if err := dropTotal(tx, cs.tr, commit); err != nil {
		return err
	}
	return setTotal(tx, cs.tr, commit, id)
}

func (cs *postgresCommitStore) SetDiffFileSet(ctx context.Context, commit *pfsdb.Commit, id fileset.ID) error {
	return dbutil.WithTx(ctx, cs.db, func(ctx context.Context, tx *pachsql.Tx) error {
		return cs.SetDiffFileSetTx(tx, commit, id)
	})
}

func (cs *postgresCommitStore) SetDiffFileSetTx(tx *pachsql.Tx, commit *pfsdb.Commit, id fileset.ID) error {
	if err := cs.dropDiff(tx, commit); err != nil {
		return err
	}
	return setDiff(tx, cs.tr, commit, id)
}

func (cs *postgresCommitStore) DropFileSets(ctx context.Context, commit *pfsdb.Commit) error {
	return dbutil.WithTx(ctx, cs.db, func(ctx context.Context, tx *pachsql.Tx) error {
		return cs.DropFileSetsTx(tx, commit)
	})
}

func (cs *postgresCommitStore) DropFileSetsTx(tx *pachsql.Tx, commit *pfsdb.Commit) error {
	if err := dropTotal(tx, cs.tr, commit); err != nil {
		return errors.Wrap(err, "drop file set tx")
	}
	return cs.dropDiff(tx, commit)
}

func (cs *postgresCommitStore) dropDiff(tx *pachsql.Tx, commit *pfsdb.Commit) error {
	diffIDs, err := getDiff(tx, commit)
	if err != nil {
		return err
	}
	for _, diffID := range diffIDs {
		trackID := commitDiffTrackerID(commit, diffID)
		if err := cs.tr.DeleteTx(tx, trackID); err != nil {
			return errors.Wrap(err, "drop diff")
		}
	}
	if _, err := tx.Exec(`DELETE FROM pfs.commit_diffs WHERE commit_int_id = $1`, commit.ID); err != nil {
		return errors.Wrap(err, "drop diff")
	}
	return nil
}

func getDiff(tx *pachsql.Tx, commit *pfsdb.Commit) ([]fileset.ID, error) {
	var ids []fileset.ID
	if err := tx.Select(&ids,
		`SELECT fileset_id FROM pfs.commit_diffs
		WHERE commit_int_id = $1
		ORDER BY num
		`, commit.ID); err != nil {
		return nil, errors.Wrap(err, "get diff")
	}
	return ids, nil
}

func getTotal(tx *pachsql.Tx, commit *pfsdb.Commit) (*fileset.ID, error) {
	var id fileset.ID
	if err := tx.Get(&id, `SELECT total_fileset_id FROM pfs.commits WHERE int_id = $1 AND total_fileset_id IS NOT NULL`,
		commit.ID); err != nil {
		return nil, errors.Wrap(err, "get total")
	}
	return &id, nil
}

func dropTotal(tx *pachsql.Tx, tr track.Tracker, commit *pfsdb.Commit) error {
	id, err := getTotal(tx, commit)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	trackID := commitTotalTrackerID(commit, *id)
	if err := tr.DeleteTx(tx, trackID); err != nil {
		return errors.Wrap(err, "drop total")
	}
	_, err = tx.Exec(`UPDATE pfs.commits SET total_fileset_id = NULL WHERE int_id = $1`, commit.ID)
	return errors.Wrap(err, "drop total")
}

func setTotal(tx *pachsql.Tx, tr track.Tracker, commit *pfsdb.Commit, id fileset.ID) error {
	oid := commitTotalTrackerID(commit, id)
	pointsTo := []string{id.TrackerID()}
	if err := tr.CreateTx(tx, oid, pointsTo, track.NoTTL); err != nil {
		return errors.Wrap(err, "set total")
	}
	_, err := tx.Exec(`UPDATE pfs.commits SET total_fileset_id = $1 WHERE int_id = $2`, id, commit.ID)
	return errors.Wrap(err, "set total")
}

func setDiff(tx *pachsql.Tx, tr track.Tracker, commit *pfsdb.Commit, id fileset.ID) error {
	oid := commitDiffTrackerID(commit, id)
	pointsTo := []string{id.TrackerID()}
	if err := tr.CreateTx(tx, oid, pointsTo, track.NoTTL); err != nil {
		return errors.Wrap(err, "set diff")
	}
	if commit.ID == 0 {
		return errors.New(fmt.Sprintf("cannot set diff for commit %v when ID is 0", commit.CommitInfo.Commit.Key()))
	}
	_, err := tx.Exec(`INSERT INTO pfs.commit_diffs (commit_int_id, fileset_id)
	VALUES ($1, $2)
	`, commit.ID, id)
	return errors.Wrap(err, "set diff")
}

func commitDiffTrackerID(commit *pfsdb.Commit, fs fileset.ID) string {
	return commitTrackerPrefix + fmt.Sprintf("%d", commit.ID) + "/diff/" + fs.HexString()
}

func commitTotalTrackerID(commit *pfsdb.Commit, fs fileset.ID) string {
	return commitTrackerPrefix + fmt.Sprintf("%d", commit.ID) + "/total/" + fs.HexString()
}
