package chunk

import (
	"context"
	"github.com/pachyderm/pachyderm/v2/src/internal/meters"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/renew"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
)

type Renewer struct {
	ss *renew.StringSet
}

func NewRenewer(ctx context.Context, tr track.Tracker, name string, ttl time.Duration) *Renewer {
	ctx = pctx.Child(ctx, "trackerRenewer", pctx.WithCounter("renewals", 0))
	renewFunc := func(ctx context.Context, x string, ttl time.Duration) error {
		_, err := tr.SetTTL(ctx, x, ttl)
		meters.Inc(ctx, "renewals", 1)
		return errors.EnsureStack(err)
	}
	composeFunc := renew.NewTmpComposer(tr, name)
	return &Renewer{
		ss: renew.NewStringSet(ctx, ttl, renewFunc, composeFunc),
	}
}

func (r *Renewer) Add(ctx context.Context, id ID) error {
	return r.ss.Add(ctx, id.TrackerID())
}

func (r *Renewer) Close() error {
	return r.ss.Close()
}
