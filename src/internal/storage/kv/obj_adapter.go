package kv

import (
	"bytes"
	"context"
	"io"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
)

type objectAdapter struct {
	objC                     obj.Client
	maxKeySize, maxValueSize int
}

// NewFromObjectClient converts an object client into a key value store.
// This can provide more natural interface for small values, but it will read the entire object into memory
func NewFromObjectClient(objC obj.Client, maxKeySize, maxValueSize int) Store {
	return &objectAdapter{
		objC:         objC,
		maxKeySize:   maxKeySize,
		maxValueSize: maxValueSize,
	}
}

func (s *objectAdapter) Put(ctx context.Context, key, value []byte) error {
	if len(key) > s.maxKeySize {
		return errors.Errorf("max key size %d exceeded. len(key)=%d", s.maxKeySize, len(key))
	}
	if len(value) > s.maxValueSize {
		return errors.Errorf("max value size %d exceeded. len(value)=%d", s.maxKeySize, len(value))
	}
	err := s.objC.Put(ctx, string(key), bytes.NewReader(value))
	return errors.EnsureStack(err)
}

func (s *objectAdapter) Get(ctx context.Context, key []byte, buf []byte) (int, error) {
	sw := &sliceWriter{buf: buf}
	if err := s.objC.Get(ctx, string(key), sw); err != nil {
		return 0, errors.EnsureStack(err)
	}
	return sw.Len(), nil
}

func (s *objectAdapter) Delete(ctx context.Context, key []byte) error {
	return errors.EnsureStack(s.objC.Delete(ctx, string(key)))
}

func (s *objectAdapter) Exists(ctx context.Context, key []byte) (bool, error) {
	res, err := s.objC.Exists(ctx, string(key))
	return res, errors.EnsureStack(err)
}

func (s *objectAdapter) NewKeyIterator(span Span) stream.Iterator[[]byte] {
	panic("not implemented")
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
