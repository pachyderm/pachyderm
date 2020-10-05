package index

import (
	"context"
	"testing"

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
	withStore(func(x Store) {
		idx := &Index{}
		require.NoError(t, x.PutIndex(ctx, "test", idx, 0))
		actual, err := x.GetIndex(ctx, "test")
		require.NoError(t, err)
		require.Equal(t, idx, actual)
	})
}
