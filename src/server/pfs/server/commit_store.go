package server

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

const commitTrackerPrefix = "commit/"

type commitStore interface {
	// AddFileset appends a fileset to the diff.
	AddFileset(ctx context.Context, commit *pfs.Commit, filesetID fileset.ID) error
	// AddFilesetTx is identical to AddFileset except it runs in the provided transaction.
	AddFilesetTx(tx *sqlx.Tx, commit *pfs.Commit, filesetID fileset.ID) error
	// SetTotalFileset sets the total fileset for the commit, overwriting whatever is there.
	SetTotalFileset(ctx context.Context, commit *pfs.Commit, id fileset.ID) error
	// GetTotalFileset returns the total fileset for a commit.
	GetTotalFileset(ctx context.Context, commit *pfs.Commit) (*fileset.ID, error)
	// GetDiffFileset returns the diff fileset for a commit
	GetDiffFileset(ctx context.Context, commit *pfs.Commit) (*fileset.ID, error)
	// DropFilesets clears the diff and total filesets for the commit.
	DropFilesets(ctx context.Context, commit *pfs.Commit) error
	// DropFilesetsTx is identical to DropFilesets except it runs in the provided transaction.
	DropFilesetsTx(tx *sqlx.Tx, commit *pfs.Commit) error
}

var _ commitStore = &postgresCommitStore{}

// TODO: add deleter for the commitStore and stop making permanent filesets, keep the filesets
// around, by referencing them with commit-fileset objects.
type postgresCommitStore struct {
	db *sqlx.DB
	s  *fileset.Storage
	tr track.Tracker
}

func newPostgresCommitStore(db *sqlx.DB, tr track.Tracker, s *fileset.Storage) *postgresCommitStore {
	return &postgresCommitStore{
		db: db,
		s:  s,
		tr: tr,
	}
}

func (cs *postgresCommitStore) AddFileset(ctx context.Context, commit *pfs.Commit, id fileset.ID) error {
	return dbutil.WithTx(ctx, cs.db, func(tx *sqlx.Tx) error {
		return cs.AddFilesetTx(tx, commit, id)
	})
}

func (cs *postgresCommitStore) AddFilesetTx(tx *sqlx.Tx, commit *pfs.Commit, id fileset.ID) error {
	id2, err := cs.s.CloneTx(tx, id, defaultTTL)
	if err != nil {
		return err
	}
	id = *id2

	oid := commitDiffTrackerID(commit, id)
	pointsTo := []string{id.TrackerID()}
	if _, err := tx.Exec(
		`INSERT INTO pfs.commit_diffs (commit_id, fileset_id)
	VALUES ($1, $2)
`, commit.ID, id); err != nil {
		return err
	}
	if _, err := tx.Exec(
		`DELETE FROM pfs.commit_totals WHERE commit_id = $1
		`, commit.ID); err != nil {
		return err
	}
	return cs.tr.CreateTx(tx, oid, pointsTo, track.NoTTL)
}

func (cs *postgresCommitStore) GetTotalFileset(ctx context.Context, commit *pfs.Commit) (*fileset.ID, error) {
	var result *fileset.ID
	if err := dbutil.WithTx(ctx, cs.db, func(tx *sqlx.Tx) error {
		id, err := getTotal(tx, commit)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		if id == nil {
			result, err = cs.s.ComposeTx(tx, nil, defaultTTL)
			return err
		}
		result, err = cs.s.CloneTx(tx, *id, defaultTTL)
		return err
	}); err != nil {
		return nil, err
	}
	return result, nil
}

func (cs *postgresCommitStore) GetDiffFileset(ctx context.Context, commit *pfs.Commit) (*fileset.ID, error) {
	var result *fileset.ID
	if err := dbutil.WithTx(ctx, cs.db, func(tx *sqlx.Tx) error {
		ids, err := getDiff(tx, commit)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		result, err = cs.s.ComposeTx(tx, ids, defaultTTL)
		return err
	}); err != nil {
		return nil, err
	}
	return result, nil
}

