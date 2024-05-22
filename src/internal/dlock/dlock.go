// Package dlock implements a distributed lock on top of etcd.
package dlock

import (
	"context"
	"time"

	etcd "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"go.uber.org/zap"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
)

// DLock is a handle to a distributed lock.
type DLock interface {
	// Lock acquries the distributed lock, blocking if necessary.  If
	// the lock is acquired, it returns a context that should be used
	// in any subsequent blocking requests, so that if you lose the lock,
	// the requests get cancelled correctly.
	Lock(context.Context) (context.Context, error)
	// TryLock is like Lock, but returns an error if the lock is already locked.
	TryLock(context.Context) (context.Context, error)
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

func (d *etcdImpl) Lock(ctx context.Context) (_ context.Context, retErr error) {
	ctx = pctx.Child(ctx, "", pctx.WithFields(zap.String("withLock", d.prefix)))
	defer log.Span(ctx, "DLock.Lock")(log.Errorp(&retErr))

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
	start := time.Now()
	log.Debug(ctx, "acquired lock ok")

	ctx, cancel := pctx.WithCancel(pctx.Child(ctx, "", pctx.WithFields(zap.Bool("locked", true))))
	go func() {
		select {
		case <-ctx.Done():
			log.Debug(ctx, "lock's context is done", zap.Error(context.Cause(ctx)), zap.Duration("lockLifetime", time.Since(start)))
		case <-session.Done():
			log.Debug(ctx, "lock's session is done; cancelling associated context", zap.Duration("lockLifetime", time.Since(start)))
			cancel()
		}
	}()

	d.session = session
	d.mutex = mutex
	return ctx, nil
}

func (d *etcdImpl) TryLock(ctx context.Context) (_ context.Context, retErr error) {
	ctx = pctx.Child(ctx, "", pctx.WithFields(zap.String("withLock", d.prefix)))
	defer log.Span(ctx, "DLock.TryLock")(log.Errorp(&retErr))

	session, err := concurrency.NewSession(d.client, concurrency.WithContext(ctx), concurrency.WithTTL(15))
	if err != nil {
		return nil, errors.EnsureStack(err)
	}

	mutex := concurrency.NewMutex(session, d.prefix)
	if err := mutex.TryLock(ctx); err != nil {
		return nil, errors.EnsureStack(err)
	}
	start := time.Now()
	log.Debug(ctx, "acquired lock ok")

	ctx, cancel := pctx.WithCancel(pctx.Child(ctx, "", pctx.WithFields(zap.Bool("locked", true))))
	go func() {
		select {
		case <-ctx.Done():
			log.Debug(ctx, "lock's context is done", zap.Error(context.Cause(ctx)), zap.Duration("lockLifetime", time.Since(start)))
		case <-session.Done():
			log.Debug(ctx, "lock's session is done; cancelling associated context", zap.Duration("lockLifetime", time.Since(start)))
			cancel()
		}
	}()

	d.session = session
	d.mutex = mutex
	return ctx, nil
}

func (d *etcdImpl) Unlock(ctx context.Context) (retErr error) {
	defer log.Span(ctx, "DLock.Unlock", zap.String("prefix", d.prefix))(log.Errorp(&retErr))

	if err := d.mutex.Unlock(ctx); err != nil {
		return errors.EnsureStack(err)
	}
	if err := d.session.Close(); err != nil {
		return errors.EnsureStack(err)
	}
	log.Debug(ctx, "relinquished lock ok", zap.String("prefix", d.prefix))
	return nil
}
