package kv

import (
	"bytes"
	"context"
	"sync"

	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pacherr"
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

func (s *objectAdapter) Put(ctx context.Context, key, value []byte) (retErr error) {
	return s.objC.Put(ctx, string(key), bytes.NewReader(value))
}

func (s *objectAdapter) Get(ctx context.Context, key []byte, cb ValueCallback) (retErr error) {
	return s.withBuffer(func(buf *bytes.Buffer) error {
		if err := s.objC.Get(ctx, string(key), buf); err != nil {
			if pacherr.IsNotExist(err) {
				err = pacherr.NewNotExist("kv", string(key))
			}
			return err
		}
		return cb(buf.Bytes())
	})
}

func (s *objectAdapter) Delete(ctx context.Context, key []byte) error {
	return s.objC.Delete(ctx, string(key))
}

func (s *objectAdapter) Exists(ctx context.Context, key []byte) (bool, error) {
	return s.objC.Exists(ctx, string(key))
}

func (s *objectAdapter) Walk(ctx context.Context, prefix []byte, cb func(key []byte) error) error {
	return s.objC.Walk(ctx, string(prefix), func(p string) error {
		return cb([]byte(p))
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
