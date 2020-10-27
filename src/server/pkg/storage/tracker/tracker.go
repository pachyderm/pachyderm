package tracker

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/dbutil"
)

var (
	ErrObjectExists  = errors.Errorf("object exists")
	ErrDanglingRef   = errors.Errorf("the operation would create a dangling reference")
	ErrTombstone     = errors.Errorf("cannot create object because it is marked as a tombstone")
	ErrNotTombstone  = errors.Errorf("object cannot be deleted because it is not marked as a tombstone")
	ErrSelfReference = errors.Errorf("object cannot reference itself")
)

// Tracker tracks objects and their references to one another.
type Tracker interface {
	// CreateObject creates an object with id=id, and pointers to everything in pointsTo
	// It errors with ErrObjectExists if the object already exists.  Callers may be able to ignore this.
	// It errors with ErrDanglingRef if any of the elements in pointsTo do not exist
	// It errors with ErrTombstone if the object exists and is marked as a tombstone. Callers should retry until
	// they successfully create the object.
	// CreateObject should always be called *before* adding an object to an auxillary store.
	CreateObject(ctx context.Context, id string, pointsTo []string, ttl time.Duration) error

	// SetTTLPrefix sets the expiration time to current_time + ttl for all objects with ids starting with prefix
	SetTTLPrefix(ctx context.Context, prefix string, ttl time.Duration) (time.Time, error)
	// TODO: thoughts on these?
	// SetTTL(ctx context.Context, id string, ttl time.Duration) error
	// SetTTLBatch(ctx context.Context, ids []string, ttl time.Duration) error

	// GetDownstream gets all objects immediately downstream of (pointed to by) object with id
	GetDownstream(ctx context.Context, id string) ([]string, error)

	// GetUpstream gets all objects immediately upstream of (pointing to) the object with id
	GetUpstream(ctx context.Context, id string) ([]string, error)

	// MarkTombstone causes the object to appear deleted.  It cannot be created, and objects referencing it cannot
	// be created until it is deleted.
	// It errors if id has other objects referencing it.
	// Marking something as a tombstone which is already a tombstone is not an error
	MarkTombstone(ctx context.Context, id string) error

	// FinishDelete deletes the object
	// It is an error to call DeleteObject without calling MarkTombstone.
	FinishDelete(ctx context.Context, id string) error

	// IterateExpired calls cb with all the objects which have expired or are tombstones.
	IterateExpired(ctx context.Context, cb func(id string) error) error
}

// TestTracker runs a TestSuite to ensure Tracker is properly implemented
func TestTracker(t *testing.T, withTracker func(func(Tracker))) {
	ctx := context.Background()
	type test struct {
		Name string
		F    func(*testing.T, Tracker)
	}
	tests := []test{
		{
			"CreateSingleObject",
			func(t *testing.T, tracker Tracker) {
				require.Nil(t, tracker.CreateObject(ctx, "test-id", []string{}, 0))
			},
		},
		{
			"CreateObjectDanglingRef",
			func(t *testing.T, tracker Tracker) {
				require.Equal(t, ErrDanglingRef, tracker.CreateObject(ctx, "test-id", []string{"none", "of", "these", "exist"}, 0))
			},
		},
		{
			"CreateMultipleObjects",
			func(t *testing.T, tracker Tracker) {
				require.Nil(t, tracker.CreateObject(ctx, "1", []string{}, 0))
				require.Nil(t, tracker.CreateObject(ctx, "2", []string{}, 0))
				require.Nil(t, tracker.CreateObject(ctx, "3", []string{"1", "2"}, 0))
			},
		},
		{
			"GetReferences",
			func(t *testing.T, tracker Tracker) {
				require.Nil(t, tracker.CreateObject(ctx, "1", []string{}, 0))
				require.Nil(t, tracker.CreateObject(ctx, "2", []string{}, 0))
				require.Nil(t, tracker.CreateObject(ctx, "3", []string{"1", "2"}, 0))

				dwn, err := tracker.GetDownstream(ctx, "3")
				require.Nil(t, err)
				require.ElementsEqual(t, []string{"1", "2"}, dwn)
				ups, err := tracker.GetUpstream(ctx, "2")
				require.Nil(t, err)
				require.ElementsEqual(t, []string{"3"}, ups)
			},
		},
		{
			"DeleteSingleObject",
			func(t *testing.T, tracker Tracker) {
				id := "test"
				require.Nil(t, tracker.CreateObject(ctx, id, []string{}, 0))
				require.Nil(t, tracker.MarkTombstone(ctx, id))
				require.Nil(t, tracker.MarkTombstone(ctx, id)) // repeat mark tombstones should be allowed
				require.Nil(t, tracker.FinishDelete(ctx, id))
			},
		},
		{
			"ExpireSingleObject",
			func(t *testing.T, tracker Tracker) {
				require.Nil(t, tracker.CreateObject(ctx, "keep", []string{}, time.Hour))
				require.Nil(t, tracker.CreateObject(ctx, "expire", []string{}, time.Microsecond))
				time.Sleep(time.Millisecond)

				var toExpire []string
				tracker.IterateExpired(ctx, func(id string) error {
					toExpire = append(toExpire, id)
					return nil
				})
				require.ElementsEqual(t, []string{"expire"}, toExpire)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			withTracker(func(tracker Tracker) {
				test.F(t, tracker)
			})
		})
	}
}

func WithTestTracker(t testing.TB, cb func(tracker Tracker)) {
	dbutil.WithTestDB(t, func(db *sqlx.DB) {
		db.MustExec("CREATE SCHEMA storage")
		db.MustExec(schema)
		tr := NewPGTracker(db)
		cb(tr)
	})
}
