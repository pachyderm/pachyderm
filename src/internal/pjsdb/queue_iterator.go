package pjsdb

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
)

// TODO(Fahad): add queue filter once filter is designed.
type IterateQueuesRequest struct {
	IteratorConfiguration
}

// QueuesIterator implements stream.Iterator[T] on Queue objects following the
// pattern set by the pfsdb package.
// Iterators are the underlying type used by list-style crud operations.
type QueuesIterator struct {
	paginator pageIterator[queueRecord]
	extCtx    sqlx.ExtContext
}

var _ stream.Iterator[Queue] = &QueuesIterator{} // catches changes that break the interface.

func (i *QueuesIterator) Next(ctx context.Context, dst *Queue) error {
	if dst == nil {
		return errors.Errorf("dst queue cannot be nil")
	}
	queueRecord, err := i.paginator.next(ctx, i.extCtx)
	if err != nil {
		return err
	}
	queue, err := queueRecord.toQueue()
	if err != nil {
		return errors.Wrap(err, "next")
	}
	*dst = queue
	return nil
}

func NewQueuesIterator(extCtx sqlx.ExtContext, req IterateQueuesRequest) *QueuesIterator {
	var values []any
	// The current storage system supports cloned filesets with the same content hash.
	// therefore, programs must be aggregated.
	query := `WITH queues AS (SELECT DISTINCT program_hash FROM pjs.jobs)
			  SELECT 
				queues.program_hash AS "id", 
				array_agg(j.id ORDER BY j.id ASC) AS "jobs",
				array_agg(j.program ORDER BY j.id ASC) AS "programs",
				count(j.id) AS "size"
			  FROM queues JOIN pjs.jobs j ON j.program_hash = queues.program_hash
			  GROUP BY queues.program_hash`
	query = extCtx.Rebind(query)
	if req.PageSize == 0 {
		req.PageSize = defaultPageSize
	}
	return &QueuesIterator{
		paginator: newPageIterator[queueRecord](query, values, req.StartPage, req.PageSize, 0),
		extCtx:    extCtx,
	}
}

// ForEachQueue calculates and iterates over each Queue 'queue' in the pfs.jobs table and executes the callback cb(queue).
func ForEachQueue(ctx context.Context, db *pachsql.DB, req IterateQueuesRequest, cb func(queue Queue) error) error {
	ctx = pctx.Child(ctx, "forEachQueue")
	if err := stream.ForEach[Queue](ctx, NewQueuesIterator(db, req), cb); err != nil {
		return errors.Wrap(err, "for each queue")
	}
	return nil
}
