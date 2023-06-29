package kv

import (
	"context"
	"io"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"

	"gocloud.dev/blob"
	"gocloud.dev/gcerrors"
)

type ObjectStore2 struct {
	b                        *blob.Bucket
	maxKeySize, maxValueSize int
}

func NewFromBucket(b *blob.Bucket, maxKeySize, maxValueSize int) *ObjectStore2 {
	return &ObjectStore2{
		b:            b,
		maxKeySize:   maxKeySize,
		maxValueSize: maxValueSize,
	}
}

func (s *ObjectStore2) Get(ctx context.Context, key []byte, buf []byte) (int, error) {
	ctx, cf := context.WithCancel(ctx)
	defer cf()
	r, err := s.b.NewReader(ctx, string(key), nil)
	if err != nil {
		return 0, err
	}
	defer r.Close()
	return miscutil.ReadInto(buf, r)
}

func (s *ObjectStore2) Put(ctx context.Context, key []byte, value []byte) error {
	if len(key) > s.maxKeySize {
		return errors.Errorf("max key size %d exceeded. len(key)=%d", s.maxKeySize, len(key))
	}
	if len(value) > s.maxValueSize {
		return errors.Errorf("max value size %d exceeded. len(value)=%d", s.maxKeySize, len(value))
	}
	ctx, cf := context.WithCancel(ctx)
	defer cf()
	w, err := s.b.NewWriter(ctx, string(key), &blob.WriterOptions{
		MaxConcurrency: 1,
		BufferSize:     s.maxValueSize,
	})
	if err != nil {
		return err
	}
	if _, err := w.Write(value); err != nil {
		return err
	}
	return w.Close()
}

func (s *ObjectStore2) Delete(ctx context.Context, key []byte) error {
	err := s.b.Delete(ctx, string(key))
	if gcerrors.Code(err) == gcerrors.NotFound {
		err = nil
	}
	return err
}

func (s *ObjectStore2) Exists(ctx context.Context, key []byte) (bool, error) {
	return s.b.Exists(ctx, string(key))
}

func (s *ObjectStore2) NewKeyIterator(span Span) stream.Iterator[[]byte] {
	it1 := s.b.List(&blob.ListOptions{})
	return &objIterator{it: it1, span: span}
}

type objIterator struct {
	it   *blob.ListIterator
	span Span
}

func (it *objIterator) Next(ctx context.Context, dst *[]byte) error {
	if it.span.End != nil || it.span.Begin != nil {
		// TODO: we might be able to do this efficiently by enumerating prefixes in a span.
		return errors.New("kv.ObjectStore2: span iteration not supported")
	}
	x, err := it.it.Next(ctx)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return stream.EOS()
		}
		return err
	}
	*dst = append((*dst)[:0], x.Key...)
	return nil
}
