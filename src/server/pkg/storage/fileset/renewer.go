package fileset

import (
	"context"
	"sync"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// renewFunc is a function that renews a particular path.
type renewFunc func(ctx context.Context, path string, ttl time.Duration) error

// WithRenewer provides a scoped fileset renewer.
func WithRenewer(ctx context.Context, ttl time.Duration, renew renewFunc, cb func(context.Context, *Renewer) error) error {
	r := newRenewer(ttl, renew)
	cancelCtx, cf := context.WithCancel(ctx)
	eg, errCtx := errgroup.WithContext(cancelCtx)
	eg.Go(func() error {
		return r.run(errCtx)
	})
	eg.Go(func() error {
		defer cf()
		return cb(errCtx, r)
	})
	return eg.Wait()
}

// Renewer manages renewing the TTL on a set of paths.
type Renewer struct {
	ttl   time.Duration
	renew renewFunc
	mu    sync.Mutex
	paths map[string]struct{}
}

func newRenewer(ttl time.Duration, renew renewFunc) *Renewer {
	return &Renewer{
		ttl:   ttl,
		renew: renew,
		paths: make(map[string]struct{}),
	}
}

// run runs the renewer until the context is canceled.
func (r *Renewer) run(ctx context.Context) (retErr error) {
	defer func() {
		if errors.Is(ctx.Err(), context.Canceled) {
			retErr = nil
		}
	}()
	ticker := time.NewTicker(r.ttl / 2)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := r.renewPaths(ctx); err != nil {
				return err
			}
		}
	}
}

func (r *Renewer) renewPaths(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for p := range r.paths {
		if err := r.renew(ctx, p, r.ttl); err != nil {
			return err
		}
	}
	return nil
}

// Add adds p to the path set.
func (r *Renewer) Add(p string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.paths[p] = struct{}{}
}

// Remove removes p from the path set
func (r *Renewer) Remove(p string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.paths, p)
}
