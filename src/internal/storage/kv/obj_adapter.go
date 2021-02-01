package kv

import (
	"bytes"
	"context"
	"io"
	"sync"

	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
)

type objectAdapter struct {
	objC    obj.Client
	bufPool sync.Pool
}

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
	wc, err := s.objC.Writer(ctx, string(key))
	if err != nil {
		return err
	}
	defer func() {
		if err := wc.Close(); retErr == nil {
			retErr = err
		}
	}()
	if _, err := wc.Write(value); err != nil {
		return err
	}
	return nil
}

func (s *objectAdapter) GetF(ctx context.Context, key []byte, cb ValueCallback) (retErr error) {
	rc, err := s.objC.Reader(ctx, string(key), 0, 0)
	if err != nil {
		if s.objC.IsNotExist(err) {
			err = ErrKeyNotFound
		}
		return err
	}
	defer func() {
		if err := rc.Close(); retErr == nil {
			retErr = err
		}
	}()
	return s.withBuffer(func(buf *bytes.Buffer) error {
		if _, err := io.Copy(buf, rc); err != nil {
			return err
		}
		return AssertNotModified(buf.Bytes(), cb)
	})
}

func (s *objectAdapter) Delete(ctx context.Context, key []byte) error {
	return s.objC.Delete(ctx, string(key))
}

func (s *objectAdapter) Exists(ctx context.Context, key []byte) (bool, error) {
	exists := s.objC.Exists(ctx, string(key))
	return exists, nil
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
