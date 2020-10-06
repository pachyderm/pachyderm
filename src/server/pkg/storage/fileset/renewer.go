package fileset

import (
	"context"
	"sync"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
)

// Renewer manages renewing the TTL on a set of paths.
type Renewer struct {
	storage *Storage
	ttl     time.Duration
	mu      sync.Mutex
	paths   map[string]struct{}
}

func newRenewer(s *Storage, ttl time.Duration) *Renewer {
	return &Renewer{
		storage: s,
		ttl:     ttl,
		paths:   make(map[string]struct{}),
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
		if _, err := r.storage.SetTTL(ctx, p, r.ttl); err != nil {
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
