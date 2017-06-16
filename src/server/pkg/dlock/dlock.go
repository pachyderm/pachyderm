// Package dlock implements a distributed lock on top of etcd.
package dlock

import (
	"context"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
)

// DLock is a handle to a distributed lock.
type DLock interface {
	// Lock acquries the distributed lock, blocking if necessary.  If
	// the lock is acquired, it returns a context that should be used
	// in any subsequent blocking requests, so that if you lose the lock,
	// the requests get cancelled correctly.
	Lock(context.Context) (context.Context, error)
	// Unlock releases the distributed lock.
	Unlock() error
}

type etcdImpl struct {
	client *etcd.Client
	prefix string

	session *concurrency.Session
	mutex   *concurrency.Mutex
	ctx     context.Context
}

// NewDLock attempts to acquire a distributed lock that locks a given prefix
// in the data store.
func NewDLock(client *etcd.Client, prefix string) DLock {
	return &etcdImpl{
		client: client,
		prefix: prefix,
	}
}

func (d *etcdImpl) Lock(ctx context.Context) (context.Context, error) {
	session, err := concurrency.NewSession(d.client, concurrency.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-ctx.Done():
		case <-session.Done():
			cancel()
		}
	}()

	mutex := concurrency.NewMutex(session, d.prefix)
	if err := mutex.Lock(ctx); err != nil {
		return nil, err
	}

	d.session = session
	d.mutex = mutex
	d.ctx = ctx
	return ctx, nil
}

func (d *etcdImpl) Unlock() error {
	if err := d.mutex.Unlock(d.ctx); err != nil {
		return err
	}
	return d.session.Close()
}

func (d *etcdImpl) Context() context.Context {
	return d.ctx
}
