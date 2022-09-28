package datum

import (
	"context"
	"fmt"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/client"
)

// SetOption configures a set.
type SetOption func(*Set)

// WithMetaOutput sets the Client for the meta output.
func WithMetaOutput(mf client.ModifyFile) SetOption {
	return func(s *Set) {
		s.metaOutputClient = mf
	}
}

// WithPFSOutput sets the Client for the pfs output.
func WithPFSOutput(mf client.ModifyFile) SetOption {
	return func(s *Set) {
		s.pfsOutputClient = mf
	}
}

// WithStats sets the stats to fill in.
func WithStats(stats *Stats) SetOption {
	return func(s *Set) {
		s.stats = stats
	}
}

// Option configures a datum.
type Option func(*Datum)

// WithRetry sets the number of retries.
func WithRetry(numRetries int) Option {
	return func(d *Datum) {
		d.numRetries = numRetries
	}
}

// WithRecoveryCallback sets the recovery callback.
func WithRecoveryCallback(cb func(context.Context) error) Option {
	return func(d *Datum) {
		d.recoveryCallback = cb
	}
}

// WithTimeout sets the timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(d *Datum) {
		d.timeout = timeout
	}
}

// WithPrefixIndex prefixes the datum directory name (both locally and in PFS) with its index value.
func WithPrefixIndex() Option {
	return func(d *Datum) {
		d.IDPrefix = fmt.Sprintf("%032d", d.meta.Index) + "-"
	}
}
