package server

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

type commitStore interface {
	// AddFileset appends a fileset to the diff.
	AddFileset(ctx context.Context, commit *pfs.Commit, filesetID fileset.ID) error
	// SetTotalFileset sets the total fileset for the commit, overwriting whatever is there.
	SetTotalFileset(ctx context.Context, commit *pfs.Commit, id fileset.ID) error
	// GetTotalFileset returns the total fileset for a commit.
	GetTotalFileset(ctx context.Context, commit *pfs.Commit) (*fileset.ID, error)
	// GetDiffFileset returns the diff fileset for a commit
	GetDiffFileset(ctx context.Context, commit *pfs.Commit) (*fileset.ID, error)
	// DropFilesets clears the diff and total filesets for the commit.
	DropFilesets(ctx context.Context, commit *pfs.Commit) error
	// DropFilesetsTx clears the diff and total filesets for the commit.
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
	// clone to remove the ttl.
	id2, err := cs.s.Clone(ctx, id, track.NoTTL)
	if err != nil {
		return err
	}
	return dbutil.WithTx(ctx, cs.db, func(tx *sqlx.Tx) error {
		if _, err := cs.db.ExecContext(ctx,
			`INSERT INTO pfs.commit_diffs (commit_id, fileset_id)
		VALUES ($1, $2)
	`, commit.ID, *id2); err != nil {
			return err
		}
		if _, err := cs.db.ExecContext(ctx,
			`DELETE FROM pfs.commit_totals WHERE commit_id = $1
			`, commit.ID); err != nil {
			return err
		}
		return nil
	})
}

func (cs *postgresCommitStore) GetTotalFileset(ctx context.Context, commit *pfs.Commit) (*fileset.ID, error) {
	id, err := getTotal(ctx, cs.db, commit)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == sql.ErrNoRows {
		return cs.s.Compose(ctx, nil, defaultTTL)
	}
	return cs.s.Clone(ctx, *id, defaultTTL)
}

func (cs *postgresCommitStore) GetDiffFileset(ctx context.Context, commit *pfs.Commit) (*fileset.ID, error) {
	ids, err := getDiff(ctx, cs.db, commit)
	if err != nil {
		return nil, err
	}
	return cs.s.Compose(ctx, ids, defaultTTL)
}

func (cs *postgresCommitStore) SetTotalFileset(ctx context.Context, commit *pfs.Commit, id fileset.ID) error {
	_, err := cs.db.ExecContext(ctx,
		`INSERT INTO pfs.commit_totals (commit_id, fileset_id)
		VALUES ($1, $2)
		ON CONFLICT (commit_id) DO UPDATE
		SET fileset_id = $2
		WHERE commit_totals.commit_id = $1
		`, commit.ID, id)
	return err
}

func (cs *postgresCommitStore) DropFilesets(ctx context.Context, commit *pfs.Commit) error {
	return dbutil.WithTx(ctx, cs.db, func(tx *sqlx.Tx) error {
		return cs.DropFilesetsTx(tx, commit)
	})
}

func (cs *postgresCommitStore) DropFilesetsTx(tx *sqlx.Tx, commit *pfs.Commit) error {
	ctx := context.Background()
	if err := cs.dropTotal(ctx, tx, commit); err != nil {
		return err
	}
	return cs.dropDiff(ctx, tx, commit)
}

func (cs *postgresCommitStore) dropDiff(ctx context.Context, db dbutil.Interface, commit *pfs.Commit) error {
	// TODO: do something about the potential dangling references
	diffIDs, err := getDiff(ctx, db, commit)
	if err != nil {
		return err
	}
	for _, id := range diffIDs {
		if err := cs.s.Drop(ctx, id); err != nil {
			return err
		}
	}
	if _, err := cs.db.ExecContext(ctx, `DELETE FROM pfs.commit_diffs WHERE commit_id = $1`, commit.ID); err != nil {
		return err
	}
	return nil
}

func (cs *postgresCommitStore) dropTotal(ctx context.Context, db dbutil.Interface, commit *pfs.Commit) error {
	id, err := getTotal(ctx, db, commit)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	if err := cs.s.Drop(ctx, *id); err != nil {
		return err
	}
	if _, err := cs.db.ExecContext(ctx, `DELETE FROM pfs.commit_totals WHERE commit_id = $1`, commit.ID); err != nil {
		return err
	}
	return nil
}

func getDiff(ctx context.Context, db dbutil.Interface, commit *pfs.Commit) ([]fileset.ID, error) {
	var ids []fileset.ID
	if err := db.SelectContext(ctx, &ids,
		`SELECT fileset_id FROM pfs.commit_diffs
		WHERE commit_id = $1
		ORDER BY num
		`, commit.ID); err != nil {
		return nil, err
	}
	return ids, nil
}

func getTotal(ctx context.Context, db dbutil.Interface, commit *pfs.Commit) (*fileset.ID, error) {
	var id fileset.ID
	if err := db.GetContext(ctx, &id,
		`SELECT fileset_id FROM pfs.commit_totals
		WHERE commit_id = $1
	`, commit.ID); err != nil {
		return nil, err
	}
	return &id, nil
}

// SetupPostgresCommitStoreV0 runs SQL to setup the commit store.
func SetupPostgresCommitStoreV0(ctx context.Context, tx *sqlx.Tx) error {
	_, err := tx.ExecContext(ctx, `
		CREATE TABLE pfs.commit_diffs (
			commit_id VARCHAR(64) NOT NULL,
			num BIGSERIAL NOT NULL,
			fileset_id VARCHAR(64) NOT NULL,
			PRIMARY KEY(commit_id, num)
		);

		CREATE TABLE pfs.commit_totals (
			commit_id VARCHAR(64) NOT NULL,
			fileset_id VARCHAR(64) NOT NULL,
			PRIMARY KEY(commit_id)
		);
	`)
	return err
}
