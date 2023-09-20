package stream

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/stream/heap"
)

// Merged is the type of elements emitted by the Merger Iterator
type Merged[T any] struct {
	// Values is a slice of values, which are not ordered relative to one another.
	Values []T
	// Indexes is the index of the stream that produced each value.
	Indexes []int
}

func (m *Merged[T]) First() (T, int) {
	return m.Values[0], m.Indexes[0]
}

func (m *Merged[T]) Last() (T, int) {
	l := len(m.Values)
	return m.Values[l-1], m.Indexes[l-1]
}

// CopyMerged efficiently copies a Merged from src to dst.
func CopyMerged[T any](dst, src *Merged[T]) {
	dst.Indexes = append(dst.Indexes[:0], src.Indexes...)
	dst.Values = append(dst.Values[:0], src.Values...)
}

var (
	_ Iterator[Merged[struct{}]] = &Merger[struct{}]{}
)

type mergeEntry[T any] struct {
	it    Peekable[T]
	index int

	peek T
}

type Merger[T any] struct {
	its  []Peekable[T]
	heap heap.Heap[*mergeEntry[T]]
	lt   func(a, b T) bool

	isSetup bool
}

// NewMerger creates an iterator which aggregates entries which are equal in value.
// The entries will come out in ascending order.
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
				return a.index < b.index
			}
		}),
	}
	return m
}

func (m *Merger[T]) Next(ctx context.Context, dst *Merged[T]) error {
	if err := m.ensureSetup(ctx); err != nil {
		return err
	}
	dst.Values = dst.Values[:0]
	dst.Indexes = dst.Indexes[:0]

	// read into dst
	me, exists := m.heap.Pop()
	if !exists {
		return EOS()
	}
	dst.Indexes = append(dst.Indexes, me.index)
	if err := appendNext[T](ctx, me.it, &dst.Values); err != nil {
		return err // any error is an error, since we already peaked.
	}
	// need to put back the stream, read into me.peek for comparison in the heap.
	if err := me.it.Peek(ctx, &me.peek); err != nil && !IsEOS(err) {
		return err
	} else if !IsEOS(err) {
		m.heap.Push(me)
	}

	// get equal elements from other iterators
	for {
		me, exists := m.heap.Pop()
		if !exists {
			break
		}
		if m.lt(dst.Values[0], me.peek) {
			m.heap.Push(me)
			break
		}
		dst.Indexes = append(dst.Indexes, me.index)
		if err := appendNext[T](ctx, me.it, &dst.Values); err != nil {
			return err
		}
		if err := me.it.Peek(ctx, &me.peek); err != nil {
			if IsEOS(err) {
				continue
			}
			return err
		}
		m.heap.Push(me)
	}
	return nil
}

func (m *Merger[T]) ensureSetup(ctx context.Context) error {
	if !m.isSetup {
		for i := range m.its {
			me := &mergeEntry[T]{
				it:    m.its[i],
				index: i,
			}
			if err := m.its[i].Peek(ctx, &me.peek); err != nil {
				if IsEOS(err) {
					continue
				}
				return err
			}
			m.heap.Push(me)
		}
		m.isSetup = true
	}
	return nil
}
