package snapshotdb

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/snapshot"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/version"
)

const (
	selectSnapshotPrefix = `SELECT * FROM recovery.snapshots`
	insertSnapshot       = `
		insert into recovery.snapshots (chunkset_id, pachyderm_version, metadata) 
		values ($1, $2, $3) returning id`
	insertSnapshotWithoutMetadata = `
		insert into recovery.snapshots (chunkset_id, pachyderm_version) 
		values ($1, $2) returning id`
)

func CreateSnapshot(ctx context.Context, tx *pachsql.Tx, s *fileset.Storage, metadata []byte) (snapshot.SnapshotID, error) {
	chunksetID, err := s.CreateChunkSet(ctx, tx)
	if err != nil {
		return 0, errors.Wrap(err, "create chunkset")
	}
	query := insertSnapshot
	if len(metadata) == 0 {
		query = insertSnapshotWithoutMetadata
	}
	var id snapshot.SnapshotID
	if err := tx.GetContext(ctx, &id, query,
		chunksetID, version.Version.String()); err != nil {
		return 0, errors.Wrap(err, "create snapshot row")
	}
	return id, nil
}

func GetSnapshot(ctx context.Context, tx *pachsql.Tx, id snapshot.SnapshotID) (snapshot.Snapshot, error) {
	record := snapshotRecord{}
	if err := sqlx.GetContext(ctx, tx, &record, selectSnapshotPrefix+`
	WHERE recovery.snapshots.id = $1`, id); err != nil {
		return snapshot.Snapshot{}, errors.Wrap(err, "get snapshot row")
	}
	st, err := record.toSnapshot()
	if err != nil {
		return snapshot.Snapshot{}, errors.Wrap(err, "create snapshot row")
	}
	return st, nil
}

func ListSnapshotTxByFilter(ctx context.Context, tx *pachsql.Tx, req IterateSnapshotsRequest) ([]snapshot.Snapshot, error) {
	ctx = pctx.Child(ctx, "listSnapshotTxByFilter")
	var snapshots []snapshot.Snapshot
	if err := ForEachSnapshotTxByFilter(ctx, tx, req, func(s snapshot.Snapshot) error {
		s_ := s
		snapshots = append(snapshots, s_)
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "list snapshots tx by filter")
	}
	return snapshots, nil
}
