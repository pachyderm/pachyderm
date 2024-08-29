package pjsdb

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/pjs"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
)

const (
	maxDepth                  = 10_000
	recursiveTraverseChildren = `
    	WITH RECURSIVE children(parent, id) AS (
			-- basecase
			    SELECT parent, id, 1 as depth, ARRAY[id] AS path
				FROM pjs.jobs
				WHERE id = $1
			UNION ALL
			-- recursive case
				SELECT c.id, j.id, c.depth+1, c.path || j.id
				FROM children c, pjs.jobs j
				WHERE j.parent = c.id AND depth <= $2
		)
	`
	selectJobRecordPrefix = `
		SELECT 
			j.*,
			ARRAY_REMOVE(ARRAY_AGG(jf_input.fileset ORDER BY jf_input.array_position), NULL) as "inputs",
			ARRAY_REMOVE(ARRAY_AGG(jf_output.fileset ORDER BY jf_output.array_position), NULL) as "outputs"
		FROM pjs.jobs j
		LEFT JOIN pjs.job_filesets jf_input ON j.id = jf_input.job_id AND jf_input.fileset_type = 'input'
		LEFT JOIN pjs.job_filesets jf_output ON j.id = jf_output.job_id AND jf_output.fileset_type = 'output'
	`
)

// functions in the CRUD API assume that the JobContext token has already been resolved upstream to a job by the
// job system. Some functions take a request object such as an IterateJobsRequest. Requests bundle associated fields
// that can be validated before sql statements are executed.

// CreateJobRequest is a bundle of related fields required for a CreateJob() invocation.
// In pfsdb, we used a pattern of forcing the caller to convert their resources to a _Info struct, but its problematic since
// each database function only really needs a subset of the fields and it is not clear to the caller which fields are required.
type CreateJobRequest struct {
	Parent  JobID
	Inputs  []fileset.PinnedFileset
	Program fileset.PinnedFileset
}

// IsSanitized is a utility function that wraps sanitize() for the purposes of testing.
func (req CreateJobRequest) IsSanitized(ctx context.Context, tx *pachsql.Tx) error {
	if _, err := req.sanitize(ctx, tx); err != nil {
		return err
	}
	return nil
}

func (req CreateJobRequest) sanitize(ctx context.Context, tx *pachsql.Tx) (createJobRequest, error) {
	defaultID := fileset.PinnedFileset{}
	if req.Program == defaultID {
		return createJobRequest{}, errors.New("program cannot be nil")
	}
	sanitizedReq := createJobRequest{
		Program:     []byte(fileset.ID(req.Program).HexString()), // there aren't real pins as of yet.
		ProgramHash: []byte(fileset.ID(req.Program).HexString()), // eventually the ID will be a hash.
	}
	// validate parent.
	sanitizedReq.Parent = sql.NullInt64{Valid: false}
	if req.Parent != 0 {
		sanitizedReq.Parent.Int64 = int64(req.Parent)
		if _, err := GetJob(ctx, tx, req.Parent); err != nil {
			if errors.As(err, &JobNotFoundError{}) {
				return createJobRequest{}, errors.Join(ErrParentNotFound, errors.Wrap(err, "sanitize"))
			}
			return createJobRequest{}, errors.Wrap(err, "sanitize")
		}
		sanitizedReq.Parent.Valid = true
	}
	for i, input := range req.Inputs {
		if input == defaultID {
			continue
		}
		sanitizedReq.Inputs = append(sanitizedReq.Inputs, jobFilesetsRow{
			JobID:         0, // the job id is not known until the job row is inserted in create job.
			Type:          "input",
			ArrayPosition: i,
			Fileset:       []byte(fileset.ID(input).HexString()),
		})
	}
	return sanitizedReq, nil
}

type createJobRequest struct {
	Parent      sql.NullInt64
	Inputs      []jobFilesetsRow
	Program     []byte
	ProgramHash []byte
}

