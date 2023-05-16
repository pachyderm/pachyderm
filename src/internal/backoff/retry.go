package backoff

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"go.uber.org/zap"
)

// An Operation is executing by Retry() or RetryNotify().
// The operation will be retried using a backoff policy if it returns an error.
type Operation func() error

// Notify is a notify-on-error function. It receives an operation error and
// backoff delay if the operation failed (with an error).
//
// If the notify function returns an error itself, we stop retrying and return
// the error.
//
// NOTE that if the backoff policy stated to stop retrying,
// the notify function isn't called.
type Notify func(error, time.Duration) error

// NotifyCtx is a convenience function for use with RetryNotify that exits if
// 'ctx' is closed, and otherwise logs the error and retries.
//
// Note that an alternative, if the only goal is to retry until 'ctx' is closed,
// is to use RetryUntilCancel, which will not even call the given 'notify'
// function if its context is cancelled (RetryUntilCancel with notify=nil is
// similar to RetryNotify with NotifyCtx, except that with the former, the
// backoff will be ended early if the ctx is cancelled, whereas the latter will
// sleep for the full backoff duration and then call the operation again).
func NotifyCtx(ctx context.Context, name string) Notify {
	return func(err error, d time.Duration) error {
		select {
		case <-ctx.Done():
			return errors.EnsureStack(context.Cause(ctx))
		default:
			log.Info(ctx, "NotifyCtx: error; retrying", zap.String("name", name), zap.Error(err), zap.Duration("retryAfter", d))
		}
		return nil
	}
}

// ErrContinue is a sentinel error designed to be used with
// Retry/RetryNotify/RetryUntilCancel and NotifyContinue.
//
// Retry/RetryNotify/RetryUntilCancel: if operation() returns ErrContinue,
// RetryUntilCancel will reset its backoff and call notify() (if set) with the
// error.
//
// NotifyContinue: by extension, NotifyContinue always returns nil (causing
// operation() to be retried) when operation() returns ErrContinue (the typical
// usage of ErrContinue).
//
// The two combined allow semantics similar to 'continue' in a regular loop:
// operation() is re-run from the beginning, as a loop's body would be.
var ErrContinue = errors.New("looping through backoff")

// NotifyContinue is a convenience function for use with RetryUntilCancel. If
// 'inner' is set to a Notify function, it's called if 'err' is anything other
// than ErrContinue. If 'inner' is a string or any other non-nil value of another
// type, any error other than ErrContinue is logged, and the loop is re-run (in
// this case, there is no way to escape the backoff--RetryUntilCancel must be
// used to avoid an infinite loop). If 'inner' is nil, any error other than
// ErrContinue is returned.
//
// This is useful for e.g. monitoring functions that want to repeatedly execute
// the same control loop until their context is cancelled.
func NotifyContinue(inner interface{}) Notify {
	return func(err error, d time.Duration) error {
		if errors.Is(err, ErrContinue) {
			return nil
		}
		if inner != nil {
			switch n := inner.(type) {
			case Notify:
				return n(err, d) // fallthrough doesn't work for type switches
			case func(error, time.Duration) error:
				return n(err, d)
			default:
				log.Info(pctx.TODO(), "NotifyContinue: error; retrying", zap.Any("inner", inner), zap.Error(err), zap.Duration("retryAfter", d))
				return nil
			}
		}
		return err
	}
}

// MustLoop is a convenience function for use with RetryUntilCancel. It wraps
// 'operation' in another Operation that returns ErrContinue if the inner
// operation returns nil. If used with RetryUntilCancel and NotifyContinue, this
// guarantees that 'operation'is retried until 'ctx' is cancelled, even if it
// returns nil.
func MustLoop(operation Operation) Operation {
	return func() error {
		switch err := operation(); err {
		case nil:
			return ErrContinue
		default:
			return err
		}
	}
}

// Retry the operation o until it does not return error or BackOff stops.
// o is guaranteed to be run at least once.
// It is the caller's responsibility to reset b after Retry returns.
//
// Retry sleeps the goroutine for the duration returned by BackOff after a
// failed operation returns.
func Retry(o Operation, b BackOff) error { return RetryNotify(o, b, nil) }

// RetryNotify calls notify function with the error and wait duration
// for each failed attempt before sleep.
func RetryNotify(operation Operation, b BackOff, notify Notify) error {
	return RetryUntilCancel(context.Background(), operation, b, notify)
}

// RetryUntilCancel is the same as RetryNotify, except that it will not retry if
// the given context is canceled.
func RetryUntilCancel(ctx context.Context, operation Operation, b BackOff, notify Notify) error {
	var err error
	var next time.Duration

	b.Reset()
	for {
		err = operation()
		if err == nil {
			return nil
		}
		if context.Cause(ctx) != nil {
			return errors.EnsureStack(context.Cause(ctx)) // return if cancel() was called inside operation()
		}
		if errors.Is(err, ErrContinue) {
			b.Reset()
		}

		if next = b.NextBackOff(); next == Stop {
			return err
		}
		if notify != nil {
			if err := notify(err, next); err != nil {
				return err
			}
		}
		if context.Cause(ctx) != nil {
			// return if cancel() was called inside notify() (select may not catch it
			// if 'next' is 0)
			return errors.EnsureStack(context.Cause(ctx))
		}

		select {
		case <-ctx.Done():
			return errors.EnsureStack(context.Cause(ctx)) // break early if ctx is cancelled in another goro
		case <-time.After(next):
		}
	}
}
