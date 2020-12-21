package datum

import (
	"context"
	"time"
)

// SetOption configures a set.
type SetOption func(*Set)

// WithMetaOutput sets the AppendFileClient for the meta output.
func WithMetaOutput(afc AppendFileClient) SetOption {
	return func(s *Set) {
		s.metaOutputClient = afc
	}
}

// WithPFSOutput sets the AppendFileClient for the pfs output.
func WithPFSOutput(afc AppendFileClient) SetOption {
	return func(s *Set) {
		s.pfsOutputClient = afc
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
