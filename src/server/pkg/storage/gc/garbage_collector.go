package gc

import (
	"context"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
)

const (
	defaultPolling = 5 * time.Minute
	defaultTimeout = 30 * time.Minute
)

type garbageCollector struct {
	objClient        obj.Client
	db               *gorm.DB
	polling, timeout time.Duration
}

// Run runs the garbage collector.
func Run(ctx context.Context, objClient obj.Client, db *gorm.DB, opts ...Option) error {
	gc := &garbageCollector{
		objClient: objClient,
		db:        db,
		polling:   defaultPolling,
		timeout:   defaultTimeout,
	}
	for _, opt := range opts {
		opt(gc)
	}
	return gc.pollingFunc(ctx)
}

func (gc *garbageCollector) maybeDeleteTemporaryRefs(ctx context.Context) error {
	seconds := strconv.Itoa(int(gc.timeout / time.Second))
	stmtFuncs := []statementFunc{
		func(txn *gorm.DB) *gorm.DB {
			return txn.Exec(`
				DELETE FROM refs
				WHERE sourcetype = 'temporary'
				AND created < NOW() - INTERVAL '` + seconds + `' second
			`)
		},
	}
	return runTransaction(ctx, gc.db, stmtFuncs, nil)
}

func (gc *garbageCollector) maybeDeleteChunks(ctx context.Context) error {
	chunksToDelete := []chunkModel{}
	stmtFuncs := []statementFunc{
		func(txn *gorm.DB) *gorm.DB {
			return txn.Raw(`
				SELECT chunks.chunk
				FROM chunks
				LEFT OUTER JOIN refs
				ON chunks.chunk = refs.chunk
				WHERE refs.chunk IS NULL
			`).Scan(&chunksToDelete)
		},
	}
	if err := runTransaction(ctx, gc.db, stmtFuncs, nil); err != nil {
		return err
	}
	return gc.deleteChunks(ctx, convertChunks(chunksToDelete))
}

func (gc *garbageCollector) pollingFunc(ctx context.Context) error {
	return retry("polling", func() error {
		for {
			if err := gc.maybeDeleteTemporaryRefs(ctx); err != nil {
				return err
			}
			if err := gc.maybeDeleteChunks(ctx); err != nil {
				return err
			}
			select {
			case <-time.After(gc.polling):
			case <-ctx.Done():
				return nil
			}
		}
	})
}

func (gc *garbageCollector) deleteChunks(ctx context.Context, chunks []string) error {
	if len(chunks) == 0 {
		return nil
	}
	// Mark the chunks as deleting.
	var toDelete []string
	var err error
	if err := retry("marking chunks as deleting", func() error {
		toDelete, err = gc.markChunksDeleting(ctx, chunks)
		return err
	}); err != nil {
		return err
	}
	// Delete the chunks from object storage.
	if err := retry("deleting chunks", func() error {
		chunks := toDelete
		for len(chunks) > 0 {
			if err := gc.objClient.Delete(ctx, chunks[0]); err != nil {
				return err
			}
			chunks = chunks[1:]
		}
		return nil
	}); err != nil {
		return err
	}
	// Remove the chunk rows.
	transitiveDeletes := []string{}
	if err := retry("removing chunk rows", func() error {
		transitiveDeletes, err = gc.removeChunkRows(ctx, toDelete)
		return err
	}); err != nil {
		return err
	}
	return gc.deleteChunks(ctx, transitiveDeletes)
}

func (gc *garbageCollector) markChunksDeleting(ctx context.Context, chunks []string) ([]string, error) {
	if len(chunks) == 0 {
		return nil, nil
	}
	chunksDeleting := []chunkModel{}
	stmtFuncs := []statementFunc{
		func(txn *gorm.DB) *gorm.DB {
			// Set the deleting field for the passed in chunks, excluding
			// the chunks that had a reference added before the deleting process
			// began.
			return txn.Raw(`
				UPDATE chunks
				SET deleting = NOW()
				WHERE chunk IN (?)
				AND chunk NOT IN (
					SELECT DISTINCT chunk
					FROM refs
					WHERE chunk IN (?)
				)
				RETURNING chunk
			`, chunks, chunks).Scan(&chunksDeleting)
		},
	}
	if err := runTransaction(ctx, gc.db, stmtFuncs, nil); err != nil {
		return nil, err
	}
	return convertChunks(chunksDeleting), nil
}

func (gc *garbageCollector) removeChunkRows(ctx context.Context, chunks []string) ([]string, error) {
	if len(chunks) == 0 {
		return nil, nil
	}
	deletedChunks := []chunkModel{}
	stmtFuncs := []statementFunc{
		func(txn *gorm.DB) *gorm.DB {
			// Delete the chunks and references from the deleted chunks.
			// Return the chunks that should be transitively deleted (if any).
			return txn.Raw(`
				WITH deleted_chunks AS (
					DELETE FROM chunks
					WHERE chunk IN (?)
					RETURNING chunk
				), deleted_refs AS (
					DELETE FROM refs
					USING deleted_chunks
					WHERE refs.sourcetype = 'chunk'
					AND refs.source = deleted_chunks.chunk
					RETURNING refs.chunk
				)
				SELECT deleted_refs.chunk
				FROM deleted_refs
				JOIN refs
				ON deleted_refs.chunk = refs.chunk
				GROUP BY 1
				HAVING COUNT(*) = 1
		`, chunks).Scan(&deletedChunks)
		},
	}
	if err := runTransaction(ctx, gc.db, stmtFuncs, nil); err != nil {
		return nil, err
	}
	return convertChunks(deletedChunks), nil
}
