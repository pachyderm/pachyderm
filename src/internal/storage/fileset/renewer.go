package fileset

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/storage/renew"
)

// Renewer renews file sets by ID
// It is a renew.StringSet wrapped for type safety
type Renewer struct {
	r *renew.StringSet
}

func newRenewer(ctx context.Context, storage *Storage, ttl time.Duration) *Renewer {
	rf := func(ctx context.Context, handleStr string, ttl time.Duration) error {
		handle, err := ParseHandle(handleStr)
		if err != nil {
			return err
		}
		_, err = storage.SetTTL(ctx, handle, ttl)
		return err
	}
	cf := func(ctx context.Context, handleStrs []string, ttl time.Duration) (string, error) {
		var handles []*Handle
		for _, handleStr := range handleStrs {
			handle, err := ParseHandle(handleStr)
			if err != nil {
				return "", err
			}
			handles = append(handles, handle)
		}
		handle, err := storage.Compose(ctx, handles, ttl)
		if err != nil {
			return "", err
		}
		return handle.HexString(), nil
	}
	return &Renewer{
		r: renew.NewStringSet(ctx, ttl, rf, cf),
	}
}

func (r *Renewer) Add(ctx context.Context, handle *Handle) error {
	return r.r.Add(ctx, handle.HexString())
}

func (r *Renewer) Context() context.Context {
	return r.r.Context()
}

func (r *Renewer) Close() error {
	return r.r.Close()
}
