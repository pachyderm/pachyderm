package stream

import (
	"context"
	"errors"
)

// Ordered enforces ascending order or returns an error
type Ordered[T any] struct {
	inner      Iterator[T]
	lt         func(a, b T) bool
	lastExists bool
	last       T
	err        error
}

func NewOrdered[T any](inner Iterator[T], lt func(a, b T) bool) *Ordered[T] {
	return &Ordered[T]{
		inner: inner,
		lt:    lt,
	}
}

func (o *Ordered[T]) Next(ctx context.Context, dst *T) error {
	if o.err != nil {
		return o.err
	}
	if err := o.inner.Next(ctx, dst); err != nil {
		return err
	}
	if o.lastExists && !o.lt(o.last, *dst) {
		o.err = errors.New("stream is unordered")
	}
	o.last = *dst
	o.lastExists = true
	return nil
}
