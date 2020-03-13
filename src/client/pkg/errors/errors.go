package errors

import (
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
