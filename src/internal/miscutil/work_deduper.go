package miscutil

import (
	"context"
	"sync"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

type WorkDeduper[K comparable] struct {
	futures sync.Map
}

// Do is not properly documented.
//
// Concurrent calls to Do will block until cb has been completed by one of them, then the others will run, possibly concurrently.
// The motivating use case is to eliminate network round trips when populating a cache.
// Callers should check and populate the cache inside cb.
//
// TODO: document
func (wd *WorkDeduper[K]) Do(ctx context.Context, k K, cb func() error) error {
	fut, created := wd.getOrCreateFuture(k)
	if created {
		defer wd.removeFuture(k)
		fut.err = cb()
		close(fut.done)
	}
	return fut.await(ctx)
}

func (wd *WorkDeduper[K]) getOrCreateFuture(key K) (*future, bool) {
	fut := &future{
		done: make(chan struct{}),
	}
	x, loaded := wd.futures.LoadOrStore(key, fut)
	return x.(*future), !loaded
}

func (wd *WorkDeduper[K]) removeFuture(k K) {
	wd.futures.Delete(k)
}

type future struct {
	done chan struct{}
	err  error
}

func (f future) await(ctx context.Context) error {
	select {
	case <-f.done:
		return f.err
	case <-ctx.Done():
		return errors.EnsureStack(context.Cause(ctx))
	}
}
