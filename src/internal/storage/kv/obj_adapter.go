package kv

import (
	"bytes"
	"context"
	"sync"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pacherr"
	log "github.com/sirupsen/logrus"
)

type objectAdapter struct {
	objC    obj.Client
	bufPool sync.Pool
}

// NewFromObjectClient converts an object client into a key value store.
// This can provide more natural interface for small values, but it will read the entire object into memory
func NewFromObjectClient(objC obj.Client) Store {
	return &objectAdapter{
		objC: objC,
		bufPool: sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(nil)
			},
		},
	}
}

func (s *objectAdapter) Put(ctx context.Context, key, value []byte) error {
	return s.retry(ctx, func() error {
		return errors.EnsureStack(s.objC.Put(ctx, string(key), bytes.NewReader(value)))
	})
}

func (s *objectAdapter) Get(ctx context.Context, key []byte, cb ValueCallback) error {
	return s.retry(ctx, func() error {
		return s.withBuffer(func(buf *bytes.Buffer) error {
			if err := s.objC.Get(ctx, string(key), buf); err != nil {
				return errors.EnsureStack(err)
			}
			return cb(buf.Bytes())
		})
	})
}

func (s *objectAdapter) Delete(ctx context.Context, key []byte) error {
	return s.retry(ctx, func() error {
		return errors.EnsureStack(s.objC.Delete(ctx, string(key)))
	})
}

func (s *objectAdapter) Exists(ctx context.Context, key []byte) (bool, error) {
	var res bool
	if err := s.retry(ctx, func() error {
		var err error
		res, err = s.objC.Exists(ctx, string(key))
		return errors.EnsureStack(err)
	}); err != nil {
		return false, err
	}
	return res, nil
}

// TODO: Should a retry be added to this function? That behavior might be strange from a client standpoint.
// Also, we may end up removing this functionality anyways because it has no current use.
func (s *objectAdapter) Walk(ctx context.Context, prefix []byte, cb func(key []byte) error) error {
	return errors.EnsureStack(s.objC.Walk(ctx, string(prefix), func(p string) error {
		return cb([]byte(p))
	}))
}

func (s *objectAdapter) retry(ctx context.Context, cb func() error) error {
	return backoff.RetryUntilCancel(ctx, cb, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		if errors.As(err, &pacherr.TransientError{}) {
			log.Errorf("transient error in object adapter: %v, retrying in %v", err, d)
			return nil
		}
		return err
	})
}

// withBuffer gets a buffer from the pool, calls cb with it, resets it, then returns it to the pool.
func (s *objectAdapter) withBuffer(cb func(*bytes.Buffer) error) error {
	buf := s.bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		s.bufPool.Put(buf)
	}()
	buf.Reset()
	return cb(buf)
}
