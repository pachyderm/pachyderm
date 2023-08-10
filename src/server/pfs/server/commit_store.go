package server

import (
	"context"
	"database/sql"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

var errNoTotalFileSet = errors.Errorf("no total fileset")

const commitTrackerPrefix = "commit/"

type commitStore struct {
	db *pachsql.DB
	s  *fileset.Storage
	tr track.Tracker
}

func newCommitStore(db *pachsql.DB, tr track.Tracker, s *fileset.Storage) *commitStore {
	return &commitStore{
		db: db,
		s:  s,
		tr: tr,
	}
}

func (cs *commitStore) AddFileSet(ctx context.Context, commit *pfs.Commit, h fileset.Handle) error {
	return dbutil.WithTx(ctx, cs.db, func(ctx context.Context, tx *pachsql.Tx) error {
		return cs.AddFileSetTx(tx, commit, h)
	})
}

func (cs *commitStore) AddFileSetTx(tx *pachsql.Tx, commit *pfs.Commit, h fileset.Handle) error {
	id, err := cs.s.Import(tx, h)
	if err != nil {
		return err
	}

	oid := commitDiffTrackerID(commit, *id)
	pointsTo := []string{id.TrackerID()}
	if _, err := tx.Exec(
		`INSERT INTO pfs.commit_diffs (commit_id, fileset_id)
	VALUES ($1, $2)
`, pfsdb.CommitKey(commit), id); err != nil {
		return errors.EnsureStack(err)
	}
	if _, err := tx.Exec(
		`DELETE FROM pfs.commit_totals WHERE commit_id = $1
		`, pfsdb.CommitKey(commit)); err != nil {
		return errors.EnsureStack(err)
	}
	return errors.EnsureStack(cs.tr.CreateTx(tx, oid, pointsTo, track.NoTTL))
}

func (cs *commitStore) GetTotalFileSet(ctx context.Context, commit *pfs.Commit) (*fileset.Handle, error) {
	var h *fileset.Handle
	if err := dbutil.WithTx(ctx, cs.db, func(ctx context.Context, tx *pachsql.Tx) error {
		var err error
		h, err = cs.GetTotalFileSetTx(tx, commit)
		return err
	}); err != nil {
		return nil, err
	}
	return h, nil
}

func (cs *commitStore) GetTotalFileSetTx(tx *pachsql.Tx, commit *pfs.Commit) (*fileset.Handle, error) {
	id, err := getTotal(tx, commit)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if id == nil {
		return nil, errNoTotalFileSet
	}
	return cs.s.Export(tx, *id, defaultTTL)
}

func (cs *commitStore) GetDiffFileSet(ctx context.Context, commit *pfs.Commit) (*fileset.Handle, error) {
	var h *fileset.Handle
	if err := dbutil.WithTx(ctx, cs.db, func(ctx context.Context, tx *pachsql.Tx) error {
		ids, err := getDiff(tx, commit)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		h, err = cs.s.ExportCompose(tx, ids, defaultTTL)
		return err
	}); err != nil {
		return nil, err
	}
	return h, nil
}

func (cs *commitStore) SetTotalFileSet(ctx context.Context, commit *pfs.Commit, id fileset.Handle) error {
	return dbutil.WithTx(ctx, cs.db, func(ctx context.Context, tx *pachsql.Tx) error {
		return cs.SetTotalFileSetTx(tx, commit, id)
	})
}

func (cs *commitStore) SetTotalFileSetTx(tx *pachsql.Tx, commit *pfs.Commit, h fileset.Handle) error {
	if err := dropTotal(tx, cs.tr, commit); err != nil {
		return err
	}
	id, err := cs.s.Import(tx, h)
	if err != nil {
		return err
	}
	return setTotal(tx, cs.tr, commit, *id)
}

func (cs *commitStore) SetDiffFileSet(ctx context.Context, commit *pfs.Commit, id fileset.Handle) error {
	return dbutil.WithTx(ctx, cs.db, func(ctx context.Context, tx *pachsql.Tx) error {
		return cs.SetDiffFileSetTx(tx, commit, id)
	})
}

func (cs *commitStore) SetDiffFileSetTx(tx *pachsql.Tx, commit *pfs.Commit, h fileset.Handle) error {
	if err := cs.dropDiff(tx, commit); err != nil {
		return err
	}
	id, err := cs.s.Import(tx, h)
	if err != nil {
		return err
	}
	return setDiff(tx, cs.tr, commit, *id)
}

func (cs *commitStore) DropFileSets(ctx context.Context, commit *pfs.Commit) error {
	return dbutil.WithTx(ctx, cs.db, func(ctx context.Context, tx *pachsql.Tx) error {
		return cs.DropFileSetsTx(tx, commit)
	})
}

func (cs *commitStore) DropFileSetsTx(tx *pachsql.Tx, commit *pfs.Commit) error {
	if err := dropTotal(tx, cs.tr, commit); err != nil {
		return errors.EnsureStack(err)
	}
	return cs.dropDiff(tx, commit)
}

func (cs *commitStore) dropDiff(tx *pachsql.Tx, commit *pfs.Commit) error {
	diffIDs, err := getDiff(tx, commit)
	if err != nil {
		return err
	}
	for _, diffID := range diffIDs {
		trackID := commitDiffTrackerID(commit, diffID)
		if err := cs.tr.DeleteTx(tx, trackID); err != nil {
			return errors.EnsureStack(err)
		}
	}
	if _, err := tx.Exec(`DELETE FROM pfs.commit_diffs WHERE commit_id = $1`, pfsdb.CommitKey(commit)); err != nil {
		return errors.EnsureStack(err)
	}
	return nil
}

func getDiff(tx *pachsql.Tx, commit *pfs.Commit) ([]fileset.ID, error) {
	var ids []fileset.ID
	if err := tx.Select(&ids,
		`SELECT fileset_id FROM pfs.commit_diffs
		WHERE commit_id = $1
		ORDER BY num
		`, pfsdb.CommitKey(commit)); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return ids, nil
}

func getTotal(tx *pachsql.Tx, commit *pfs.Commit) (*fileset.ID, error) {
	var id fileset.ID
	if err := tx.Get(&id,
		`SELECT fileset_id FROM pfs.commit_totals
		WHERE commit_id = $1
	`, pfsdb.CommitKey(commit)); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &id, nil
}

func dropTotal(tx *pachsql.Tx, tr track.Tracker, commit *pfs.Commit) error {
	id, err := getTotal(tx, commit)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	trackID := commitTotalTrackerID(commit, *id)
	if err := tr.DeleteTx(tx, trackID); err != nil {
		return errors.EnsureStack(err)
	}
	_, err = tx.Exec(`DELETE FROM pfs.commit_totals WHERE commit_id = $1`, pfsdb.CommitKey(commit))
	return errors.EnsureStack(err)
}

func setTotal(tx *pachsql.Tx, tr track.Tracker, commit *pfs.Commit, id fileset.ID) error {
	oid := commitTotalTrackerID(commit, id)
	pointsTo := []string{id.TrackerID()}
	if err := tr.CreateTx(tx, oid, pointsTo, track.NoTTL); err != nil {
		return errors.EnsureStack(err)
	}
	_, err := tx.Exec(`INSERT INTO pfs.commit_totals (commit_id, fileset_id)
	VALUES ($1, $2)
	ON CONFLICT (commit_id) DO UPDATE
	SET fileset_id = $2
	WHERE commit_totals.commit_id = $1
	`, pfsdb.CommitKey(commit), id)
	return errors.EnsureStack(err)
}

func setDiff(tx *pachsql.Tx, tr track.Tracker, commit *pfs.Commit, id fileset.ID) error {
	oid := commitDiffTrackerID(commit, id)
	pointsTo := []string{id.TrackerID()}
	if err := tr.CreateTx(tx, oid, pointsTo, track.NoTTL); err != nil {
		return errors.EnsureStack(err)
	}
	_, err := tx.Exec(`INSERT INTO pfs.commit_diffs (commit_id, fileset_id)
	VALUES ($1, $2)
	`, pfsdb.CommitKey(commit), id)
	return errors.EnsureStack(err)
}

func commitDiffTrackerID(commit *pfs.Commit, fs fileset.ID) string {
	return commitTrackerPrefix + pfsdb.CommitKey(commit) + "/diff/" + trimPrefix(fs.TrackerID(), fileset.TrackerPrefix)
}

func commitTotalTrackerID(commit *pfs.Commit, fs fileset.ID) string {
	return commitTrackerPrefix + pfsdb.CommitKey(commit) + "/total/" + trimPrefix(fs.TrackerID(), fileset.TrackerPrefix)
}

func trimPrefix(x, prefix string) string {
	if !strings.HasPrefix(x, prefix) {
		panic(x)
	}
	return strings.TrimPrefix(x, prefix)
}
