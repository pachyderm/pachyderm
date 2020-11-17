package fileset

import (
	"context"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
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
	PutIndex(ctx context.Context, p string, idx *index.Index) error
	GetIndex(ctx context.Context, p string) (*index.Index, error)
	Delete(ctx context.Context, p string) error
	Walk(ctx context.Context, prefix string, cb func(string) error) error
}

// StoreTestSuite runs tests to ensure Stores returned from newStore correctly implement Store.
func StoreTestSuite(t *testing.T, newStore func(t testing.TB) Store) {
	ctx := context.Background()
	t.Run("PutGet", func(t *testing.T) {
		x := newStore(t)
		idx := &index.Index{}
		require.NoError(t, x.PutIndex(ctx, "test", idx))
		actual, err := x.GetIndex(ctx, "test")
		require.NoError(t, err)
		require.Equal(t, idx, actual)
	})
	t.Run("Delete", func(t *testing.T) {
		x := newStore(t)
		require.NoError(t, x.Delete(ctx, "keys that don't exist should not cause delete to error"))
		idx := &index.Index{}
		require.NoError(t, x.PutIndex(ctx, "test", idx))
		require.NoError(t, x.Delete(ctx, "test"))
		_, err := x.GetIndex(ctx, "test")
		require.Equal(t, ErrPathNotExists, err)
	})
}

func copyPath(ctx context.Context, src, dst Store, srcPath, dstPath string) error {
	idx, err := src.GetIndex(ctx, srcPath)
	if err != nil {
		return err
	}
	return dst.PutIndex(ctx, dstPath, idx)
}
