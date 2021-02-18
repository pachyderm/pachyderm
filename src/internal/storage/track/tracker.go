package track

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

var (
	// ErrObjectExists the object already exists
	ErrObjectExists = errors.Errorf("object exists")
	// ErrDanglingRef the operation would create a dangling reference
	ErrDanglingRef = errors.Errorf("the operation would create a dangling reference")
	// ErrTombstone cannot create object because it is marked as a tombstone
	ErrTombstone = errors.Errorf("cannot create object because it is marked as a tombstone")
	// ErrNotTombstone object cannot be deleted because it is not marked as a tombstone
	ErrNotTombstone = errors.Errorf("object cannot be deleted because it is not marked as a tombstone")
	// ErrSelfReference object cannot reference itself
	ErrSelfReference = errors.Errorf("object cannot reference itself")
)

// NoTTL will cause the object to live forever
const NoTTL = time.Duration(0)

// ExpireNow will expire the object immediately.
const ExpireNow = time.Duration(math.MinInt32)

// Tracker tracks objects and their references to one another.
type Tracker interface {
	// DB returns the database the tracker is using
	DB() *sqlx.DB

	// CreateTx creates an object with id=id, and pointers to everything in pointsTo
	// It errors with ErrObjectExists if the object already exists.  Callers may be able to ignore this.
	// It errors with ErrDanglingRef if any of the elements in pointsTo do not exist
	// It errors with ErrTombstone if the object exists and is marked as a tombstone. Callers should retry until
	// they successfully create the object.
	CreateTx(tx *sqlx.Tx, id string, pointsTo []string, ttl time.Duration) error

	// SetTTLPrefix sets the expiration time to current_time + ttl for all objects with ids starting with prefix
	SetTTLPrefix(ctx context.Context, prefix string, ttl time.Duration) (time.Time, error)

	// GetDownstream gets all objects immediately downstream of (pointed to by) object with id
	GetDownstream(ctx context.Context, id string) ([]string, error)

	// GetUpstream gets all objects immediately upstream of (pointing to) the object with id
	GetUpstream(ctx context.Context, id string) ([]string, error)

	// DeleteTx deletes the object, or returns ErrDanglingRef if it deleting it would create dangling refs.
	// If the id doesn't exist, no error is returned
	DeleteTx(tx *sqlx.Tx, id string) error

	// IterateDeletable calls cb with all the objects objects which are no longer referenced and have expired or are tombstoned
	IterateDeletable(ctx context.Context, cb func(id string) error) error
}

// TestTracker runs a TestSuite to ensure Tracker is properly implemented
func TestTracker(t *testing.T, newTracker func(testing.TB) Tracker) {
	ctx := context.Background()
	type test struct {
		Name string
		F    func(*testing.T, Tracker)
	}
	tests := []test{
		{
			"CreateSingleObject",
			func(t *testing.T, tracker Tracker) {
				require.NoError(t, Create(ctx, tracker, "test-id", []string{}, 0))
			},
		},
		{
			"CreateObjectDanglingRef",
			func(t *testing.T, tracker Tracker) {
				require.Equal(t, ErrDanglingRef, Create(ctx, tracker, "test-id", []string{"none", "of", "these", "exist"}, 0))
			},
		},
		{
			"ErrObjectExists",
			func(t *testing.T, tracker Tracker) {
				require.Nil(t, Create(ctx, tracker, "test-id", []string{}, 0))
				require.Equal(t, ErrObjectExists, Create(ctx, tracker, "test-id", []string{}, 0))
			},
		},
		{
			"CreateMultipleObjects",
			func(t *testing.T, tracker Tracker) {
				require.Nil(t, Create(ctx, tracker, "1", []string{}, 0))
				require.Nil(t, Create(ctx, tracker, "2", []string{}, 0))
				require.Nil(t, Create(ctx, tracker, "3", []string{"1", "2"}, 0))
			},
		},
		{
			"GetReferences",
			func(t *testing.T, tracker Tracker) {
				require.Nil(t, Create(ctx, tracker, "1", []string{}, 0))
				require.Nil(t, Create(ctx, tracker, "2", []string{}, 0))
				require.Nil(t, Create(ctx, tracker, "3", []string{"1", "2"}, 0))

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
				require.NoError(t, Create(ctx, tracker, id, []string{}, ExpireNow))
				require.NoError(t, Delete(ctx, tracker, id))
				require.NoError(t, Delete(ctx, tracker, id))
			},
		},
		{
			"ExpireSingleObject",
			func(t *testing.T, tracker Tracker) {
				require.Nil(t, Create(ctx, tracker, "keep", []string{}, time.Hour))
				require.Nil(t, Create(ctx, tracker, "expire", []string{}, ExpireNow))

				var toExpire []string
				err := tracker.IterateDeletable(ctx, func(id string) error {
					toExpire = append(toExpire, id)
					return nil
				})
				require.NoError(t, err)
				require.ElementsEqual(t, []string{"expire"}, toExpire)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			tr := newTracker(t)
			test.F(t, tr)
		})
	}
}

// NewTestTracker returns a tracker scoped to the lifetime of the test
func NewTestTracker(t testing.TB, db *sqlx.DB) Tracker {
	db.MustExec("CREATE SCHEMA IF NOT EXISTS storage")
	db.MustExec(schema)
	return NewPostgresTracker(db)
}
