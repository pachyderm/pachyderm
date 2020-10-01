package fileset

import (
	"context"
	"sync"
	"time"
)

// Renewer manages renewing the TTL on a set of paths
type Renewer struct {
	storage *Storage
	ttl     time.Duration

	ctx  context.Context
	cf   context.CancelFunc
	err  error
	done chan struct{}

	mu    sync.Mutex
	paths map[string]struct{}
}

// NewRenewer creates a renewer with 0 paths.
func NewRenewer(s *Storage, ttl time.Duration) *Renewer {
	ctx, cf := context.WithCancel(context.Background())
	r := &Renewer{
		ttl:   ttl,
		ctx:   ctx,
		cf:    cf,
		done:  make(chan struct{}),
		paths: make(map[string]struct{}),
	}
	go r.run(ctx)
	return r
}

func (r *Renewer) run(ctx context.Context) {
	ticker := time.NewTicker(r.ttl / 2)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			close(r.done)
			return
		case <-ticker.C:
			ctx, cf := context.WithTimeout(ctx, r.ttl/3)
			if err := r.renewPaths(ctx); err != nil {
				r.cf()
				r.err = err
			}
			cf()
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

// Close stops renew loop. If an error is returned some of the paths may have been deleted.
func (r *Renewer) Close() error {
	r.cf()
	<-r.done
	return r.err
}
