package pkgsync //import "go.pedge.io/pkg/sync"

import "errors"

var (
	// ErrAlreadyDestroyed is the error returned when a function is called
	// on a Destroyable that has already been destroyed.
	ErrAlreadyDestroyed = errors.New("pkgsync: already destroyed")
)

// Destroyable is a wrapper for any object that allows an atomic destroy operation
// to be performed, and will monitor if other functions are called after the object
// has already been destroyed.
type Destroyable interface {
	Destroy() error
	Do(func() (interface{}, error)) (interface{}, error)
}

// NewDestroyable creates a new Destroyable.
func NewDestroyable() Destroyable {
	return newDestroyable()
}

// VolatileBool is a volatile bool.
//
// TODO(pedge): Is this even needed? Need to understand go memory model better.
type VolatileBool interface {
	// Get the current value.
	Value() bool
	// Set the value to newBool, and return oldBool == newBool.
	CompareAndSwap(oldBool bool, newBool bool) bool
}

// NewVolatileBool creates a new VolatileBool.
func NewVolatileBool(initialBool bool) VolatileBool {
	return newVolatileBool(initialBool)
}
