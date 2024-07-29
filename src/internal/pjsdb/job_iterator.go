package pjsdb

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jmoiron/sqlx"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
)

// IterateJobsFilter is translated to a 'where' SQL clause for use by the job iterator.
// keeping this as its own separate struct makes it easy to check to see if a
// filter is defined. It also allows filters to be combined later if needed.
// TODO(Fahad): combine filters by adding filter.And(other Filter) and filter.Or(other Filter)
type IterateJobsFilter struct {
	// Operation determines how fields are joined into a predicate. The default is 'AND'
	Operation filterOperation

	// fields to filter on
	Parent      JobID
	HasInput    []byte
	Program     []byte
	ProgramHash []byte
	HasOutput   []byte
	Error       string
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

func (f IterateJobsFilter) apply() (where string, values []any) {
	var conditions []string
	// from pjs.jobs
	if f.Parent != 0 {
		conditions = append(conditions, "j.parent = ?")
		values = append(values, f.Parent)
	}
	if f.Program != nil {
		conditions = append(conditions, "j.program = ?")
		values = append(values, f.Program)
	}
	if f.ProgramHash != nil {
		conditions = append(conditions, "j.program_hash = ?")
		values = append(values, f.ProgramHash)
	}
	if f.Error != "" {
		conditions = append(conditions, "j.error = ?")
		values = append(values, f.Error)
	}
	// from pjs.job_filesets
	if f.HasInput != nil {
		conditions = append(conditions, "jf_input.fileset IN (?)")
		values = append(values, f.HasInput)
	}
	if f.HasOutput != nil {
		conditions = append(conditions, "jf_output.fileset IN (?)")
		values = append(values, f.HasOutput)
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

type IterateJobsRequest struct {
	IteratorConfiguration
	Filter IterateJobsFilter
}

// JobsIterator implements stream.Iterator[T] on Job objects following the
// pattern set by the pfsdb package.
// Iterators are the underlying type used by list-style crud operations.
type JobsIterator struct {
	paginator pageIterator[jobRecord]
	extCtx    sqlx.ExtContext
}

var _ stream.Iterator[Job] = &JobsIterator{} // catch changes that break the interface.

func (i *JobsIterator) Next(ctx context.Context, dst *Job) error {
	if dst == nil {
		return errors.Errorf("dst jobRow cannot be nil")
	}
	record, err := i.paginator.next(ctx, i.extCtx)
	if err != nil {
		return err
	}
	job, err := record.toJob()
	if err != nil {
		return err
	}
	*dst = job
	return nil
}

func NewJobsIterator(extCtx sqlx.ExtContext, req IterateJobsRequest) *JobsIterator {
	var values []any
	query := selectJobRecordPrefix
	where, values := req.Filter.apply()
	query += where
	query += "\nGROUP BY j.id "
	query += req.orderBy()
	query = extCtx.Rebind(query)
	if req.PageSize == 0 {
		req.PageSize = defaultPageSize
	}
	return &JobsIterator{
		paginator: newPageIterator[jobRecord](query, values, req.StartPage, req.PageSize, 0),
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
