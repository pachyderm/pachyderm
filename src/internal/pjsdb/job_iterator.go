package pjsdb

import (
	"context"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
)

// IterateJobsFilter is translated to a 'where' SQL clause by the job iterator.
// keeping this as its own separate struct makes it easy to check to see if a
// filter is defined. It also allows filters to be combined later if needed.
type IterateJobsFilter struct {
	Parent   JobID
	Input    []byte
	Spec     []byte
	SpecHash []byte
	Output   []byte
	Error    string

	// TODO(Fahad): think through how to support querying on times.
	//Queued     time.Time
	//Processing time.Time
	//Done       time.Time
}

func (f IterateJobsFilter) IsEmpty() bool {
	if diff := cmp.Diff(f, IterateJobsFilter{}, cmpopts.EquateEmpty()); diff != "" {
		return false
	}
	return true
}

func (f IterateJobsFilter) apply() (conditions []string, values []any) {
	if f.Parent != 0 {
		conditions = append(conditions, "parent = ?")
		values = append(values, f.Parent)
	}
	if f.Spec != nil {
		conditions = append(conditions, "spec = ?")
		values = append(values, f.Spec)
	}
	if f.Input != nil {
		conditions = append(conditions, "input = ?")
		values = append(values, f.Input)
	}
	if f.SpecHash != nil {
		conditions = append(conditions, "spec_hash = ?")
		values = append(values, f.SpecHash)
	}
	if f.Output != nil {
		conditions = append(conditions, "output = ?")
		values = append(values, f.Output)
	}
	if f.Error != "" {
		conditions = append(conditions, "error = ?")
		values = append(values, f.Error)
	}
	return conditions, values
}

type IterateJobsRequest struct {
	IteratorConfiguration
	Filter IterateJobsFilter
}

// JobsIterator implements stream.Iterator[T] on Job objects following the
// pattern set by the pfsdb package.
// Iterators are the underlying type used by list-style crud operations.
type JobsIterator struct {
	paginator pageIterator[jobRow]
	extCtx    sqlx.ExtContext
}

var _ stream.Iterator[Job] = &JobsIterator{} // catch changes that break the interface.

func (i *JobsIterator) Next(ctx context.Context, dst *Job) error {
	if dst == nil {
		return errors.Errorf("dst jobRow cannot be nil")
	}
	row, err := i.paginator.next(ctx, i.extCtx)
	if err != nil {
		return err
	}
	*dst = row.ToJob()
	return nil
}

func NewJobsIterator(extCtx sqlx.ExtContext, req IterateJobsRequest) *JobsIterator {
	var conditions []string
	var values []any
	query := `SELECT * FROM pjs.jobs`
	conditions, values = req.Filter.apply()
	if len(conditions) > 0 {
		query += "\n" + fmt.Sprintf("WHERE %s", strings.Join(conditions, " AND "))
	}
	query += req.orderBy()
	query = extCtx.Rebind(query)
	if req.PageSize == 0 {
		req.PageSize = defaultPageSize
	}
	return &JobsIterator{
		paginator: newPageIterator[jobRow](query, values, req.StartPage, req.PageSize, 0),
		extCtx:    extCtx,
	}
}

// ForEachJob iterates over each Job 'job' in the jobs table and executes the callback cb(job).
func ForEachJob(ctx context.Context, db *pachsql.DB, req IterateJobsRequest, cb func(job Job) error) error {
	ctx = pctx.Child(ctx, "forEachJob")
	if err := stream.ForEach[Job](ctx, NewJobsIterator(db, req), cb); err != nil {
		return errors.Wrap(err, "for each job")
	}
	return nil
}

// ForEachJobTxByFilter is like ForEachJob but requires a *pachsql.Tx and a IterateJobsFilter.
func ForEachJobTxByFilter(ctx context.Context, tx *pachsql.Tx, req IterateJobsRequest, cb func(job Job) error) error {
	ctx = pctx.Child(ctx, "forEachJobTxByFilter")
	if req.Filter.IsEmpty() {
		return errors.Errorf("filter cannot be empty")
	}
	if err := stream.ForEach[Job](ctx, NewJobsIterator(tx, req), func(job Job) error {
		return cb(job)
	}); err != nil {
		return errors.Wrap(err, "for each job tx by filter")
	}
	return nil
}
