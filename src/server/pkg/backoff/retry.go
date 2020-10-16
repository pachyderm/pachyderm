package backoff

import (
	"context"
	"errors"
	"time"

	log "github.com/sirupsen/logrus"
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
			return ctx.Err()
		default:
			log.Errorf("error in %s: %v: retrying in: %v", name, err, d)
		}
		return nil
	}
}

var Continue = errors.New("looping through backoff")

// NotifyContinue is a convenience function for use with RetryUntilCancel. If
// 'inner' is set to a Notify function, it's called if 'err' is anything other
// than Continue. If 'inner' is a string or any other type, any error other than
// Continue is logged, and the loop is re-run (in this case, there is no way to
// escape the backoff--RetryUntilCancel must be used to avoid an infinite loop).
//
// This is useful for e.g. monitoring functions that want to repeatedly execute
// the same control loop until their context is cancelled.
func NotifyContinue(inner interface{}) Notify {
	return func(err error, d time.Duration) error {
		if errors.Is(err, Continue) {
			return nil
		}
		switch n := inner.(type) {
		case Notify:
			return n(err, d) // fallthrough doesn't work for type switches
		case func(error, time.Duration) error:
			return n(err, d)
		default:
			log.Errorf("error in %v: %v (retrying in: %v) (%T)", n, err, d, n)
		}
		return nil
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
		if err = operation(); err == nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if next = b.NextBackOff(); next == Stop {
			return err
		}

		if notify != nil {
			if err := notify(err, next); err != nil {
				return err
			}
		}

		select {
		case <-time.After(next):
			if ctx.Err() != nil {
				// Corner case: if 'b' is ZeroBackOff, and 'ctx' is cancelled inside of
				// notify(), then it's random which case is chosen, and the cancellation
				// won't necessarily prevent operation() from running again. We can't
				// guarantee that 'operation' is called with a live context, but this
				// protects against this corner case.
				return ctx.Err()
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
