package kv

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/meters"
	"github.com/pachyderm/pachyderm/v2/src/internal/pacherr"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
)

var _ Store = &Metered{}

type Metered struct {
	inner Store
	name  string
}

func NewMetered(inner Store, name string) *Metered {
	return &Metered{inner: inner, name: name}
}

func (m *Metered) Put(ctx context.Context, key, value []byte) error {
	ctx = pctx.Child(ctx, m.name)
	if err := m.inner.Put(ctx, key, value); err != nil {
		meters.Inc(ctx, "putErr", 1)
		return err
	}
	meters.Inc(ctx, "put", 1)
	return nil
}

func (m *Metered) Get(ctx context.Context, key, buf []byte) (int, error) {
	ctx = pctx.Child(ctx, m.name)
	n, err := m.inner.Get(ctx, key, buf)
	if err != nil {
		if pacherr.IsNotExist(err) {
			meters.Inc(ctx, "notExist", 1)
		} else {
			meters.Inc(ctx, "getErr", 1)
		}
		return n, err
	}
	meters.Inc(ctx, "get", 1)
	return 0, err
}

func (m *Metered) Delete(ctx context.Context, key []byte) error {
	ctx = pctx.Child(ctx, m.name)
	if err := m.inner.Delete(ctx, key); err != nil {
		meters.Inc(ctx, "deleteErr", 1)
		return err
	}
	meters.Inc(ctx, "delete", 1)
	return nil
}

func (m *Metered) Exists(ctx context.Context, key []byte) (bool, error) {
	ctx = pctx.Child(ctx, m.name)
	exists, err := m.inner.Exists(ctx, key)
	if err != nil {
		meters.Inc(ctx, "existsErr", 1)
		return false, err
	}
	meters.Inc(ctx, "exists", 1)
	return exists, nil
}

func (m *Metered) NewKeyIterator(span Span) stream.Iterator[[]byte] {
	return m.inner.NewKeyIterator(span)
}