// CreateJob creates a job entry in postgres.
func CreateJob(ctx context.Context, tx *pachsql.Tx, req CreateJobRequest) (JobID, error) {
	ctx = pctx.Child(ctx, "createJob")
	sReq, err := req.sanitize(ctx, tx)
	if err != nil {
		return 0, errors.Wrap(err, "create job")
	}
	// insert into the jobs table.
	var id JobID
	row := tx.QueryRowxContext(ctx, `
		INSERT INTO pjs.jobs 
		(program, program_hash, parent) 
		VALUES ($1, $2, $3) 
		RETURNING id`, sReq.Program, sReq.ProgramHash, sReq.Parent)
	if err := row.Scan(&id); err != nil {
		return 0, errors.Wrap(err, "create job: inserting row")
	}
	// insert into the jobs_filesets table.
	for _, input := range sReq.Inputs {
		input.JobID = id
		_, err := sqlx.NamedExecContext(ctx, tx, `
		INSERT INTO pjs.job_filesets 
		(job_id, fileset_type, array_position, fileset) 
		VALUES (:job_id, :fileset_type, :array_position, :fileset);`, input)
		if err != nil {
			return 0, errors.Wrap(err, "create job: inserting job_filesets row")
		}
	}
	return id, nil
}

// GetJob returns a job by its JobID. GetJob should not be used to claim a job.
func GetJob(ctx context.Context, tx *pachsql.Tx, id JobID) (Job, error) {
	ctx = pctx.Child(ctx, "getJob")
	record := jobRecord{}
	err := sqlx.GetContext(ctx, tx, &record, selectJobRecordPrefix+`
	WHERE j.id = $1 GROUP BY j.id;`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Job{}, &JobNotFoundError{ID: id}
		}
		return Job{}, errors.Wrap(err, "get job")
	}

	job, err := record.toJob()
	if err != nil {
		return Job{}, errors.Wrap(err, "get job")
	}
	return job, nil
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
		RETURNING id;`, job.ID, maxDepth); err != nil {
		return nil, errors.Wrap(err, "cancel job")
	}
	return ids, nil
}

// WalkAlgorithm is an enumerator for walking algorithms. It reflects pjs.WalkAlgorithm.
// Unfortunately, pjs.WalkAlgorithm cannot be used otherwise it would introduce a circular
// dependency.
type WalkAlgorithm int32

const (
	Unknown WalkAlgorithm = iota
	LevelOrder
	PreOrder
	MirroredPostOrder
)

// WalkJob walks from job 'id' down to all of its children.
func WalkJob(ctx context.Context, tx *pachsql.Tx, id JobID, algo WalkAlgorithm, depth uint64) ([]Job, error) {
	pctx.Child(ctx, "walkJobPJSDB")
	job, err := GetJob(ctx, tx, id)
	if err != nil {
		return nil, errors.Wrap(err, "get job")
	}
	var records []jobRecord
	var walker func(ctx context.Context, tx *pachsql.Tx, id JobID, maxDepth uint64) ([]jobRecord, error)
	walkerName := ""
	//exhaustive:enforce
	switch algo {
	case LevelOrder:
		walker = walkLevelOrder
		walkerName = "levelOrder"
	case PreOrder:
		walker = walkPreOrder
		walkerName = "preOrder"
	case MirroredPostOrder:
		walker = walkMirroredPostOrder
		walkerName = "mirroredPostOrder"
	default:
		return nil, errors.New("unknown walk algorithm is provided")
	}
	if depth == 0 || depth > 10_000 {
		depth = maxDepth
	}
	records, err = walker(ctx, tx, job.ID, depth)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("walker (%s)", walkerName))
	}
	jobs := make([]Job, 0)
	for i, record := range records {
		job, err := record.toJob()
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("to job, iteration=%d/%d", i, len(records)))
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func walkLevelOrder(ctx context.Context, tx *pachsql.Tx, id JobID, depth uint64) (records []jobRecord, err error) {
	if err = sqlx.SelectContext(ctx, tx, &records, recursiveTraverseChildren+selectJobRecordPrefix+`
		INNER JOIN children c ON j.id = c.id
		GROUP BY j.id
		ORDER BY MIN(depth) ASC;
	`, id, depth); err != nil {
		return nil, errors.Wrap(err, "select context")
	}
	return records, nil
}

func walkPreOrder(ctx context.Context, tx *pachsql.Tx, id JobID, depth uint64) (records []jobRecord, err error) {
	if err = sqlx.SelectContext(ctx, tx, &records, recursiveTraverseChildren+selectJobRecordPrefix+`
		INNER JOIN children c ON j.id = c.id
		GROUP BY j.id, c.path
		ORDER BY c.path;
	`, id, depth); err != nil {
		return nil, errors.Wrap(err, "select context")
	}
	return records, nil
}

func walkMirroredPostOrder(ctx context.Context, tx *pachsql.Tx, id JobID, depth uint64) (records []jobRecord, err error) {
	if err = sqlx.SelectContext(ctx, tx, &records,
		recursiveTraverseChildren+`,
		post_order AS (
			SELECT id, ROW_NUMBER() OVER (ORDER BY path DESC) AS post_order
			FROM children
		)
		`+selectJobRecordPrefix+`
		JOIN post_order p ON p.id = j.id 
		GROUP BY j.id, p.post_order
		ORDER BY p.post_order;
	`, id, depth); err != nil {
		return nil, errors.Wrap(err, "select context")
	}
	return records, nil
}

// ListJobTxByFilter returns a list of Job objects matching the filter criteria in req.Filter.
// req.Filter must not be nil.
func ListJobTxByFilter(ctx context.Context, tx *pachsql.Tx, req IterateJobsRequest) ([]Job, error) {
	ctx = pctx.Child(ctx, "listJobTxByFilter")
	var jobs []Job
	if err := ForEachJobTxByFilter(ctx, tx, req, func(job Job) error {
		j := job
		jobs = append(jobs, j)
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
	RETURNING id;`, job.ID, maxDepth); err != nil {
		return nil, errors.Wrap(err, "cancel job")
	}
	return ids, nil
}

