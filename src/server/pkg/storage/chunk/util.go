package chunk

import (
	"math/rand"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/dbutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/tracker"
)

// WithLocalStorage creates a local storage instance for testing during the lifetime of
// the callback.
func WithTestStorage(t testing.TB, f func(obj.Client, *Storage) error, opts ...StorageOption) {
	dbutil.WithTestDB(t, func(db *sqlx.DB) {
		tracker.WithTestTracker(t, db, func(tracker tracker.Tracker) {
			db.MustExec(schema)
			mdstore := NewPostgresStore(db)
			require.NoError(t, obj.WithLocalClient(func(objClient obj.Client) error {
				return f(objClient, NewStorage(objClient, mdstore, tracker, opts...))
			}))
		})
	})
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
