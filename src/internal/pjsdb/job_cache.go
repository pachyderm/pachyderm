package pjsdb

// job_cache.go defines functions for interacting with the server-side job cache.

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachhash"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"go.uber.org/zap"
)

func jobCacheKey(programHash []byte, inputHashes [][]byte) []byte {
	hasher := pachhash.New()
	hasher.Write(programHash)
	for _, input := range inputHashes {
		hasher.Write(input)
	}
	return hasher.Sum(nil)
}

func readFromJobCache(ctx context.Context, extCtx sqlx.ExtContext, jobHash []byte) (job Job, err error) {
	if len(jobHash) == 0 {
		return Job{}, errors.Wrap(err, "no job hash supplied.")
	}
	ctx = pctx.Child(ctx, "readFromJobCache")
	record := jobRecord{}
	err = sqlx.GetContext(ctx, extCtx, &record, selectJobRecordPrefix+`
	WHERE jc.job_hash = $1 GROUP BY j.id, jc.job_hash, jc.cache_read, jc.cache_write;`, jobHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Job{}, &JobCacheCacheMissError{JobHash: string(jobHash)}
		}
		return Job{}, errors.Wrap(err, "get context")
	}
	log.Debug(ctx, "found job in the pjs job cache", zap.Uint64("job", uint64(record.ID)))
	job, err = record.toJob()
	if err != nil {
		return Job{}, errors.Wrap(err, "to job")
	}
	return job, nil
}

func createJobFromCache(ctx context.Context, tx *pachsql.Tx, cachedJob Job) (JobID, error) {
	ctx = pctx.Child(ctx, "createJobFromCache")
	// copy the row matching the cached job into the jobs table.
	var id JobID
	row := tx.QueryRowxContext(ctx, `
		INSERT INTO pjs.jobs
		(parent, program, program_hash, error, queued, processing, done)
			SELECT parent, program, program_hash, error, queued, processing, done
			FROM pjs.jobs j
		    WHERE j.id = $1
		RETURNING id`, cachedJob.ID)
	if err := row.Scan(&id); err != nil {
		return 0, errors.Wrap(err, "inserting row into pjs.jobs")
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO pjs.job_filesets
		(job_id, fileset_type, array_position, fileset)
			SELECT $1, fileset_type, array_position, fileset FROM pjs.job_filesets jf WHERE jf.job_id = $2
	`, cachedJob.ID, id)
	if err != nil {
		return 0, errors.Wrap(err, "copying job_filesets rows")
	}
	return id, nil
}

// jobCacheWriteWrapper defines code that's common to the functions that update or insert into the pjs.jobs_cache table.
func jobCacheWriteWrapper(ctx context.Context, extCtx sqlx.ExtContext, query string, args ...interface{}) error {
	result, err := extCtx.ExecContext(ctx, query, args...)
	if err != nil {
		return errors.Wrap(err, "inserting into pjs.job_cache")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "getting rowsAffected affected")
	}
	if rowsAffected == 0 {
		return errors.New("no rows affected")
	}
	return nil

}

// writeStubToJobCache creates a 'stub' entry in the cache on a cache miss or when the cache read is disabled for a specific job.
func writeStubToJobCache(ctx context.Context, extCtx sqlx.ExtContext, job Job) error {
	ctx = pctx.Child(ctx, "writeStubToJobCache")
	return jobCacheWriteWrapper(ctx, extCtx, `
		INSERT INTO pjs.job_cache (job_id, cache_read, cache_write) 
		VALUES ($1, $2, $3)
	`, job.ID, job.ReadEnabled, job.WriteEnabled)
}

// writeToJobCache does a full write, current this happens only on a cache hit.
func writeToJobCache(ctx context.Context, extCtx sqlx.ExtContext, job Job) error {
	ctx = pctx.Child(ctx, "writeToJobCache")
	if len(job.JobHash) == 0 {
		return errors.New("job hash cannot be empty")
	}
	return jobCacheWriteWrapper(ctx, extCtx, `
		INSERT INTO pjs.job_cache (job_id, job_hash, cache_read, cache_write) 
		VALUES ($1, $2, $3, $4)
	`, job.ID, job.JobHash, job.ReadEnabled, job.WriteEnabled)
}

// setJobCacheJobHash sets the job hash when a job is done and cache options were passed in -- either successfully or on an error.
func setJobCacheJobHash(ctx context.Context, extCtx sqlx.ExtContext, job Job) error {
	ctx = pctx.Child(ctx, "setJobCacheJobHash")
	if len(job.JobHash) == 0 {
		return errors.New("job hash cannot be empty")
	}
	return jobCacheWriteWrapper(ctx, extCtx, `
		UPDATE pjs.job_cache SET job_hash = $1 WHERE job_id = $2
	`, job.JobHash, job.ID)
}