func (cs *postgresCommitStore) SetTotalFileset(ctx context.Context, commit *pfs.Commit, id fileset.ID) error {
	return dbutil.WithTx(ctx, cs.db, func(tx *sqlx.Tx) error {
		if err := cs.dropTotal(tx, commit); err != nil {
			return err
		}
		oid := commitTotalTrackerID(commit, id)
		pointsTo := []string{id.TrackerID()}
		if err := cs.tr.CreateTx(tx, oid, pointsTo, track.NoTTL); err != nil {
			return err
		}
		_, err := tx.Exec(`INSERT INTO pfs.commit_totals (commit_id, fileset_id)
		VALUES ($1, $2)
		ON CONFLICT (commit_id) DO UPDATE
		SET fileset_id = $2
		WHERE commit_totals.commit_id = $1
		`, commit.ID, id)
		return err
	})
}

func (cs *postgresCommitStore) DropFilesets(ctx context.Context, commit *pfs.Commit) error {
	return dbutil.WithTx(ctx, cs.db, func(tx *sqlx.Tx) error {
		return cs.DropFilesetsTx(tx, commit)
	})
}

func (cs *postgresCommitStore) DropFilesetsTx(tx *sqlx.Tx, commit *pfs.Commit) error {
	if err := cs.dropTotal(tx, commit); err != nil {
		return err
	}
	return cs.dropDiff(tx, commit)
}

func (cs *postgresCommitStore) dropDiff(tx *sqlx.Tx, commit *pfs.Commit) error {
	diffIDs, err := getDiff(tx, commit)
	if err != nil {
		return err
	}
	for _, diffID := range diffIDs {
		trackID := commitDiffTrackerID(commit, diffID)
		if err := cs.tr.DeleteTx(tx, trackID); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(`DELETE FROM pfs.commit_diffs WHERE commit_id = $1`, commit.ID); err != nil {
		return err
	}
	return nil
}

func (cs *postgresCommitStore) dropTotal(tx *sqlx.Tx, commit *pfs.Commit) error {
	id, err := getTotal(tx, commit)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	trackID := commitTotalTrackerID(commit, *id)
	if err := cs.tr.DeleteTx(tx, trackID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM pfs.commit_totals WHERE commit_id = $1`, commit.ID); err != nil {
		return err
	}
	return nil
}

func getDiff(tx *sqlx.Tx, commit *pfs.Commit) ([]fileset.ID, error) {
	var ids []fileset.ID
	if err := tx.Select(&ids,
		`SELECT fileset_id FROM pfs.commit_diffs
		WHERE commit_id = $1
		ORDER BY num
		`, commit.ID); err != nil {
		return nil, err
	}
	return ids, nil
}

func getTotal(tx *sqlx.Tx, commit *pfs.Commit) (*fileset.ID, error) {
	var id fileset.ID
	if err := tx.Get(&id,
		`SELECT fileset_id FROM pfs.commit_totals
		WHERE commit_id = $1
	`, commit.ID); err != nil {
		return nil, err
	}
	return &id, nil
}

func commitDiffTrackerID(commit *pfs.Commit, fs fileset.ID) string {
	return commitTrackerPrefix + commit.ID + "/diff/" + fs.HexString()
}

func commitTotalTrackerID(commit *pfs.Commit, fs fileset.ID) string {
	return commitTrackerPrefix + commit.ID + "/total/" + fs.HexString()
}

// SetupPostgresCommitStoreV0 runs SQL to setup the commit store.
func SetupPostgresCommitStoreV0(ctx context.Context, tx *sqlx.Tx) error {
	_, err := tx.ExecContext(ctx, `
		CREATE TABLE pfs.commit_diffs (
			commit_id VARCHAR(64) NOT NULL,
			num BIGSERIAL NOT NULL,
			fileset_id UUID NOT NULL,
			PRIMARY KEY(commit_id, num)
		);

		CREATE TABLE pfs.commit_totals (
			commit_id VARCHAR(64) NOT NULL,
			fileset_id UUID NOT NULL,
			PRIMARY KEY(commit_id)
		);
	`)
	return errors.EnsureStack(err)
}
