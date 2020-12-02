package track

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Deleter is used to delete data external to a tracker associated with a tracked object
type Deleter interface {
	Delete(ctx context.Context, id string) error
}

// DeleterMux returns a Deleter based on the id being deleted
type DeleterMux func(string) Deleter

// Delete implements Deleter
func (dm DeleterMux) Delete(ctx context.Context, id string) error {
	deleter := dm(id)
	if deleter == nil {
		return errors.Errorf("deleter mux does not have deleter for (%s)", id)
	}
	return deleter.Delete(ctx, id)
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

// Run runs the gc loop, until the context is cancelled. It returns ErrContextCancell on exit.
func (gc *GarbageCollector) Run(ctx context.Context) error {
	ticker := time.NewTicker(gc.period)
	defer ticker.Stop()
	for {
		if err := func() error {
			ctx, cf := context.WithTimeout(ctx, gc.period/2)
			defer cf()
			return gc.runUntilEmpty(ctx)
		}(); err != nil {
			logrus.Errorf("gc: %v", err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (gc *GarbageCollector) runUntilEmpty(ctx context.Context) error {
	for {
		n, err := gc.runOnce(ctx)
		if err != nil {
			return err
		}
		if n == 0 {
			break
		}
	}
	return nil
}

func (gc *GarbageCollector) runOnce(ctx context.Context) (int, error) {
	var n int
	err := gc.tracker.IterateDeletable(ctx, func(id string) error {
		if err := gc.deleteObject(ctx, id); err != nil {
			logrus.Errorf("error deleting object (%s): %v", id, err)
		} else {
			n++
		}
		return nil
	})
	return n, err
}

func (gc *GarbageCollector) deleteObject(ctx context.Context, id string) error {
	if err := gc.tracker.MarkTombstone(ctx, id); err != nil {
		return err
	}
	if err := gc.deleter.Delete(ctx, id); err != nil {
		return err
	}
	if err := gc.tracker.FinishDelete(ctx, id); err != nil {
		return err
	}
	return nil
}
