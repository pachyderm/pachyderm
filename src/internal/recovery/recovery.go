// Package recovery implements subsystem-independent disaster recovery.
package recovery

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/version"
)

type SnapshotID int64

// String implements fmt.Stringer.
func (id SnapshotID) String() string {
	var invalid string
	if id < 1 {
		// a postgres bigserial is 1-2^63 stored in a signed int64.
		invalid = "Invalid"
	}
	return fmt.Sprintf("%vSnapshotID(%v)", invalid, int64(id))
}

func createSnapshotRow(ctx context.Context, tx *sqlx.Tx, s *fileset.Storage) (result SnapshotID, _ error) {
	chunksetID, err := s.CreateChunkSet(ctx, tx)
	if err != nil {
		return 0, errors.Wrap(err, "create chunkset")
	}
	if err := tx.GetContext(ctx, &result, `insert into recovery.snapshots (chunkset_id, pachyderm_version) values ($1, $2) returning id`, chunksetID, version.Version.String()); err != nil {
		return 0, errors.Wrap(err, "create snapshot row")
	}
	return result, nil
}

func addDatabaseDump(ctx context.Context, tx *sqlx.Tx, snapshotID SnapshotID, filesetID fileset.ID) error {
	result, err := tx.ExecContext(ctx, `update recovery.snapshots set sql_dump_fileset_id=$1 where id=$2`, filesetID[:], snapshotID)
	if err != nil {
		return errors.Wrap(err, "update snapshot to contain database dump fileset")
	}
	got, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "get affected row count")
	}
	if got != 1 {
		return errors.Errorf("rows affected by snapshot update: got %v want 1", got)
	}
	return nil
}
