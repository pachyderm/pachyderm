package renew

import (
	"context"
	"sync"
	"time"
)

// StringFunc is the type of a function used to renew a string
type StringFunc func(ctx context.Context, x string, ttl time.Duration) error

// StringSet renews a set of strings until it is closed
type StringSet struct {
	*Renewer
	mu  sync.Mutex
	set map[string]struct{}
}

// NewStringSet returns a StringSet it will renew every string in the set for ttl each period.
// See Renewer
func NewStringSet(ctx context.Context, ttl time.Duration, renewFunc StringFunc) *StringSet {
	ss := &StringSet{
		set: map[string]struct{}{},
	}
	ss.Renewer = NewRenewer(ctx, ttl, func(ctx context.Context, ttl time.Duration) error {
		ss.mu.Lock()
		defer ss.mu.Unlock()
		for s := range ss.set {
			if err := renewFunc(ctx, s, ttl); err != nil {
				return err
			}
		}
		return nil
	})
	return ss
}

// Add adds x to the set of strings being renewed
func (ss *StringSet) Add(x string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.set[x] = struct{}{}
}

// Remove removes x from the set of strings being renewed
func (ss *StringSet) Remove(x string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	delete(ss.set, x)
}

// WithStringSet creates a StringSet using ttl and rf. It calls cb with the StringSets context and the new StringSet.
// If ctx is cancelled, the StringSet will be Closed, and the cancellation will propagate down to the context passed to cb.
func WithStringSet(ctx context.Context, ttl time.Duration, rf StringFunc, cb func(ctx context.Context, ss *StringSet) error) (retErr error) {
	ss := NewStringSet(ctx, ttl, rf)
	defer func() {
		if err := ss.Close(); retErr == nil {
			retErr = err
		}
	}()
	return cb(ss.Context(), ss)
}
