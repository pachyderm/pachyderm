package pjsdb

import (
	"context"
	"database/sql"

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

type DequeueResponse struct {
	ID         JobID
	JobContext JobContext
}

// DequeueAndProcess processes the first job in a given queue, and removes that element from queue
func DequeueAndProcess(ctx context.Context, tx *pachsql.Tx, programHash []byte) (*DequeueResponse, error) {
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &DequeueFromEmptyQueueError{ID: string(programHash)}
		}
		return nil, errors.Wrap(err, "dequeue and process")
	}
	jobCtx, err := CreateJobContext(ctx, tx, jobID)
	if err != nil {
		return nil, errors.Wrap(err, "dequeue and process")
	}
	resp := &DequeueResponse{
		ID:         jobID,
		JobContext: jobCtx,
	}
	return resp, nil
}
