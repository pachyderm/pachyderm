package chunk

import (
	"context"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/kv"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
)

// NewTestStorage creates a local storage instance for testing during the lifetime of
// the callback.
func NewTestStorage(t testing.TB, db *pachsql.DB, tr track.Tracker, opts ...StorageOption) (obj.Client, *Storage) {
	objC := dockertestenv.NewTestObjClient(t)
	db.MustExec(`CREATE SCHEMA IF NOT EXISTS storage`)
	require.NoError(t, dbutil.WithTx(context.Background(), db, SetupPostgresStoreV0))
	return objC, NewStorage(objC, kv.NewMemCache(10), db, tr, opts...)
}

// FullRef creates a data reference for the full chunk referenced by a data reference.
func FullRef(dataRef *DataRef) *DataRef {
	chunkDataRef := &DataRef{}
	chunkDataRef.Ref = dataRef.Ref
	chunkDataRef.SizeBytes = dataRef.Ref.SizeBytes
	return chunkDataRef
}

func StableDataRefs(dataRefs []*DataRef) bool {
	for i, dataRef := range dataRefs {
		// Edge chunks in the middle are not stable since they were not created by content defined chunking.
		if i != 0 && i != len(dataRefs)-1 && dataRef.Ref.Edge {
			return false
		}
		// Each data reference should refer to the full chunk.
		if dataRef.OffsetBytes != 0 || dataRef.SizeBytes != dataRef.Ref.SizeBytes {
			return false
		}
	}
	return true
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
