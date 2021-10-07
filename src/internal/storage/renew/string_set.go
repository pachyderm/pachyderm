package renew

import (
	"context"
	"sync"
	"time"
)

// RenewFunc is the type of a function used to renew a string
type RenewFunc func(ctx context.Context, x string, ttl time.Duration) error

// ComposeFunc is the type of a function used to compose a set of strings
type ComposeFunc func(ctx context.Context, xs []string, ttl time.Duration) (string, error)

// StringSet renews a set of strings until it is closed
type StringSet struct {
	*Renewer
	ttl         time.Duration
	mu          sync.Mutex
	strings     [][]string
	composeFunc ComposeFunc
}

// NewStringSet returns a StringSet it will renew every string in the set for ttl each period.
// See Renewer
func NewStringSet(ctx context.Context, ttl time.Duration, renewFunc RenewFunc, composeFunc ComposeFunc) *StringSet {
	ss := &StringSet{
		ttl:         ttl,
		strings:     [][]string{[]string{}},
		composeFunc: composeFunc,
	}
	ss.Renewer = NewRenewer(ctx, ttl, func(ctx context.Context, ttl time.Duration) error {
		ss.mu.Lock()
		defer ss.mu.Unlock()
		for _, strings := range ss.strings {
			for _, s := range strings {
				if err := renewFunc(ctx, s, ttl); err != nil {
					return err
				}
			}
		}
		return nil
	})
	return ss
}

// Add adds x to the set of strings being renewed
func (ss *StringSet) Add(ctx context.Context, x string) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.strings[0] = append(ss.strings[0], x)
	for i := 0; len(ss.strings[i]) > 100; i++ {
		id, err := ss.composeFunc(ctx, ss.strings[i], ss.ttl)
		if err != nil {
			return err
		}
		ss.strings[i] = []string{}
		if i == len(ss.strings)-1 {
			ss.strings = append(ss.strings, []string{})
		}
		ss.strings[i+1] = append(ss.strings[i+1], id)
	}
	return nil
}

// WithStringSet creates a StringSet using ttl and rf. It calls cb with the StringSets context and the new StringSet.
// If ctx is cancelled, the StringSet will be Closed, and the cancellation will propagate down to the context passed to cb.
func WithStringSet(ctx context.Context, ttl time.Duration, rf RenewFunc, cf ComposeFunc, cb func(ctx context.Context, ss *StringSet) error) (retErr error) {
	ss := NewStringSet(ctx, ttl, rf, cf)
	defer func() {
		if err := ss.Close(); retErr == nil {
			retErr = err
		}
	}()
	return cb(ss.Context(), ss)
}
