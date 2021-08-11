package fileset

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/storage/renew"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
)

// Renewer renews filesets by ID
// It is a renew.StringSet wrapped for type safety
type Renewer struct {
	r *renew.StringSet
}

func newRenewer(ctx context.Context, tracker track.Tracker, ttl time.Duration) *Renewer {
	return &Renewer{
		r: renew.NewStringSet(ctx, ttl, func(ctx context.Context, id string, ttl time.Duration) error {
			_, err := tracker.SetTTL(ctx, id, ttl)
			return err
		}),
	}
}

func (r *Renewer) Add(id ID) {
	r.r.Add(id.TrackerID())
}

func (r *Renewer) Remove(id ID) {
	r.r.Remove(id.TrackerID())
}

func (r *Renewer) Context() context.Context {
	return r.r.Context()
}

func (r *Renewer) Close() error {
	return r.r.Close()
}
