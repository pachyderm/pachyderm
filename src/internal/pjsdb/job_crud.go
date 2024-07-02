package pjsdb

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
)

const (
	recursiveTraverseChildren = `
      WITH RECURSIVE children(parent, id) AS (
	    -- basecase
	        SELECT parent, id, 1 as "depth"
	        FROM pjs.jobs
	        WHERE id = $1
	    UNION ALL
	    -- recursive case
	        SELECT c.id, j.id, c.depth+1
	        FROM children c, pjs.jobs j
	        WHERE j.parent = c.id AND depth < 10000
	)
	`
)

// functions in the CRUD API assume that the JobContext token has already been resolved upstream to a job by the
// job system. Some functions take a request object such as an IterateJobsRequest. Requests bundle associated fields
// that can be validated before sql statements are executed.

// CreateJob creates a job entry in postgres.
func CreateJob(ctx context.Context, tx *pachsql.Tx, req createJobRequest) (JobID, error) {
	ctx = pctx.Child(ctx, "createJob")
	var id JobID
	row := tx.QueryRowxContext(ctx, `
		INSERT INTO pjs.jobs (input, spec, spec_hash, parent) VALUES 
		($1, $2, $3, $4) RETURNING id`, req.input, req.spec, req.specHash, req.parent)
	if err := row.Scan(&id); err != nil {
		return 0, errors.Wrap(err, "create job: inserting row")
	}
	return id, nil
}

// GetJob returns a job by its JobID. GetJob should not be used to claim a job.
func GetJob(ctx context.Context, tx *pachsql.Tx, id JobID) (*Job, error) {
	ctx = pctx.Child(ctx, "getJob")
	rows := []JobRow{{}}
	err := sqlx.SelectContext(ctx, tx, &rows, `SELECT * FROM pjs.jobs WHERE id = $1`, id)
	if err != nil {
		return nil, errors.Wrap(err, "get job")
	}
	if len(rows) == 0 {
		return nil, &JobNotFoundError{ID: id}
	}
	// since id is the primary key, it should be impossible for more than one row to be returned.
	return &Job{JobRow: rows[0]}, nil
}

// CancelJob cancels job with ID 'id' and all child jobs of 'id'.
func CancelJob(ctx context.Context, tx *pachsql.Tx, id JobID) ([]JobID, error) {
	ctx = pctx.Child(ctx, "cancelJob")
	job, err := GetJob(ctx, tx, id)
	if err != nil {
		return nil, errors.Wrap(err, "cancel job")
	}
	ids := make([]JobID, 0)
	// TODO(Fahad): Should the recursion happen here? Or should the CancelJob() RPC do it?
	// CancelJob skips jobs that have already completed.
	if err = sqlx.SelectContext(ctx, tx, &ids, recursiveTraverseChildren+`
	UPDATE pjs.jobs SET done = CURRENT_TIMESTAMP, error = 'canceled' 
	WHERE id IN (select id FROM children) AND done IS NULL 
	RETURNING id;`, job.ID); err != nil {
		return nil, errors.Wrap(err, "cancel job")
	}
	return ids, nil
}

// WalkJob walks from job 'id' down to all of its children.
func WalkJob(ctx context.Context, tx *pachsql.Tx, id JobID) ([]*Job, error) {
	pctx.Child(ctx, "walkJob")
	job, err := GetJob(ctx, tx, id)
	if err != nil {
		return nil, errors.Wrap(err, "walk job")
	}
	rows := make([]JobRow, 0)
	if err = sqlx.SelectContext(ctx, tx, &rows, recursiveTraverseChildren+`
    SELECT j.* FROM pjs.jobs j
	INNER JOIN children c ON j.id = c.id
    GROUP BY j.id, j.parent, j.spec, j.input, j.spec_hash, j.output, j.error
	ORDER BY MIN(depth) ASC;
	`, job.ID); err != nil {
		return nil, errors.Wrap(err, "walk job")
	}
	jobs := make([]*Job, 0)
	for _, row := range rows {
		jobs = append(jobs, &Job{row})
	}
	return jobs, nil
}

// ListJobTxByFilter returns a list of Job objects matching the filter criteria in req.Filter.
// req.Filter must not be nil.
func ListJobTxByFilter(ctx context.Context, tx *pachsql.Tx, req IterateJobsRequest) ([]*Job, error) {
	ctx = pctx.Child(ctx, "listJobTxByFilter")
	var jobs []*Job
	if err := ForEachJobTxByFilter(ctx, tx, req, func(job Job) error {
		j := job
		jobs = append(jobs, &j)
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "list jobs tx by filter")
	}
	return jobs, nil
}

// DeleteJob deletes a job and its child jobs from the jobs table. It returns a list of jobs that were deleted.
// A job may only be deleted once it is done. This usually happens through cancellation.
func DeleteJob(ctx context.Context, tx *pachsql.Tx, id JobID) ([]JobID, error) {
	ctx = pctx.Child(ctx, "deleteJob")
	job, err := GetJob(ctx, tx, id)
	if err != nil {
		return nil, errors.Wrap(err, "delete job")
	}
	if err := validateJobTree(ctx, tx, id); err != nil {
		return nil, errors.Wrap(err, "delete job")
	}
	ids := make([]JobID, 0)
	if err = sqlx.SelectContext(ctx, tx, &ids, recursiveTraverseChildren+`
	DELETE FROM pjs.jobs WHERE id IN (SELECT id FROM children) AND done IS NOT NULL
	RETURNING id;`, job.ID); err != nil {
		return nil, errors.Wrap(err, "cancel job")
	}
	return ids, nil
}

// validateJobTree walks jobs from job 'id' and confirms that no parent job with processing or queued child jobs is done.
func validateJobTree(ctx context.Context, tx *pachsql.Tx, id JobID) error {
	ctx = pctx.Child(ctx, "validateJobTree")
	job, err := GetJob(ctx, tx, id)
	if err != nil {
		return errors.Wrap(err, "validateJobTree")
	}
	rows := make([]JobRow, 0)
	if err = sqlx.SelectContext(ctx, tx, &rows, recursiveTraverseChildren+`
		SELECT j.id, j.parent FROM pjs.jobs j
		INNER JOIN children c ON j.id = c.id
		INNER JOIN pjs.jobs p ON j.parent = p.id
		WHERE p.done IS NOT NULL AND j.done IS NULL;`, job.ID); err != nil {
		return errors.Wrap(err, "validateJobTree")
	}
	errs := make([]error, 0)
	for _, row := range rows {
		errs = append(errs, errors.New(fmt.Sprint(row.Parent.Int64, " is done before child job ", row.ID)))
	}
	return errors.Join(errs...)
}
