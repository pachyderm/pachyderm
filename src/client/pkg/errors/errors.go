package errors

import (
	"runtime"

	"github.com/pkg/errors"
)

var (
	New       = errors.New
	Errorf    = errors.Errorf
	Wrap      = errors.Wrap
	Wrapf     = errors.Wrapf
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
