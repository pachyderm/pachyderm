package track

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
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
		panic("must provide non-zero TTL for track.Rewnewer")
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
		_, err := r.tracker.SetTTLPrefix(ctx, r.id+"/", ttl)
		return err
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
	_, err := r.tracker.SetTTLPrefix(ctx, r.id+"/", ExpireNow)
	return err
}

func (r *Renewer) nextInt() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := r.n
	r.n++
	return n
}
