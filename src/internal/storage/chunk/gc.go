package chunk

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// GarbageCollector removes unused chunks from object storage
type GarbageCollector struct {
	s            *Storage
	period       time.Duration
	isConcurrent bool
}

// NewGC returns a new garbage collector operating on s
func NewGC(s *Storage, d time.Duration, opts ...GCOption) *GarbageCollector {
	gc := &GarbageCollector{s: s, period: d}

	for _, opt := range opts {
		opt(gc)
	}

	return gc
}

// RunForever calls RunOnce until the context is cancelled, logging any errors.
func (gc *GarbageCollector) RunForever(ctx context.Context) error {
	ticker := time.NewTicker(gc.period)
	defer ticker.Stop()
	for {
		if err := gc.RunOnce(ctx); err != nil {
			select {
			case <-ctx.Done():
				return err
			default:
			}
			log.Error(ctx, "error during chunk GC", zap.Error(err))
		}
		select {
		case <-ctx.Done():
			return errors.EnsureStack(context.Cause(ctx))
		case <-ticker.C:
		}
	}
}

// RunOnce runs 1 cycle of garbage collection.
func (gc *GarbageCollector) RunOnce(ctx context.Context) (retErr error) {
	ctx, end := log.SpanContext(ctx, "RunOnce")
	defer end(log.Errorp(&retErr))
	rows, err := gc.s.db.QueryxContext(ctx, `
	SELECT chunk_id, gen, uploaded FROM storage.chunk_objects
	WHERE tombstone = true
	`)
	if err != nil {
		return errors.EnsureStack(err)
	}
	defer errors.Close(&retErr, rows, "close rows")
	eg, ctx := errgroup.WithContext(ctx)
	if !gc.isConcurrent {
		eg.SetLimit(1)
	}
	for rows.Next() {
		var ent Entry
		if err := rows.StructScan(&ent); err != nil {
			return errors.EnsureStack(err)
		}
		eg.Go(func() error {
			return gc.deleteHelper(ctx, ent)
		})
	}

	if err := eg.Wait(); err != nil {
		return errors.EnsureStack(err)
	}

	return errors.EnsureStack(rows.Err())
}

func (gc *GarbageCollector) deleteOneWithLogs(ctx context.Context, ent Entry) error {
	fields := []log.Field{zap.Stringer("chunkID", ent.ChunkID), zap.Uint64("gen", ent.Gen)}
	if !ent.Uploaded {
		log.Info(ctx, "possibility for untracked chunk", fields...)
	}
	if err := gc.deleteOne(ctx, ent); err != nil {
		return err
	}
	log.Info(ctx, "deleting object for chunk entry", fields...)

	return nil
}

func (gc *GarbageCollector) deleteOne(ctx context.Context, ent Entry) error {
	if err := gc.deleteObject(ctx, ent.ChunkID, ent.Gen); err != nil {
		return err
	}
	return gc.deleteEntry(ctx, ent.ChunkID, ent.Gen)
}

func (gc *GarbageCollector) deleteObject(ctx context.Context, chunkID ID, gen uint64) error {
	return errors.EnsureStack(gc.s.store.Delete(ctx, chunkKey(chunkID, gen)))
}

func (gc *GarbageCollector) deleteEntry(ctx context.Context, chunkID ID, gen uint64) error {
	_, err := gc.s.db.ExecContext(ctx, `
	DELETE FROM storage.chunk_objects
	WHERE chunk_id = $1 AND gen = $2 AND tombstone = TRUE
	`, chunkID, gen)
	return errors.EnsureStack(err)
}
