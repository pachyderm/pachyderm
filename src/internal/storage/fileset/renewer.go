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
	rf := func(ctx context.Context, id string, ttl time.Duration) error {
		fsid, err := ParseID(id)
		if err != nil {
			return err
		}
		_, err = storage.SetTTL(ctx, *fsid, ttl)
		return err
	}
	cf := func(ctx context.Context, ids []string, ttl time.Duration) (string, error) {
		var fsids []ID
		for _, id := range ids {
			fsid, err := ParseID(id)
			if err != nil {
				return "", err
			}
			fsids = append(fsids, *fsid)
		}
		fsid, err := storage.Compose(ctx, fsids, ttl)
		if err != nil {
			return "", err
		}
		return fsid.HexString(), nil
	}
	return &Renewer{
		r: renew.NewStringSet(ctx, ttl, rf, cf),
	}
}

func (r *Renewer) Add(ctx context.Context, id ID) error {
	return r.r.Add(ctx, id.HexString())
}

func (r *Renewer) Context() context.Context {
	return r.r.Context()
}

func (r *Renewer) Close() error {
	return r.r.Close()
}
