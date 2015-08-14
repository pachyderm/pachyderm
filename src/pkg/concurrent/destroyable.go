package concurrent

import (
	"errors"
	"sync"
	"sync/atomic"
)

var (
	// ErrAlreadyDestroyed is the error returned when a function is called
	// on a Destroyable that has already been destroyed.
	ErrAlreadyDestroyed = errors.New("concurrent: already destroyed")
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

type destroyable struct {
	cv            *sync.Cond
	destroyed     VolatileBool
	numOperations int32
}

func newDestroyable() *destroyable {
	return &destroyable{
		sync.NewCond(&sync.Mutex{}),
		NewVolatileBool(false),
		0,
	}
}

func (d *destroyable) Destroy() error {
	d.cv.L.Lock()
	if !d.destroyed.CompareAndSwap(false, true) {
		d.cv.L.Unlock()
		return ErrAlreadyDestroyed
	}
	for atomic.LoadInt32(&d.numOperations) > 0 {
		d.cv.Wait()
	}
	d.cv.L.Unlock()
	return nil
}

func (d *destroyable) Do(f func() (interface{}, error)) (interface{}, error) {
	d.cv.L.Lock()
	if d.destroyed.Value() {
		d.cv.L.Unlock()
		return nil, ErrAlreadyDestroyed
	}
	atomic.AddInt32(&d.numOperations, 1)
	d.cv.L.Unlock()
	value, err := f()
	atomic.AddInt32(&d.numOperations, -1)
	d.cv.Signal()
	return value, err
}
