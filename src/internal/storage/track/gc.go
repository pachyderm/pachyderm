package track

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
)

// Deleter is used to delete external data associated with a tracked object
type Deleter interface {
	DeleteTx(tx *pachsql.Tx, id string) error
}

// DeleterMux returns a Deleter based on the id being deleted
type DeleterMux func(string) Deleter

// DeleteTx implements Deleter
func (dm DeleterMux) DeleteTx(tx *pachsql.Tx, id string) error {
	deleter := dm(id)
	if deleter == nil {
		return errors.Errorf("deleter mux does not have deleter for (%s)", id)
	}
	return errors.EnsureStack(deleter.DeleteTx(tx, id))
}

// GarbageCollector periodically runs garbage collection on tracker objects
type GarbageCollector struct {
	tracker Tracker
	period  time.Duration
	deleter Deleter
}

// NewGarbageCollector returns a garbage collector monitoring tracker, and kicking off a cycle every period.
// It will use deleter to deleted associated data before deleting objects from the Tracker
func NewGarbageCollector(tracker Tracker, period time.Duration, deleter Deleter) *GarbageCollector {
	return &GarbageCollector{
		tracker: tracker,
		period:  period,
		deleter: deleter,
	}
}

// RunForever runs the gc loop, until the context is cancelled. It returns context.Canceled on exit.
func (gc *GarbageCollector) RunForever(ctx context.Context) error {
	ticker := time.NewTicker(gc.period)
	defer ticker.Stop()
	for {
		if err := gc.RunUntilEmpty(ctx); err != nil {
			log.Error(ctx, "gc: error from RunUntilEmpty", zap.Error(err))
		}
		select {
		case <-ctx.Done():
			return errors.EnsureStack(context.Cause(ctx))
		case <-ticker.C:
		}
	}
}

// RunUntilEmpty calls RunOnce repeatedly until it returns an error or 0.
func (gc *GarbageCollector) RunUntilEmpty(ctx context.Context) error {
	for {
		n, err := gc.RunOnce(ctx)
		if err != nil {
			return err
		}
		if n == 0 {
			break
		}
	}
	return nil
}

// RunOnce run's one cycle of garbage collection.
func (gc *GarbageCollector) RunOnce(ctx context.Context) (n int, retErr error) {
	ctx, done := log.SpanContext(ctx, "RunOnce")
	defer done(log.Errorp(&retErr), zap.Int("n", n))
	err := gc.tracker.IterateDeletable(ctx, func(id string) error {
		if err := gc.deleteObject(ctx, id); err != nil {
			log.Error(ctx, "error deleting object", zap.String("id", id), zap.Error(err))
		} else {
			n++
		}
		return nil
	})
	return n, errors.EnsureStack(err)
}

func (gc *GarbageCollector) deleteObject(ctx context.Context, id string) error {
	db := gc.tracker.DB()
	return dbutil.WithTx(ctx, db, func(ctx context.Context, tx *pachsql.Tx) error {
		if err := gc.tracker.DeleteTx(tx, id); err != nil {
			return errors.EnsureStack(err)
		}
		return errors.EnsureStack(gc.deleter.DeleteTx(tx, id))
	})
}
