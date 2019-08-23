// Package erronce provides a sync-Once-like type that can return an error
package erronce

import "sync"

type ErrOnce struct {
	once sync.Once
	err  error
}

func (o *ErrOnce) Do(f func() error) error {
	o.once.Do(func() {
		o.err = f()
	})
	return o.err
}
