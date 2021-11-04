package track

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/renew"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
)

// TmpTrackerPrefix is the tracker prefix for temporary objects.
const TmpTrackerPrefix = "tmp/"

type tmpDeleter struct{}

// NewTmpDeleter creates a new temporary deleter.
func NewTmpDeleter() Deleter {
	return &tmpDeleter{}
}

func (*tmpDeleter) DeleteTx(tx *sqlx.Tx, _ string) error {
	return nil
}

// Renewer renews a add-only set of objects
type Renewer struct {
	id      string
	tracker Tracker
	ttl     time.Duration
	r       *renew.Renewer

	mu sync.Mutex
	n  int
}

// NewRenewer returns a renewer renewing objects in tracker with ttl
func NewRenewer(tracker Tracker, name string, ttl time.Duration) *Renewer {
	if ttl == 0 {
		panic("must provide non-zero TTL for track.Renewer")
	}
	if name == "" {
		panic("must provide non-empty name for track.Renewer")
	}
	r := &Renewer{
		id:      TmpTrackerPrefix + name + "-" + uuid.NewWithoutDashes(),
		tracker: tracker,
		ttl:     ttl,
	}
	r.r = renew.NewRenewer(context.Background(), ttl, func(ctx context.Context, ttl time.Duration) error {
		_, _, err := r.tracker.SetTTLPrefix(ctx, r.id+"/", ttl)
		return errors.EnsureStack(err)
	})
	return r
}

// Add adds an object to the set of objects being renewed.
func (r *Renewer) Add(ctx context.Context, id string) error {
	n := r.nextInt()
	id2 := fmt.Sprintf("%s/%04d", r.id, n)
	// create an object whos sole purpose is to reference id, and to have a structured name
	// which can be renewed in bulk by prefix
	return Create(ctx, r.tracker, id2, []string{id}, r.ttl)
}

// Close stops the background renewal
func (r *Renewer) Close() (retErr error) {
	defer func() {
		if err := r.r.Close(); retErr == nil {
			retErr = err
		}
	}()
	ctx := context.Background()
	_, n, err := r.tracker.SetTTLPrefix(ctx, r.id+"/", ExpireNow)
	if err != nil {
		return errors.EnsureStack(err)
	}
	// TODO: try to check this earlier. beware of the race when incrementing r.n and creating a new entry.
	if n != r.n {
		return errors.Errorf("renewer prefix has wrong count HAVE: %d WANT: %d", n, r.n)
	}
	return errors.EnsureStack(err)
}

func (r *Renewer) nextInt() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := r.n
	r.n++
	return n
}
