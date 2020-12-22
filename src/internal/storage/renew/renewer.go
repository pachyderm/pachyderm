package renew

import (
	"context"
	"errors"
	"time"
)

// Func is a function called to renew something for ttl time.
type Func func(ctx context.Context, ttl time.Duration) error

// Renewer manages renewing something in the background
type Renewer struct {
	renewFunc Func
	ttl       time.Duration

	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
	err    error
}

// NewRenewer returns a renewer that will call renewFunc with ttl every period;
// where period will be some fraction ttl.
// If ctx is cancelled the renewer will be closed.
func NewRenewer(ctx context.Context, ttl time.Duration, renewFunc Func) *Renewer {
	ctx, cancel := context.WithCancel(ctx)
	r := &Renewer{
		renewFunc: renewFunc,
		ttl:       ttl,

		ctx:    ctx,
		cancel: cancel,
		done:   make(chan struct{}),
	}
	go func() {
		r.err = r.renewLoop(ctx)
		r.cancel()
		close(r.done)
	}()
	return r
}

// Context returns a context which will be cancelled when the renewer is closed
func (r *Renewer) Context() context.Context {
	return r.ctx
}

// Close closes the renewer, stopping the background renewal. Close is idempotent.
func (r *Renewer) Close() error {
	r.cancel()
	<-r.done
	return r.err
}

func (r *Renewer) renewLoop(ctx context.Context) (retErr error) {
	defer func() {
		if errors.Is(ctx.Err(), context.Canceled) {
			retErr = nil
		}
	}()
	ticker := time.NewTicker(r.ttl / 3)
	defer ticker.Stop()
	for {
		if err := func() error {
			ctx, cf := context.WithTimeout(ctx, r.ttl/3)
			defer cf()
			return r.renewFunc(ctx, r.ttl)
		}(); err != nil {
			return err
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
