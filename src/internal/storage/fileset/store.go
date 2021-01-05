package fileset

import (
	"context"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
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
	Set(ctx context.Context, id ID, md *Metadata) error
	Get(ctx context.Context, id ID) (*Metadata, error)
	Delete(ctx context.Context, id ID) error
}

// StoreTestSuite is a suite of tests for a Store.
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
}

// func copyPath(ctx context.Context, src, dst Store, srcPath, dstPath string, tracker track.Tracker, ttl time.Duration) error {
// 	md, err := src.Get(ctx, srcPath)
// 	if err != nil {
// 		return err
// 	}
// 	var idxs []*index.Index
// 	switch x := md.Value.(type) {
// 	case *Metadata_Primitive:
// 		idxs = []*index.Index{x.Primitive}
// 	case *Metadata_Composite:
// 		idxs = []*index.Index{x.Composite.}
// 	}
// 	for _, prim := range comp.Layers {
// 		idxs = append(idxs, prim.Additive)
// 		idxs = append(idxs, prim.Deletive)
// 	}
// 	if err := createTrackerObject(ctx, dstPath, idxs, tracker, ttl); err != nil {
// 		return err
// 	}
// 	return dst.Set(ctx, dstPath)
// }
