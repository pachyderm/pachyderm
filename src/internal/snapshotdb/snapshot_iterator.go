package snapshotdb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
)

type IterateSnapshotsFilter struct {
	Operation    filterOperation
	CreatedAfter time.Time
}

func (f IterateSnapshotsFilter) apply() (where string, values []any) {
	var conditions []string
	if !f.CreatedAfter.IsZero() {
		conditions = append(conditions, "recovery.snapshots.created_at >= ?")
		values = append(values, f.CreatedAfter)
	}
	if len(conditions) == 0 {
		return "", nil
	}
	if f.Operation == "" {
		f.Operation = FilterOperationAND
	}
	where = fmt.Sprintf("\nWHERE (%s)", strings.Join(conditions, " "+string(f.Operation)+" "))
	return where, values
}

const (
	FilterOperationAND = filterOperation("AND")
	FilterOperationOR  = filterOperation("OR")
)

type filterOperation string

type IterateSnapshotsRequest struct {
	IteratorConfiguration
	Filter IterateSnapshotsFilter
}

type SnapshotsIterator struct {
	paginator pageIterator[snapshotRecord]
	extCtx    sqlx.ExtContext
}

var _ stream.Iterator[Snapshot] = &SnapshotsIterator{} // catch changes that break the interface.

func (i *SnapshotsIterator) Next(ctx context.Context, dst *Snapshot) error {
	if dst == nil {
		return errors.Errorf("dst snapshot row cannot be nil")
	}
	record, err := i.paginator.next(ctx, i.extCtx)
	if err != nil {
		return err
	}
	s, err := record.toSnapshot()
	if err != nil {
		return err
	}
	*dst = s
	return nil
}

func NewSnapshotsIterator(extCtx sqlx.ExtContext, req IterateSnapshotsRequest) *SnapshotsIterator {
	var values []any
	query := selectSnapshotPrefix
	where, values := req.Filter.apply()
	query += where
	query += req.orderBy()
	query = extCtx.Rebind(query)
	if req.PageSize == 0 {
		req.PageSize = defaultPageSize
	}
	return &SnapshotsIterator{
		paginator: newPageIterator[snapshotRecord](query, values, req.StartPage, req.PageSize, 0, req.EntryLimit),
		extCtx:    extCtx,
	}
}

func ForEachSnapshot(ctx context.Context, db *pachsql.DB, req IterateSnapshotsRequest, cb func(job Snapshot) error) error {
	ctx = pctx.Child(ctx, "forEachSnapshot")
	if err := stream.ForEach[Snapshot](ctx, NewSnapshotsIterator(db, req), cb); err != nil {
		return errors.Wrap(err, "for each snapshot")
	}
	return nil
}

func ForEachSnapshotTxByFilter(ctx context.Context, tx *pachsql.Tx, req IterateSnapshotsRequest, cb func(job Snapshot) error) error {
	ctx = pctx.Child(ctx, "forEachSnapshotTxByFilter")
	if err := stream.ForEach[Snapshot](ctx, NewSnapshotsIterator(tx, req), func(job Snapshot) error {
		return cb(job)
	}); err != nil {
		return errors.Wrap(err, "for each snapshot tx by filter")
	}
	return nil
}
