// Package dlock implements a distributed lock on top of etcd.
package dlock

import (
	"context"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
)

// DLock is a handle to a distributed lock.
type DLock interface {
	// Unlock releases the distributed lock.
	Unlock() error
	// Context returns a context that's cancelled if the lock has been
	// released for any reason, e.g. if the node holding the lock has
	// disconnected from etcd.
	// After you've acquire the lock, you should use this context in any
	// subsequent blocking requests, so that if you lose the lock, the
	// requests get cancelled correctly.
	Context() context.Context
}

type etcdImpl struct {
	client *etcd.Client
	prefix string

	session *concurrency.Session
	mutex   *concurrency.Mutex

	ctx context.Context
}

// NewDLock attempts to acquire a distributed lock that locks a given prefix
// in the data store.
func NewDLock(ctx context.Context, client *etcd.Client, prefix string) (DLock, error) {
	session, err := concurrency.NewSession(client, concurrency.WithContext(ctx))
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

	mutex := concurrency.NewMutex(session, prefix)
	if err := mutex.Lock(ctx); err != nil {
		return nil, err
	}

	return &etcdImpl{
		client:  client,
		prefix:  prefix,
		session: session,
		mutex:   mutex,
		ctx:     ctx,
	}, nil
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
