// Package snapshot implements subsystem-independent disaster recovery.
package snapshot

import (
	"context"
	"fmt"

	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pgjsontypes"
	"github.com/pachyderm/pachyderm/v2/src/internal/snapshotdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	snapshotpb "github.com/pachyderm/pachyderm/v2/src/snapshot"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type APIServer struct {
	snapshotpb.UnimplementedAPIServer
	DB    *pachsql.DB
	Store *fileset.Storage
}

var _ snapshotpb.APIServer = &APIServer{}

func (a *APIServer) CreateSnapshot(ctx context.Context, request *snapshotpb.CreateSnapshotRequest) (*snapshotpb.CreateSnapshotResponse, error) {
	var ret snapshotpb.CreateSnapshotResponse
	s := &Snapshotter{
		DB:      a.DB,
		Storage: a.Store,
	}
	id, err := s.CreateSnapshot(ctx, CreateSnapshotOptions{
		Metadata: pgjsontypes.StringMap{
			Data: request.Metadata,
		},
	})
	ret.Id = int64(id)
	if err != nil {
		return &ret, errors.Wrap(err, "create snapshot")
	}
	return &ret, nil
}

func (a *APIServer) InspectSnapshot(ctx context.Context, req *snapshotpb.InspectSnapshotRequest) (*snapshotpb.InspectSnapshotResponse, error) {
	var ret *snapshotpb.InspectSnapshotResponse
	if err := dbutil.WithTx(ctx, a.DB, func(ctx context.Context, sqlTx *pachsql.Tx) error {
		info, err := snapshotdb.GetSnapshot(ctx, sqlTx, req.Id)
		if err != nil {
			if errors.As(err, &snapshotdb.SnapshotNotFoundError{}) {
				return status.Errorf(codes.NotFound, "snapshot %d not found", req.Id)
			}
			return errors.Wrap(err, "get snapshot from db")
		}
		ret = &snapshotpb.InspectSnapshotResponse{Info: info}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "with Tx")
	}
	return ret, nil
}

func (a *APIServer) ListSnapshot(req *snapshotpb.ListSnapshotRequest, srv snapshotpb.API_ListSnapshotServer) (err error) {
	ctx, done := log.SpanContext(srv.Context(), "list snapshotpb")
	defer done(log.Errorp(&err))

	var snapshots []*snapshotpb.SnapshotInfo
	if err := dbutil.WithTx(ctx, a.DB, func(ctx context.Context, sqlTx *pachsql.Tx) error {
		snapshots, err = snapshotdb.ListSnapshot(ctx, sqlTx, req.Since.AsTime(), req.Limit)
		return errors.Wrap(err, "list snapshot in snapshotdb")
	}, dbutil.WithReadOnly()); err != nil {
		return errors.Wrap(err, "with tx")
	}

	for i, s := range snapshots {
		resp := &snapshotpb.ListSnapshotResponse{
			Info: s,
		}
		if err := srv.Send(resp); err != nil {
			return errors.Wrap(err, fmt.Sprintf("send, iteration=%d/%d", i, len(snapshots)))
		}
	}
	return nil
}

func (a *APIServer) DeleteSnapshot(ctx context.Context, req *snapshotpb.DeleteSnapshotRequest) (*snapshotpb.DeleteSnapshotResponse, error) {
	var ret *snapshotpb.DeleteSnapshotResponse
	if err := dbutil.WithTx(ctx, a.DB, func(ctx context.Context, sqlTx *pachsql.Tx) error {
		err := snapshotdb.DeleteSnapshot(ctx, sqlTx, req.Id)
		if err != nil {
			return errors.Wrap(err, "delete snapshot in db")
		}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "with Tx")
	}
	return ret, nil
}