// ErrorJob is called when job processing has an error. It updates job err code and
// done timestamp in database.
func ErrorJob(ctx context.Context, tx *pachsql.Tx, jobID JobID, errCode pjs.JobErrorCode) error {
	ctx = pctx.Child(ctx, "complete error")
	_, err := tx.ExecContext(ctx, `
		UPDATE pjs.jobs
		SET done = CURRENT_TIMESTAMP, error = $1
		WHERE id = $2 AND error IS NULL
	`, errCode, jobID)
	if err != nil {
		return errors.Wrapf(err, "error job: update error and state to done")
	}
	return nil
}

// CompleteJob is called when job processing without any error. It updates done timestamp
// output filesets and in database.
func CompleteJob(ctx context.Context, tx *pachsql.Tx, jobID JobID, outputs []string) error {
	ctx = pctx.Child(ctx, "complete ok")
	result, err := tx.ExecContext(ctx, `
		UPDATE pjs.jobs
		SET done = CURRENT_TIMESTAMP
		WHERE id = $1 AND error IS NULL
	`, jobID)
	if err != nil {
		return errors.Wrapf(err, "complete job: update state to done")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "complete job: get rows affected")
	}
	// if no rows are affected,
	// 1) the job's error is not null(it can be cancelled), so we should not update output
	// 2) the id does not exist which means the job has been deleted. we should not update
	// output for a deleted job
	if rowsAffected == 0 {
		return nil
	}
	for pos, output := range outputs {
		_, err := tx.ExecContext(ctx, `
		INSERT INTO pjs.job_filesets 
		(job_id, fileset_type, array_position, fileset) 
		VALUES ($1, $2, $3, $4);`, jobID, "output", pos, []byte(output))
		if err != nil {
			return errors.Wrapf(err, "complete ok: insert output fileset")
		}
	}
	return nil
}

// validateJobTree walks jobs from job 'id' and confirms that no parent job with processing or queued child jobs is done.
func validateJobTree(ctx context.Context, tx *pachsql.Tx, id JobID) error {
	ctx = pctx.Child(ctx, "validateJobTree")
	job, err := GetJob(ctx, tx, id)
	if err != nil {
		return errors.Wrap(err, "validateJobTree")
	}
	rows := make([]jobRow, 0)
	if err = sqlx.SelectContext(ctx, tx, &rows, recursiveTraverseChildren+`
		SELECT j.id, j.parent FROM pjs.jobs j
		INNER JOIN children c ON j.id = c.id
		INNER JOIN pjs.jobs p ON j.parent = p.id
		WHERE p.done IS NOT NULL AND j.done IS NULL;`, job.ID, maxDepth); err != nil {
		return errors.Wrap(err, "validateJobTree")
	}
	errs := make([]error, 0)
	for _, row := range rows {
		errs = append(errs, errors.New(fmt.Sprint(row.Parent.Int64, " is done before child job ", row.ID)))
	}
	return errors.Join(errs...)
}
