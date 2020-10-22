package tracker

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

type Deleter interface {
	Delete(ctx context.Context, id string) error
}

type DeleterFactory func(string) Deleter

type GC struct {
	tracker    Tracker
	period     time.Duration
	getDeleter DeleterFactory
}

func NewGC(tracker Tracker, period time.Duration, getDeleter DeleterFactory) *GC {
	return &GC{
		tracker:    tracker,
		period:     period,
		getDeleter: getDeleter,
	}
}

func (gc *GC) Run(ctx context.Context) error {
	ticker := time.NewTicker(gc.period)
	defer ticker.Stop()
	for {
		ctx, cf := context.WithTimeout(ctx, gc.period/2)
		if err := gc.runOnce(ctx); err != nil {
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

func (gc *GC) runOnce(ctx context.Context) error {
	return gc.tracker.IterateExpired(ctx, func(id string) error {
		if err := gc.deleteObject(ctx, id); err != nil {
			logrus.Errorf("error deleting object (%s): %v", id, err)
		}
		return nil
	})
}

func (gc *GC) deleteObject(ctx context.Context, id string) error {
	deleter := gc.getDeleter(id)
	if err := gc.tracker.MarkTombstone(ctx, id); err != nil {
		return err
	}
	if err := deleter.Delete(ctx, id); err != nil {
		return err
	}
	if err := gc.tracker.DeleteObject(ctx, id); err != nil {
		return err
	}
	return nil
}
