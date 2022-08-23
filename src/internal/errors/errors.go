package errors

import (
	"io"
	"runtime"

	"github.com/pkg/errors"
)

var (
	// New returns an error with the supplied message.
	// New also records the stack trace at the point it was called.
	New = errors.New
	// Errorf formats according to a format specifier and returns the string
	// as a value that satisfies error.
	// Errorf also records the stack trace at the point it was called.
	Errorf = errors.Errorf
	// Unwrap returns the underlying wrapped error if it exists, or nil otherwise.
	Unwrap = errors.Unwrap
	// Is reports whether any error in err's chain matches target. An error is
	// considered to match a target if it is equal to that target or if it
	// implements a method `Is(error) bool` such that `Is(target)` returns true.
	Is = errors.Is
	// Wrap returns an error annotating err with a stack trace
	// at the point Wrap is called, and the supplied message.
	// If err is nil, Wrap returns nil.
	Wrap = errors.Wrap
	// Wrapf returns an error annotating err with a stack trace
	// at the point Wrapf is called, and the format specifier.
	// If err is nil, Wrapf returns nil.
	Wrapf = errors.Wrapf
	// WithStack annotates err with a stack trace at the point WithStack was called.
	// If err is nil, WithStack returns nil.
	WithStack = errors.WithStack
)

// StackTrace is stack of Frames from innermost (newest) to outermost (oldest).
type StackTrace = errors.StackTrace

// EnsureStack will add a stack onto the given error only if it does not already
// have a stack. If err is nil, EnsureStack returns nil.
func EnsureStack(err error) error {
	if err == nil {
		return nil
	} else if err == io.EOF {
		// io.EOF is considered a sentinel value and should not be wrapped due to dumb
		// language design: https://github.com/golang/go/issues/39155
		return err
	}

	if _, ok := err.(StackTracer); ok {
		return err
	}

	return WithStack(err)
}

// Frame is the type of a StackFrame, it is an alias for errors.Frame.
type Frame struct{ errors.Frame }

// Callers returns an errors.StackTrace for the place at which it's called.
func Callers() errors.StackTrace {
	const depth = 32
	var pcs [depth]uintptr
	// 2 skips runtime.Callers and this function
	n := runtime.Callers(2, pcs[:])
	st := make(errors.StackTrace, n)
	for i, pc := range pcs[0:n] {
		st[i] = errors.Frame(pc)
	}
	return st
}

// StackTracer is an interface for errors that can return stack traces.
// Unfortuantely github.com/pkg/errors makes us define this ourselves rather
// than defining it for us.
type StackTracer interface {
	StackTrace() errors.StackTrace
}

// ForEachStackFrame calls f on each Frame in the StackTrace contained in err.
// If is a wrapper around another error it is repeatedly unwrapped and f is
// called with frames from the stack of the innermost error.
func ForEachStackFrame(err error, f func(Frame)) {
	var st errors.StackTrace
	for err != nil {
		if err, ok := err.(StackTracer); ok {
			st = err.StackTrace()
		}
		err = errors.Unwrap(err)
	}
	if len(st) > 0 {
		for _, frame := range st {
			f(Frame{frame})
		}
	}
}
