package testutil

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/kv"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
)

// NewChunkStorage creates a local storage instance for testing during the lifetime of
// the callback.
func NewChunkStorage(t testing.TB, db *sqlx.DB, tr track.Tracker, opts ...chunk.StorageOption) (obj.Client, *chunk.Storage) {
	objC, _ := NewObjectClient(t)
	db.MustExec(`CREATE SCHEMA IF NOT EXISTS storage`)
	require.NoError(t, dbutil.WithTx(context.Background(), db, chunk.SetupPostgresStoreV0))
	return objC, chunk.NewStorage(objC, kv.NewMemCache(10), db, tr, opts...)
}

// NewFilesetStorage constructs a local storage instance scoped to the lifetime of the test
func NewFilesetStorage(t testing.TB, db *sqlx.DB, tr track.Tracker) *fileset.Storage {
	_, chunks := NewChunkStorage(t, db, tr)
	store := NewFilesetStore(t, db)
	return fileset.NewStorage(store, tr, chunks)
}

// NewFilesetStore returns a Store scoped to the lifetime of the test.
func NewFilesetStore(t testing.TB, db *sqlx.DB) fileset.MetadataStore {
	ctx := context.Background()
	tx := db.MustBegin()
	tx.MustExec(`CREATE SCHEMA IF NOT EXISTS storage`)
	require.NoError(t, fileset.SetupPostgresStoreV0(ctx, tx))
	require.NoError(t, tx.Commit())
	return fileset.NewPostgresStore(db)
}
