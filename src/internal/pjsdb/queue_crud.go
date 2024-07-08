package pjsdb

import (
	"context"

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
