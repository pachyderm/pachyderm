package snapshot

//import (
//	"context"
//	"fmt"
//	"github.com/pachyderm/pachyderm/v2/src/internal/snapshotdb"
//
//	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
//	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
//	"github.com/pachyderm/pachyderm/v2/src/internal/log"
//	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
//	"github.com/pachyderm/pachyderm/v2/src/snapshot"
//)
//
//type apiServer struct {
//	snapshot.UnimplementedAPIServer
//	env Env
//}
//
//func (a *apiServer) ListSnapshot(req *snapshot.ListSnapshotRequest, srv snapshot.API_ListSnapshotServer) (err error) {
//	ctx, done := log.SpanContext(srv.Context(), "list snapshot")
//	defer done(log.Errorp(&err))
//
//	var snapshots []Snapshot
//	if err := dbutil.WithTx(ctx, a.env.DB, func(ctx context.Context, sqlTx *pachsql.Tx) error {
//		snapshots, err = snapshotdb.ListSnapshotTxByFilter(ctx, sqlTx, snapshotdb.IterateSnapshotsRequest{})
//		return errors.Wrap(err, "list snapshot in snapshotdb")
//	}, dbutil.WithReadOnly()); err != nil {
//		return errors.Wrap(err, "with tx")
//	}
//
//	for i, s := range snapshots {
//		resp := &snapshot.ListSnapshotResponse{}
//
//		if err := srv.Send(resp); err != nil {
//			return errors.Wrap(err, fmt.Sprintf("send, iteration=%d/%d", i, len(snapshots)))
//		}
//	}
//	return nil
//}
