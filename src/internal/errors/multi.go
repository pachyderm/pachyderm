package errors

import (
	stderr "errors"
	"io"
)

// Join joins the provided errors into a single multierror.  Any nil errors are skipped, and if all
// errors are nil, the return value is nil.
//
// See `go doc errors.Join` for details.
func Join(errs ...error) error {
	return EnsureStack(stderr.Join(errs...))
}

// JoinInto appends err into *into.  into can point to a nil error, and err can also be nil.
func JoinInto(into *error, err error) {
	if into == nil {
		// into can point to a nil error, but into itself can't be nil.
		// This is fine: var err error; JoinInto(&err, whatever)
		panic("misuse of errors.JoinInto: into pointer must not be nil")
	}
	*into = Join(*into, err)
}

// Close enhances a very common pattern with multierrors; adding the error from Close() to the
// return value of the current function:
//
//	 func f(what string) (retErr error) {
//	     x := thing.Open(what)
//	     defer errors.Close(&retErr, x, "close thing %v", what)
//	     return x.Do()
//	}
//
// A non-wrapping version is not available because an error without context is maddening.
func Close(into *error, x io.Closer, msg string, args ...any) {
	if err := x.Close(); err != nil {
		JoinInto(into, Wrapf(err, msg, args...))
	}
}

// Invoke is like Close, but for any function that returns an error.
func Invoke(into *error, f func() error, msg string, args ...any) {
	if err := f(); err != nil {
		JoinInto(into, Wrapf(err, msg, args...))
	}
}

// Invoke1 is like Invoke, but the f can take an argument.
func Invoke1[T any](into *error, f func(T) error, x T, msg string, args ...any) {
	if err := f(x); err != nil {
		JoinInto(into, Wrapf(err, msg, args...))
	}
}
