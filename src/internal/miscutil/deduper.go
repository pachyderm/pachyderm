package miscutil

import (
	"context"
	"sync"
)

type WorkDeduper struct {
	futures sync.Map
}

func (wd *WorkDeduper) Do(ctx context.Context, k interface{}, fn func() (interface{}, error)) error {
	fut, created := wd.getOrCreateFuture(k)
	if created {
		defer wd.removeFuture(k)
		fut.value, fut.err = fn()
		close(fut.done)
	}
	if err := fut.await(ctx); err != nil {
		return err
	}
	return nil
}

func (wd *WorkDeduper) getOrCreateFuture(key interface{}) (ret *future, created bool) {
	done := make(chan struct{}, 1)
	fut := &future{
		done: done,
	}
	x, loaded := wd.futures.LoadOrStore(key, fut)
	return x.(*future), loaded
}

func (wd *WorkDeduper) removeFuture(k interface{}) {
	wd.futures.Delete(k)
}

type future struct {
	done  chan struct{}
	err   error
	value interface{}
}

func (f future) await(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-f.done:
		return nil
	}
}
