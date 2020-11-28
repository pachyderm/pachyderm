package fileset

import (
	"context"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

var (
	// ErrPathExists path already exists
	ErrPathExists = errors.Errorf("path already exists")
	// ErrPathNotExists path does not exist
	ErrPathNotExists = errors.Errorf("path does not exist")
	// ErrNoTTLSet no ttl set on path
	ErrNoTTLSet = errors.Errorf("no ttl set on path")
)

// Store stores filesets. A fileset is a path -> index relationship
// All filesets exist in the same keyspace and can be merged by prefix
type Store interface {
	Set(ctx context.Context, p string, md *Metadata) error
	Get(ctx context.Context, p string) (*Metadata, error)
	Delete(ctx context.Context, p string) error
	Walk(ctx context.Context, prefix string, cb func(string) error) error
}

func StoreTestSuite(t *testing.T, newStore func(t testing.TB) Store) {
	ctx := context.Background()
	t.Run("SetGet", func(t *testing.T) {
		x := newStore(t)
		md := &Metadata{}
		require.NoError(t, x.Set(ctx, "test", md))
		actual, err := x.Get(ctx, "test")
		require.NoError(t, err)
		require.Equal(t, md, actual)
	})
	t.Run("Delete", func(t *testing.T) {
		x := newStore(t)
		require.NoError(t, x.Delete(ctx, "keys that don't exist should not cause delete to error"))
		md := &Metadata{}
		require.NoError(t, x.Set(ctx, "test", md))
		require.NoError(t, x.Delete(ctx, "test"))
		_, err := x.Get(ctx, "test")
		require.Equal(t, ErrPathNotExists, err)
	})
	t.Run("Walk", func(t *testing.T) {
		x := newStore(t)
		md := &Metadata{}
		ps := []string{"test/1", "test/2", "test/3"}
		for _, p := range ps {
			require.NoError(t, x.Set(ctx, p, md))
		}
		require.NoError(t, x.Walk(ctx, "test", func(p string) error {
			require.Equal(t, ps[0], p)
			ps = ps[1:]
			return nil
		}))
		require.Equal(t, 0, len(ps))
	})
}

func copyPath(ctx context.Context, src, dst Store, srcPath, dstPath string) error {
	md, err := src.Get(ctx, srcPath)
	if err != nil {
		return err
	}
	return dst.Set(ctx, dstPath, md)
}
