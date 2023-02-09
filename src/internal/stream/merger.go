package stream

import (
	"context"
	"errors"
	"io"

	"github.com/pachyderm/pachyderm/v2/src/internal/stream/heap"
)

var _ Iterator[struct{}] = &Merger[struct{}]{}

type mergeEntry[T any] struct {
	it       Peekable[T]
	priority int // lower is more important

	peek T
}

type Merger[T any] struct {
	its  []Peekable[T]
	heap heap.Heap[*mergeEntry[T]]
	lt   func(a, b T) bool

	isSetup bool
}

// NewMerger creates an iterator which merges the entries from its into a single iterator.
// The entries will come out in ascending order.
// The iterators slice can be thought of as layers, with higher layers masking the value of lower layers when the entries are equal.
func NewMerger[T any](its []Peekable[T], lt func(a, b T) bool) *Merger[T] {
	m := &Merger[T]{
		its: its,
		lt:  lt,
		heap: heap.New(func(a, b *mergeEntry[T]) bool {
			if lt(a.peek, b.peek) {
				return true
			} else if lt(b.peek, a.peek) {
				return false
			} else {
				return a.priority < b.priority
			}
		}),
	}
	return m
}

func (m *Merger[T]) Next(ctx context.Context, dst *T) error {
	if !m.isSetup {
		for i := range m.its {
			me := &mergeEntry[T]{
				it:       m.its[i],
				priority: len(m.its) - i,
			}
			if err := m.its[i].Peek(ctx, &me.peek); err != nil {
				if errors.Is(err, io.EOF) {
					continue
				}
				return err
			}
			m.heap.Push(me)
		}
		m.isSetup = true
	}

	// read into dst
	me, exists := m.heap.Pop()
	if !exists {
		return io.EOF
	}
	if err := me.it.Next(ctx, dst); err != nil {
		return err // any error is an error, since we already peaked.
	}
	// need to put back the stream, read into me.peek for comparison in the heap.
	if err := me.it.Peek(ctx, &me.peek); err != nil && !errors.Is(err, io.EOF) {
		return err
	} else if !errors.Is(err, io.EOF) {
		m.heap.Push(me)
	}

	// drain equal elements from other iterators
	for {
		me, exists := m.heap.Pop()
		if !exists {
			break
		}
		if m.lt(*dst, me.peek) {
			m.heap.Push(me)
			break
		}
		if err := Skip[T](ctx, me.it); err != nil {
			return err
		}
		if err := me.it.Peek(ctx, &me.peek); err != nil {
			if errors.Is(err, io.EOF) {
				continue
			}
			return err
		}
		m.heap.Push(me)
	}

	return nil
}
