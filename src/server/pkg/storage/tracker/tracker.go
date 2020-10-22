package tracker

import (
	"context"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

var (
	ErrObjectExists = errors.Errorf("object exists")
	ErrDanglingRef  = errors.Errorf("cannot create object because references a non-existant object")
	ErrTombstone    = errors.Errorf("cannot create object because it is marked as a tombstone")
	ErrNotTombstone = errors.Errorf("object cannot be deleted because it is not marked as a tombstone")
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

	// Delete object deletes the object
	// It is an error to call DeleteObject without calling SetTombstone.
	DeleteObject(ctx context.Context, id string) error

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
	// TODO: in memory tracker
	cb(nil)
}
