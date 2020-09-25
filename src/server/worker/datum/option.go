package datum

import (
	"context"
	"time"
)

type SetOption func(*Set)

func WithMetaOutput(ptc PutTarClient) SetOption {
	return func(s *Set) {
		s.metaOutputClient = ptc
	}
}

func WithPFSOutput(ptc PutTarClient) SetOption {
	return func(s *Set) {
		s.pfsOutputClient = ptc
	}
}

func WithStats(stats *Stats) SetOption {
	return func(s *Set) {
		s.stats = stats
	}
}

type DatumOption func(*Datum)

func WithRetry(numRetries int) DatumOption {
	return func(d *Datum) {
		d.numRetries = numRetries
	}
}

func WithRecoveryCallback(cb func(context.Context) error) DatumOption {
	return func(d *Datum) {
		d.recoveryCallback = cb
	}
}

func WithTimeout(timeout time.Duration) DatumOption {
	return func(d *Datum) {
		d.timeout = timeout
	}
}
