package kv

import (
	"bytes"
	"context"

	"go.uber.org/zap"
	"golang.org/x/exp/slices"

	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
)

var _ Store = &Prefixed{}

// Prefixed is a Store whichs stores entries beneath a prefix in a backing Store.
type Prefixed struct {
	inner  Store
	prefix []byte
}

func NewPrefixed(inner Store, prefix []byte) *Prefixed {
	if len(prefix) == 0 {
		panic("NewPrefix with empty prefix")
	}
	return &Prefixed{
		inner:  inner,
		prefix: slices.Clone(prefix),
	}
}

func (s *Prefixed) Get(ctx context.Context, key []byte, buf []byte) (int, error) {
	return s.inner.Get(ctx, s.addPrefix(key), buf)
}

func (s *Prefixed) Put(ctx context.Context, key []byte, value []byte) error {
	return s.inner.Put(ctx, s.addPrefix(key), value)
}

func (s *Prefixed) Exists(ctx context.Context, key []byte) (bool, error) {
	return s.inner.Exists(ctx, s.addPrefix(key))
}

func (s *Prefixed) Delete(ctx context.Context, key []byte) error {
	return s.inner.Delete(ctx, s.addPrefix(key))
}

func (s *Prefixed) NewKeyIterator(span Span) stream.Iterator[[]byte] {
	span2 := Span{
		Begin: s.prefix,
		End:   PrefixEnd(s.prefix),
	}
	if span.Begin != nil {
		span2.Begin = s.addPrefix(span.Begin)
	}
	if span.End != nil {
		span2.End = s.addPrefix(span.End)
	}
	return &prefixedIter{inner: s.inner.NewKeyIterator(span2), prefix: s.prefix}
}

type prefixedIter struct {
	inner  stream.Iterator[[]byte]
	prefix []byte
}

func (it *prefixedIter) Next(ctx context.Context, dst *[]byte) error {
	for {
		if err := it.inner.Next(ctx, dst); err != nil {
			return err
		}
		if bytes.HasPrefix(*dst, it.prefix) {
			*dst = bytes.TrimPrefix(*dst, it.prefix)
			return nil
		}
		log.Error(ctx, "key does not have prefix", zap.ByteString("key", *dst), zap.ByteString("prefix", it.prefix))
	}
}

func (s *Prefixed) addPrefix(key []byte) (ret []byte) {
	ret = append(ret, s.prefix...)
	ret = append(ret, key...)
	return ret
}
