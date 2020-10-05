package index

import (
	"context"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

func TestPGStore(t *testing.T) {
	StoreTestSuite(t, func(cb func(Store)) {
		WithTestStore(t, cb)
	})
}

func StoreTestSuite(t *testing.T, withStore func(func(Store))) {
	ctx := context.Background()
	t.Run("PutGet", func(t *testing.T) {
		withStore(func(x Store) {
			idx := &Index{}
			require.NoError(t, x.PutIndex(ctx, "test", idx, 0))
			actual, err := x.GetIndex(ctx, "test")
			require.NoError(t, err)
			require.Equal(t, idx, actual)
		})
	})
	t.Run("SetTTL", func(t *testing.T) {
		withStore(func(x Store) {
			idx := &Index{}
			require.NoError(t, x.PutIndex(ctx, "test", idx, time.Hour))
			expiresAt, err := x.SetTTL(ctx, "test", time.Hour)
			require.NoError(t, err)
			require.True(t, expiresAt.After(time.Now()))
		})
	})
	t.Run("Delete", func(t *testing.T) {
		withStore(func(x Store) {
			require.NoError(t, x.Delete(ctx, "keys that don't exist should not cause delete to error"))
			idx := &Index{}
			require.NoError(t, x.PutIndex(ctx, "test", idx, time.Hour))
			require.NoError(t, x.Delete(ctx, "test"))
			_, err := x.GetIndex(ctx, "test")
			require.Equal(t, ErrPathNotExists, err)
		})
	})
}
