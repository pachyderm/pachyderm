package kv

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pacherr"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
	"go.uber.org/zap"
)

type objectAdapter struct {
	objC obj.Client
}

// NewFromObjectClient converts an object client into a key value store.
// This can provide more natural interface for small values, but it will read the entire object into memory
func NewFromObjectClient(objC obj.Client) Store {
	return &objectAdapter{
		objC: objC,
	}
}

func (s *objectAdapter) Put(ctx context.Context, key, value []byte) error {
	return s.retry(ctx, func() error {
		return errors.EnsureStack(s.objC.Put(ctx, string(key), bytes.NewReader(value)))
	})
}

func (s *objectAdapter) Get(ctx context.Context, key []byte, buf []byte) (int, error) {
	// TODO: limitWriter
	sw := &sliceWriter{buf: buf}
	if err := s.retry(ctx, func() error {
		return s.objC.Get(ctx, string(key), sw)
	}); err != nil {
		return 0, err
	}
	return sw.Len(), nil
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

func (s *objectAdapter) NewKeyIterator(span Span) stream.Iterator[[]byte] {
	panic("not implemented")
}

func (s *objectAdapter) retry(ctx context.Context, cb func() error) error {
	return backoff.RetryUntilCancel(ctx, cb, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		if errors.As(err, &pacherr.TransientError{}) {
			log.Info(ctx, "transient error in object adapter; retrying", zap.Error(err), zap.Duration("retryAfter", d))
			return nil
		}
		return err
	})
}

type sliceWriter struct {
	n   int
	buf []byte
}

func (sw *sliceWriter) Write(data []byte) (int, error) {
	if sw.n+len(data) > len(sw.buf) {
		return 0, io.ErrShortBuffer
	}
	n := copy(sw.buf[sw.n:], data)
	sw.n += n
	return n, nil
}

func (sw *sliceWriter) Len() int {
	return sw.n
}
