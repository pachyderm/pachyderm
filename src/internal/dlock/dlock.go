// Package dlock implements a distributed lock on top of etcd.
package dlock

import (
	"context"

	etcd "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// DLock is a handle to a distributed lock.
type DLock interface {
	// Lock acquries the distributed lock, blocking if necessary.  If
	// the lock is acquired, it returns a context that should be used
	// in any subsequent blocking requests, so that if you lose the lock,
	// the requests get cancelled correctly.
	Lock(context.Context) (context.Context, error)
	// Unlock releases the distributed lock.
	Unlock(context.Context) error
}

type etcdImpl struct {
	client *etcd.Client
	prefix string

	session *concurrency.Session
	mutex   *concurrency.Mutex
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
	// The default TTL is 60 secs which means that if a node dies, it
	// still holds the lock for 60 secs, which is too high.
	session, err := concurrency.NewSession(d.client, concurrency.WithContext(ctx), concurrency.WithTTL(15))
	if err != nil {
		return nil, errors.EnsureStack(err)
	}

	mutex := concurrency.NewMutex(session, d.prefix)
	if err := mutex.Lock(ctx); err != nil {
		return nil, errors.EnsureStack(err)
	}

	ctx, cancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-ctx.Done():
		case <-session.Done():
			cancel()
		}
	}()

	d.session = session
	d.mutex = mutex
	return ctx, nil
}

func (d *etcdImpl) Unlock(ctx context.Context) error {
	if err := d.mutex.Unlock(ctx); err != nil {
		return errors.EnsureStack(err)
	}
	return errors.EnsureStack(d.session.Close())
}
