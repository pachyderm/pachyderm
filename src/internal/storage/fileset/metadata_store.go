package fileset

import (
	"context"
	"fmt"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

var (
	// ErrFileSetExists path already exists
	ErrFileSetExists = fmt.Errorf("path already exists")
	// ErrFileSetNotExists path does not exist
	ErrFileSetNotExists = fmt.Errorf("path does not exist")
	// ErrNoTTLSet no ttl set on path
	ErrNoTTLSet = fmt.Errorf("no ttl set on path")
)

// MetadataStore stores filesets. A fileset is a path -> index relationship
// All filesets exist in the same keyspace and can be merged by prefix
type MetadataStore interface {
	DB() *pachsql.DB
	SetTx(tx *pachsql.Tx, id Token, md *Metadata) error
	Get(ctx context.Context, id Token) (*Metadata, error)
	GetTx(tx *pachsql.Tx, id Token) (*Metadata, error)
	DeleteTx(tx *pachsql.Tx, id Token) error
	Exists(ctx context.Context, id Token) (bool, error)
}

// StoreTestSuite is a suite of tests for a Store.
func StoreTestSuite(t *testing.T, newStore func(t testing.TB) MetadataStore) {
	ctx := context.Background()
	t.Run("SetGet", func(t *testing.T) {
		x := newStore(t)
		md := &Metadata{}
		testID := newToken()
		require.NoError(t, setMetadata(ctx, x, testID, md))
		actual, err := x.Get(ctx, testID)
		require.NoError(t, err)
		require.Equal(t, md, actual)
	})
	t.Run("Delete", func(t *testing.T) {
		x := newStore(t)
		require.NoError(t, deleteMetadata(ctx, x, newToken()))
		md := &Metadata{}
		testID := newToken()
		require.NoError(t, setMetadata(ctx, x, testID, md))
		require.NoError(t, deleteMetadata(ctx, x, testID))
		_, err := x.Get(ctx, testID)
		require.ErrorIs(t, err, ErrFileSetNotExists)
	})
}

func setMetadata(ctx context.Context, mds MetadataStore, id Token, md *Metadata) error {
	return dbutil.WithTx(ctx, mds.DB(), func(ctx context.Context, tx *pachsql.Tx) error {
		return mds.SetTx(tx, id, md)
	})
}

func deleteMetadata(ctx context.Context, mds MetadataStore, id Token) error {
	return dbutil.WithTx(ctx, mds.DB(), func(ctx context.Context, tx *pachsql.Tx) error {
		return mds.DeleteTx(tx, id)
	})
}
