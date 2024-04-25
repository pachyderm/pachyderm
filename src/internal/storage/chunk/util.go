package chunk

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/kv"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
)

// NewTestStorage creates a local storage instance for testing during the lifetime of
// the callback.
func NewTestStorage(t testing.TB, db *pachsql.DB, tr track.Tracker, opts ...StorageOption) (kv.Store, *Storage) {
	p := filepath.Join(t.TempDir(), "obj-store")
	store := kv.NewFSStore(p, 512, DefaultMaxChunkSize)
	db.MustExec(`CREATE SCHEMA IF NOT EXISTS storage`)
	require.NoError(t, dbutil.WithTx(context.Background(), db, SetupPostgresStoreV0))
	return store, NewStorage(store, db, tr, opts...)
}

// FullRef creates a data reference for the full chunk referenced by a data reference.
func FullRef(dataRef *DataRef) *DataRef {
	chunkDataRef := &DataRef{}
	chunkDataRef.Ref = dataRef.Ref
	chunkDataRef.SizeBytes = dataRef.Ref.SizeBytes
	return chunkDataRef
}

func NewDataRef(chunkRef *DataRef, chunkBytes []byte, offset, size int64) *DataRef {
	dataRef := &DataRef{}
	dataRef.Ref = chunkRef.Ref
	if chunkRef.SizeBytes == size {
		dataRef.Hash = chunkRef.Hash
	} else {
		dataRef.Hash = Hash(chunkBytes[offset : offset+size])
	}
	dataRef.OffsetBytes = offset
	dataRef.SizeBytes = size
	return dataRef
}
