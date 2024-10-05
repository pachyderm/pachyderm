package snapshot

//import (
//	"context"
//	"testing"
//
//	"github.com/jmoiron/sqlx"
//	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
//	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
//	"github.com/pachyderm/pachyderm/v2/src/version"
//)
//
//type SnapshotID int64
//
//func createSnapshotRow(ctx context.Context, tx *sqlx.Tx, s *fileset.Storage) (result SnapshotID, _ error) {
//	chunksetID, err := s.CreateChunkSet(ctx, tx)
//	if err != nil {
//		return 0, errors.Wrap(err, "create chunkset")
//	}
//	if err := tx.GetContext(ctx, &result, `insert into recovery.snapshots (chunkset_id, pachyderm_version) values ($1, $2) returning id`, chunksetID, version.Version.String()); err != nil {
//		return 0, errors.Wrap(err, "create snapshot row")
//	}
//	return result, nil
//}
//
//func TestListSnapshot(t *testing.T) {
//
//}
