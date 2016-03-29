package pkgsync

import (
	"sync"
	"sync/atomic"
)

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
