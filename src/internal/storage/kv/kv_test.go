package kv

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestMemStore(t *testing.T) {
	TestStore(t, func(t testing.TB) Store {
		return NewMemStore()
	})
}

func TestFSStore(t *testing.T) {
	TestStore(t, func(t testing.TB) Store {
		dir := t.TempDir()
		return NewFSStore(dir, 1024, 1<<20)
	})
}

func TestSemaphoredStore(t *testing.T) {
	TestStore(t, func(t testing.TB) Store {
		return NewSemaphored(NewMemStore(), 1, 1)
	})
}

func TestPrefixedStore(t *testing.T) {
	TestStore(t, func(t testing.TB) Store {
		return NewPrefixed(NewMemStore(), []byte("prefix"))
	})
}

func TestObjectClient(t *testing.T) {
	TestStore(t, func(t testing.TB) Store {
		c, err := obj.NewLocalClient(t.TempDir())
		require.NoError(t, err)
		return NewFromObjectClient(c, 1024, 1<<20)
	})
}
