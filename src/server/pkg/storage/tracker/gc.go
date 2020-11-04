package tracker

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Deleter interface {
	Delete(ctx context.Context, id string) error
}

type DeleterMux func(string) Deleter

func (dm DeleterMux) Delete(ctx context.Context, id string) error {
	deleter := dm(id)
	if deleter == nil {
		return errors.Errorf("deleter mux does not have deleter for (%s)", id)
	}
	return deleter.Delete(ctx, id)
}

type GarbageCollector struct {
	tracker Tracker
	period  time.Duration
	deleter Deleter
}

func NewGarbageCollector(tracker Tracker, period time.Duration, deleter Deleter) *GarbageCollector {
	return &GarbageCollector{
		tracker: tracker,
		period:  period,
		deleter: deleter,
	}
}

func (GarbageCollector *GarbageCollector) Run(ctx context.Context) error {
	ticker := time.NewTicker(GarbageCollector.period)
	defer ticker.Stop()
	for {
		ctx, cf := context.WithTimeout(ctx, GarbageCollector.period/2)
		if err := GarbageCollector.runUntilEmpty(ctx); err != nil {
			logrus.Error(err)
		}
		cf()
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (GarbageCollector *GarbageCollector) runUntilEmpty(ctx context.Context) error {
	for {
		n, err := GarbageCollector.runOnce(ctx)
		if err != nil {
			return err
		}
		if n == 0 {
			break
		}
	}
	return nil
}

func (GarbageCollector *GarbageCollector) runOnce(ctx context.Context) (int, error) {
	var n int
	err := GarbageCollector.tracker.IterateExpired(ctx, func(id string) error {
		if err := GarbageCollector.deleteObject(ctx, id); err != nil {
			logrus.Errorf("error deleting object (%s): %v", id, err)
		} else {
			n++
		}
		return nil
	})
	return n, err
}

func (GarbageCollector *GarbageCollector) deleteObject(ctx context.Context, id string) error {
	if err := GarbageCollector.tracker.MarkTombstone(ctx, id); err != nil {
		return err
	}
	if err := GarbageCollector.deleter.Delete(ctx, id); err != nil {
		return err
	}
	if err := GarbageCollector.tracker.FinishDelete(ctx, id); err != nil {
		return err
	}
	return nil
}
