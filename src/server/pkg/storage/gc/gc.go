package gc

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	log "github.com/sirupsen/logrus"
)

type GC struct {
	db      *sqlx.DB
	objc    obj.Client
	tracker chunk.Tracker
	period  time.Duration
}

func NewGC(db *sqlx.DB, objc obj.Client, tracker chunk.Tracker, pollingPeriod time.Duration) *GC {
	return &GC{
		objc:    objc,
		tracker: tracker,
		period:  pollingPeriod,
		db:      db,
	}
}

func (gc *GC) Run(ctx context.Context) error {
	ticker := time.NewTicker(gc.period)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := gc.runOnce(ctx); err != nil {
				log.Error(err)
			}
		}
	}
}

const deleteableQuery = `
	SELECT hash_id FROM storage.chunks
	WHERE int_id IN (
		SELECT to FROM storage.chunk_refs
		WHERE count(from) = 0
		GROUP BY to
	)
	AND hash_id NOT IN (SELECT chunk_id FROM storage.paths)
`

// runOnce is one pass of the garbase collector.
// The query still needs some work. We want to expire all the semantic paths that need to be expired.
// then list all the chunks not referenced by another chunk or a semantic path.
// these do not necessarily need to be done in the same transaction.
func (gc *GC) runOnce(ctx context.Context) error {
	tmpID := fmt.Sprintf("gc-%d", time.Now().UnixNano())
	client := chunk.NewClient(gc.objc, gc.tracker, tmpID)
	defer client.Close()

	rows, err := gc.db.QueryContext(ctx, deleteableQuery)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var chunkID []byte
		if err := rows.Scan(&chunkID); err != nil {
			return err
		}
		if err := client.Delete(ctx, chunkID); err != nil {
			return err
		}
	}
	return rows.Err()
}
