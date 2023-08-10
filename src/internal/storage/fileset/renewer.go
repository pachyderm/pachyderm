package fileset

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/storage/renew"
)

// Renewer renews file sets by Handle
// It is a renew.StringSet wrapped for type safety
type Renewer struct {
	r *renew.StringSet
}

func newRenewer(ctx context.Context, storage *Storage, ttl time.Duration) *Renewer {
	rf := func(ctx context.Context, id string, ttl time.Duration) error {
		h, err := ParseHandle(id)
		if err != nil {
			return err
		}
		_, err = storage.SetTTL(ctx, *h, ttl)
		return err
	}
	cf := func(ctx context.Context, hStrs []string, ttl time.Duration) (string, error) {
		var hs []Handle
		for _, hStr := range hStrs {
			h, err := ParseHandle(hStr)
			if err != nil {
				return "", err
			}
			hs = append(hs, *h)
		}
		h, err := storage.Compose(ctx, hs, ttl)
		if err != nil {
			return "", err
		}
		return h.String(), nil
	}
	return &Renewer{
		r: renew.NewStringSet(ctx, ttl, rf, cf),
	}
}

func (r *Renewer) Add(ctx context.Context, h Handle) error {
	return r.r.Add(ctx, h.String())
}

func (r *Renewer) Context() context.Context {
	return r.r.Context()
}

func (r *Renewer) Close() error {
	return r.r.Close()
}
