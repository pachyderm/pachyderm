package chunk

import (
	"math/rand"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/kv"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
)

// NewTestStorage creates a local storage instance for testing during the lifetime of
// the callback.
func NewTestStorage(t testing.TB, db *sqlx.DB, tr track.Tracker, opts ...StorageOption) (obj.Client, *Storage) {
	objC, _ := obj.NewTestClient(t)
	return objC, NewStorage(objC, kv.NewMemCache(10), db, tr, opts...)
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// RandSeq generates a random sequence of data (n is number of bytes)
func RandSeq(n int) []byte {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return []byte(string(b))
}

// Reference creates a data reference for the full chunk referenced by a data reference.
func Reference(dataRef *DataRef) *DataRef {
	chunkDataRef := &DataRef{}
	chunkDataRef.Ref = dataRef.Ref
	chunkDataRef.SizeBytes = dataRef.Ref.SizeBytes
	return chunkDataRef
}
