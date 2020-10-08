package fileset

import (
	"context"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
)

var (
	ErrPathNotExists = errors.Errorf("path does not exist")
	ErrNoTTLSet      = errors.Errorf("no ttl set on path")
)

type PathStore interface {
	PutIndex(ctx context.Context, p string, idx *index.Index, ttl time.Duration) error
	GetIndex(ctx context.Context, p string) (*index.Index, error)
	Delete(ctx context.Context, p string) error
	Walk(ctx context.Context, prefix string, cb func(string) error) error

	SetTTL(ctx context.Context, p string, ttl time.Duration) (time.Time, error)
	GetExpiresAt(ctx context.Context, p string) (time.Time, error)
}

func copyPath(ctx context.Context, src, dst PathStore, srcPath, dstPath string) error {
	idx, err := src.GetIndex(ctx, srcPath)
	if err != nil {
		return err
	}
	return dst.PutIndex(ctx, dstPath, idx, 0)
}

func PathStoreTestSuite(t *testing.T, withStore func(func(PathStore))) {
	ctx := context.Background()
	t.Run("PutGet", func(t *testing.T) {
		withStore(func(x PathStore) {
			idx := &index.Index{}
			require.NoError(t, x.PutIndex(ctx, "test", idx, 0))
			actual, err := x.GetIndex(ctx, "test")
			require.NoError(t, err)
			require.Equal(t, idx, actual)
		})
	})
	t.Run("SetTTL", func(t *testing.T) {
		withStore(func(x PathStore) {
			idx := &index.Index{}
			require.NoError(t, x.PutIndex(ctx, "test", idx, time.Hour))
			expiresAt, err := x.SetTTL(ctx, "test", time.Hour)
			require.NoError(t, err)
			require.True(t, expiresAt.After(time.Now()))
		})
	})
	t.Run("Delete", func(t *testing.T) {
		withStore(func(x PathStore) {
			require.NoError(t, x.Delete(ctx, "keys that don't exist should not cause delete to error"))
			idx := &index.Index{}
			require.NoError(t, x.PutIndex(ctx, "test", idx, time.Hour))
			require.NoError(t, x.Delete(ctx, "test"))
			_, err := x.GetIndex(ctx, "test")
			require.Equal(t, ErrPathNotExists, err)
		})
	})
}
