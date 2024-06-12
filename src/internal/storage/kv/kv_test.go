package kv

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"gocloud.dev/blob/fileblob"
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

func TestBucket(t *testing.T) {
	TestStore(t, func(t testing.TB) Store {
		b, err := fileblob.OpenBucket(t.TempDir(), &fileblob.Options{})
		require.NoError(t, err)
		return NewFromBucket(b, 1024, 1<<20)
	})
}
