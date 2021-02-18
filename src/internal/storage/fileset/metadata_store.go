package fileset

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

var (
	// ErrFileSetExists path already exists
	ErrFileSetExists = errors.Errorf("path already exists")
	// ErrFileSetNotExists path does not exist
	ErrFileSetNotExists = errors.Errorf("path does not exist")
	// ErrNoTTLSet no ttl set on path
	ErrNoTTLSet = errors.Errorf("no ttl set on path")
)

// MetadataStore stores filesets. A fileset is a path -> index relationship
// All filesets exist in the same keyspace and can be merged by prefix
type MetadataStore interface {
	DB() *sqlx.DB
	Set(tx *sqlx.Tx, id ID, md *Metadata) error
	Get(ctx context.Context, id ID) (*Metadata, error)
	Delete(tx *sqlx.Tx, id ID) error
}

// StoreTestSuite is a suite of tests for a Store.
func StoreTestSuite(t *testing.T, newStore func(t testing.TB) MetadataStore) {
	ctx := context.Background()
	t.Run("SetGet", func(t *testing.T) {
		x := newStore(t)
		md := &Metadata{}
		require.NoError(t, setMetadata(ctx, x, "test", md))
		actual, err := x.Get(ctx, "test")
		require.NoError(t, err)
		require.Equal(t, md, actual)
	})
	t.Run("Delete", func(t *testing.T) {
		x := newStore(t)
		require.NoError(t, deleteMetadata(ctx, x, "keys that don't exist should not cause delete to error"))
		md := &Metadata{}
		require.NoError(t, setMetadata(ctx, x, "test", md))
		require.NoError(t, deleteMetadata(ctx, x, "test"))
		_, err := x.Get(ctx, "test")
		require.Equal(t, ErrFileSetNotExists, err)
	})
}

func setMetadata(ctx context.Context, mds MetadataStore, id ID, md *Metadata) error {
	return dbutil.WithTx(ctx, mds.DB(), func(tx *sqlx.Tx) error {
		return mds.Set(tx, id, md)
	})
}

func deleteMetadata(ctx context.Context, mds MetadataStore, id ID) error {
	return dbutil.WithTx(ctx, mds.DB(), func(tx *sqlx.Tx) error {
		return mds.Delete(tx, id)
	})
}
