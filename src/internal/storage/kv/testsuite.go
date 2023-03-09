package kv

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/pacherr"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestStore(t *testing.T, newStore func(t testing.TB) Store) {
	t.Run("PutGet", func(t *testing.T) {
		x := newStore(t)
		requirePut(t, x, []byte("key1"), []byte("value1"))
		v := requireGet(t, x, []byte("key1"))
		require.Equal(t, []byte("value1"), v)
	})
	t.Run("Exists", func(t *testing.T) {
		ctx := pctx.TestContext(t)
		x := newStore(t)
		_, err := x.Get(ctx, []byte("key1"), make([]byte, 1000))
		require.True(t, pacherr.IsNotExist(err))
		require.False(t, requireExists(t, x, []byte("key1")))

		requirePut(t, x, []byte("key1"), []byte("value1"))

		require.True(t, requireExists(t, x, []byte("key1")))
	})
}

func requireExists(t testing.TB, s Store, key []byte) bool {
	ctx := pctx.TestContext(t)
	exists, err := s.Exists(ctx, key)
	require.NoError(t, err)
	return exists
}

func requirePut(t testing.TB, s Putter, key, value []byte) {
	ctx := pctx.TestContext(t)
	require.NoError(t, s.Put(ctx, key, value))
}

func requireGet(t testing.TB, s Getter, key []byte) []byte {
	ctx := pctx.TestContext(t)
	buf := make([]byte, 1024)
	n, err := s.Get(ctx, key, buf)
	require.NoError(t, err)
	return buf[:n]
}
