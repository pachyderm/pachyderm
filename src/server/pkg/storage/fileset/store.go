package fileset

import (
	"context"
	"testing"

	_ "github.com/lib/pq"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

var (
	ErrPathNotExists = errors.Errorf("path does not exist")
	ErrNoTTLSet      = errors.Errorf("no ttl set on path")
)

// Store stores filesets. A fileset is a path -> index relationship
// All filesets exist in the same keyspace and can be merged by prefix
type Store interface {
	Set(ctx context.Context, p string, md *Metadata) error
	Get(ctx context.Context, p string) (*Metadata, error)
	Delete(ctx context.Context, p string) error
	Walk(ctx context.Context, prefix string, cb func(string) error) error
}

func StoreTestSuite(t *testing.T, withStore func(func(Store))) {
	ctx := context.Background()
	t.Run("SetGet", func(t *testing.T) {
		withStore(func(x Store) {
			md := &Metadata{}
			require.NoError(t, x.Set(ctx, "test", md))
			actual, err := x.Get(ctx, "test")
			require.NoError(t, err)
			require.Equal(t, md, actual)
		})
	})
	t.Run("Delete", func(t *testing.T) {
		withStore(func(x Store) {
			require.NoError(t, x.Delete(ctx, "keys that don't exist should not cause delete to error"))
			md := &Metadata{}
			require.NoError(t, x.Set(ctx, "test", md))
			require.NoError(t, x.Delete(ctx, "test"))
			_, err := x.Get(ctx, "test")
			require.Equal(t, ErrPathNotExists, err)
		})
	})
}

func copyPath(ctx context.Context, src, dst Store, srcPath, dstPath string) error {
	md, err := src.Get(ctx, srcPath)
	if err != nil {
		return err
	}
	return dst.Set(ctx, dstPath, md)
}
