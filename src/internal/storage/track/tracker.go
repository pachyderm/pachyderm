package track

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pacherr"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

var (
	// ErrDifferentObjectExists the object already exists, with different downstream objects.
	ErrDifferentObjectExists = errors.Errorf("a different object exists at that id")
	// ErrDanglingRef the operation would create a dangling reference
	ErrDanglingRef = errors.Errorf("the operation would create a dangling reference")
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
	DB() *pachsql.DB

	// CreateTx creates an object with id=id, and pointers to everything in pointsTo
	// It errors with ErrDifferentObjectExists if the object already exists.  Callers may be able to ignore this.
	// It errors with ErrDanglingRef if any of the elements in pointsTo do not exist
	CreateTx(tx *pachsql.Tx, id string, pointsTo []string, ttl time.Duration) error

	// SetTTL sets the expiration time to current_time + ttl for the specified object
	SetTTL(ctx context.Context, id string, ttl time.Duration) (time.Time, error)

	// GetExpiresAt returns the time that the object expires or a pacherr.ErrNotExist if it has expired.
	GetExpiresAt(ctx context.Context, id string) (time.Time, error)

	// GetDownstream gets all objects immediately downstream of (pointed to by) object with id
	GetDownstream(ctx context.Context, id string) ([]string, error)

	// GetUpstream gets all objects immediately upstream of (pointing to) the object with id
	GetUpstream(ctx context.Context, id string) ([]string, error)

	// DeleteTx deletes the object, or returns ErrDanglingRef if deleting it would create dangling refs.
	// If the id doesn't exist, no error is returned
	DeleteTx(tx *pachsql.Tx, id string) error

	// IterateDeletable calls cb with some top-level objects which are no longer referenced and have expired
	// Even if it deletes all top-level objects, there may be more to delete after it runs
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
				// none exist
				require.Equal(t, ErrDanglingRef, Create(ctx, tracker, "test-id", []string{"none", "of", "these", "exist"}, 0))
				// some exist
				require.NoError(t, Create(ctx, tracker, "1", []string{}, 0))
				require.NoError(t, Create(ctx, tracker, "2", []string{}, 0))
				require.Equal(t, ErrDanglingRef, Create(ctx, tracker, "test-id2", []string{"1", "2", "none", "of", "these", "exist"}, 0))
			},
		},
		{
			"ObjectExists",
			func(t *testing.T, tracker Tracker) {
				require.NoError(t, Create(ctx, tracker, "test-id", []string{}, 0))
				require.NoError(t, Create(ctx, tracker, "test-id", []string{}, 0))
			},
		},
		{
			"ErrDifferentObjectExists",
			func(t *testing.T, tracker Tracker) {
				require.NoError(t, Create(ctx, tracker, "1", []string{}, 0))
				require.NoError(t, Create(ctx, tracker, "2", []string{}, 0))

				require.NoError(t, Create(ctx, tracker, "3", []string{"1", "2"}, 0))
				require.NoError(t, Create(ctx, tracker, "3", []string{"1", "2"}, 0))

				err := Create(ctx, tracker, "3", []string{"1"}, 0)
				require.YesError(t, err)
				require.ErrorIs(t, err, ErrDifferentObjectExists)
			},
		},
		{
			"CreateMultipleObjects",
			func(t *testing.T, tracker Tracker) {
				require.NoError(t, Create(ctx, tracker, "1", []string{}, 0))
				require.NoError(t, Create(ctx, tracker, "2", []string{}, 0))
				require.NoError(t, Create(ctx, tracker, "3", []string{"1", "2"}, 0))
			},
		},
		{
			"GetReferences",
			func(t *testing.T, tracker Tracker) {
				require.NoError(t, Create(ctx, tracker, "1", []string{}, 0))
				require.NoError(t, Create(ctx, tracker, "2", []string{}, 0))
				require.NoError(t, Create(ctx, tracker, "3", []string{"1", "2"}, 0))

				dwn, err := tracker.GetDownstream(ctx, "3")
				require.NoError(t, err)
				require.ElementsEqual(t, []string{"1", "2"}, dwn)
				ups, err := tracker.GetUpstream(ctx, "2")
				require.NoError(t, err)
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
				require.NoError(t, Create(ctx, tracker, "keep", []string{}, time.Hour))
				require.NoError(t, Create(ctx, tracker, "expire", []string{}, ExpireNow))

				var toExpire []string
				err := tracker.IterateDeletable(ctx, func(id string) error {
					toExpire = append(toExpire, id)
					return nil
				})
				require.NoError(t, err)
				require.ElementsEqual(t, []string{"expire"}, toExpire)

				runGC(t, tracker)
				shouldNotExist(t, tracker, "expire")
			},
		},
		{
			"ExpireList",
			func(t *testing.T, tracker Tracker) {
				const N = 20
				for i := 0; i < N; i++ {
					var deps []string
					if i > 0 {
						deps = []string{fmt.Sprintf("%04d", i-1)}
					}
					require.NoError(t, Create(ctx, tracker, fmt.Sprintf("%04d", i), deps, ExpireNow))
				}
				require.NoError(t, Create(ctx, tracker, "keep", []string{fmt.Sprintf("%04d", N-1)}, time.Hour))
				require.NoError(t, Create(ctx, tracker, "expire", []string{}, ExpireNow))
				var toExpire []string
				err := tracker.IterateDeletable(ctx, func(id string) error {
					toExpire = append(toExpire, id)
					return nil
				})
				require.NoError(t, err)
				require.ElementsEqual(t, []string{"expire"}, toExpire)

				// get rid of "expire"
				require.Equal(t, 1, runGC(t, tracker))

				// should be no-op
				require.Equal(t, 0, runGC(t, tracker))

				// get rid of "keep"
				_, err = tracker.SetTTL(ctx, "keep", ExpireNow)
				require.NoError(t, err)
				require.Equal(t, 1, runGC(t, tracker))

				// should expire 1 each round for N+1 rounds
				for i := 0; i < N; i++ {
					require.Equal(t, 1, runGC(t, tracker))
				}
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

func shouldNotExist(t *testing.T, tracker Tracker, id string) {
	ctx := context.Background()
	_, err := tracker.GetExpiresAt(ctx, id)
	require.ErrorIs(t, err, pacherr.ErrNotExist{Collection: "tracker", ID: id})
}

func runGC(t *testing.T, tracker Tracker) int {
	ctx := context.Background()
	var count int
	err := tracker.IterateDeletable(ctx, func(id string) error {
		count++
		return Delete(ctx, tracker, id)
	})
	require.NoError(t, err)
	return count
}

// NewTestTracker returns a tracker scoped to the lifetime of the test
func NewTestTracker(t testing.TB, db *pachsql.DB) Tracker {
	db.MustExec("CREATE SCHEMA IF NOT EXISTS storage")
	db.MustExec(schema)
	return NewPostgresTracker(db)
}
