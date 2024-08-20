package pjsdb

import (
	"context"
	"database/sql"
	"github.com/pachyderm/pachyderm/v2/src/pjs"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
)

// ListQueues returns a list of Queue objects.
func ListQueues(ctx context.Context, db *pachsql.DB, req IterateQueuesRequest) ([]Queue, error) {
	ctx = pctx.Child(ctx, "listQueue")
	var queues []Queue
	if err := ForEachQueue(ctx, db, req, func(queue Queue) error {
		q := queue
		queues = append(queues, q)
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "list queue")
	}
	return queues, nil
}

// DequeueAndProcess processes the first job in a given queue, and removes that element from queue
func DequeueAndProcess(ctx context.Context, tx *pachsql.Tx, programHash []byte) (JobID, error) {
	ctx = pctx.Child(ctx, "dequeue and process")
	var jobID JobID
	if err := tx.QueryRowxContext(ctx, `
		WITH updated AS (
			SELECT id
			FROM pjs.jobs
			WHERE processing IS NULL AND done IS NULL AND queued IS NOT NULL AND program_hash = $1
			ORDER BY queued
			LIMIT 1
			FOR UPDATE
		)
		UPDATE pjs.jobs
		SET processing = CURRENT_TIMESTAMP
		FROM updated
		WHERE pjs.jobs.id = updated.id
		RETURNING pjs.jobs.id
	`, programHash).Scan(&jobID); err != nil {
		// todo(muyang): should not return an error, just await
		if errors.Is(err, sql.ErrNoRows) {
			return 0, &DequeueFromEmptyQueueError{ID: string(programHash)}
		}
		return 0, errors.Wrap(err, "dequeue and process")
	}
	return jobID, nil
}

// CompleteError is called when job processing has an error. It updates job err code and
// done timestamp in database.
func CompleteError(ctx context.Context, tx *pachsql.Tx, jobID JobID, errCode pjs.JobErrorCode) error {
	ctx = pctx.Child(ctx, "complete error")
	_, err := tx.ExecContext(ctx, `
		UPDATE pjs.jobs
		SET done = CURRENT_TIMESTAMP, error = $1
		WHERE id = $2
	`, errCode, jobID)
	if err != nil {
		return errors.Wrapf(err, "complete error")
	}
	return nil
}

// CompleteOk is called when job processing without any error. It updates done timestamp
// output filesets and in database.
func CompleteOk(ctx context.Context, tx *pachsql.Tx, jobID JobID, outputs []string) error {
	ctx = pctx.Child(ctx, "complete ok")
	_, err := tx.ExecContext(ctx, `
		UPDATE pjs.jobs
		SET done = CURRENT_TIMESTAMP
		WHERE id = $1
	`, jobID)
	if err != nil {
		return errors.Wrapf(err, "complete ok: update state to done")
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
