// Package snapshot implements subsystem-independent disaster recovery.
package snapshot

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/snapshotdb"
	"github.com/pachyderm/pachyderm/v2/src/snapshot"
)

type apiServer struct {
	snapshot.UnimplementedAPIServer
	env Env
}

func (a *apiServer) CreateSnapshot(ctx context.Context, request *snapshot.CreateSnapshotRequest) (*snapshot.CreateSnapshotResponse, error) {
	jsonData, err := json.Marshal(request.Metadata)
	if err != nil {
		return nil, errors.Wrap(err, "convert metadata to json")
	}
	var ret snapshot.CreateSnapshotResponse
	if err := dbutil.WithTx(ctx, a.env.DB, func(ctx context.Context, sqlTx *pachsql.Tx) error {
		snapshotID, err := snapshotdb.CreateSnapshot(ctx, sqlTx, a.env.store, jsonData)
		if err != nil {
			return errors.Wrap(err, "create job")
		}
		ret.Id = int64(snapshotID)
		return nil
	}); err != nil {
		return nil, err
	}
	return &ret, nil
}

func (a *apiServer) InspectSnapshot(ctx context.Context, req *snapshot.InspectSnapshotRequest) (*snapshot.InspectSnapshotResponse, error) {
	return nil, nil
}

func (a *apiServer) ListSnapshot(req *snapshot.ListSnapshotRequest, srv snapshot.API_ListSnapshotServer) (err error) {
	ctx, done := log.SpanContext(srv.Context(), "list snapshot")
	defer done(log.Errorp(&err))

	var snapshots []snapshotdb.Snapshot
	if err := dbutil.WithTx(ctx, a.env.DB, func(ctx context.Context, sqlTx *pachsql.Tx) error {
		iterReq := snapshotdb.IterateSnapshotsRequest{
			Filter: snapshotdb.IterateSnapshotsFilter{
				CreatedAfter: req.Since.AsTime(),
			},
		}
		iterReq.EntryLimit = uint64(req.Limit)
		snapshots, err = snapshotdb.ListSnapshotTxByFilter(ctx, sqlTx, iterReq)
		return errors.Wrap(err, "list snapshot in snapshotdb")
	}, dbutil.WithReadOnly()); err != nil {
		return errors.Wrap(err, "with tx")
	}

	for i, s := range snapshots {
		info := s.ToSnapshotInfo()
		resp := &snapshot.ListSnapshotResponse{
			Info: info,
		}
		if err := srv.Send(resp); err != nil {
			return errors.Wrap(err, fmt.Sprintf("send, iteration=%d/%d", i, len(snapshots)))
		}
	}
	return nil
}
