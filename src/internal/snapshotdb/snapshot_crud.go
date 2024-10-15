package snapshotdb

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	snapshotpb "github.com/pachyderm/pachyderm/v2/src/snapshot"
	"github.com/pachyderm/pachyderm/v2/src/version"
)

const (
	selectSnapshotPrefix = `SELECT * FROM recovery.snapshots`
	insertSnapshot       = `
		insert into recovery.snapshots (chunkset_id, pachyderm_version, metadata) 
		values ($1, $2, $3) returning id`
	selectSnapshots = `select * from recovery.snapshots where created_at > $1 order by created_at desc limit $2`
	defaultLimit    = 10000
)

func CreateSnapshot(ctx context.Context, tx *pachsql.Tx, s *fileset.Storage, metadata map[string]string) (int64, error) {
	chunksetID, err := s.CreateChunkSet(ctx, tx)
	if err != nil {
		return 0, errors.Wrap(err, "create chunkset")
	}
	var id snapshotID
	if err := tx.GetContext(ctx, &id, insertSnapshot,
		chunksetID, version.Version.String(), metadata); err != nil {
		return 0, errors.Wrap(err, "create snapshot row")
	}
	return int64(id), nil
}

func GetSnapshot(ctx context.Context, tx *pachsql.Tx, id int64) (*snapshotpb.SnapshotInfo, error) {
	record := snapshotRecord{}
	if err := sqlx.GetContext(ctx, tx, &record, selectSnapshotPrefix+`
	WHERE recovery.snapshots.id = $1`, snapshotID(id)); err != nil {
		return nil, errors.Wrap(err, "get snapshot row")
	}
	st := record.toSnapshot()
	return st.toSnapshotInfo(), nil
}

func ListSnapshot(ctx context.Context, tx *pachsql.Tx, since time.Time, limit int32) ([]*snapshotpb.SnapshotInfo, error) {
	if limit == 0 {
		limit = defaultLimit
	}
	snapshots := make([]snapshotRecord, 0)
	if err := sqlx.SelectContext(ctx, tx, &snapshots, selectSnapshots, since, limit); err != nil {
		return nil, errors.Wrap(err, "list snapshots")
	}
	var ret []*snapshotpb.SnapshotInfo
	for _, s := range snapshots {
		ret = append(ret, s.toSnapshot().toSnapshotInfo())
	}
	return ret, nil
}
