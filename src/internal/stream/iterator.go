package stream

import (
	"context"
	"errors"
	"io"
)

var EOS = errors.New("end of stream")

func IsEOS(err error) bool {
	return errors.Is(err, EOS)
}

type Iterator[T any] interface {
	// Next reads the next element into dst, and advances the iterator.
	// Next returns EOS when the iteration is over, dst will not be affected.
	Next(ctx context.Context, dst *T) error
}

type Peekable[T any] interface {
	Iterator[T]

	// Peek reads the next element into dst, but does not advance the iterator.
	// Peek returns EOS when the iteration is over, dst will not be affected.
	Peek(ctx context.Context, dst *T) error
}

// Next is a convenience function for allocating a T and using the iterator to read into it with it.Next
func Next[T any](ctx context.Context, it Iterator[T]) (ret T, _ error) {
	err := it.Next(ctx, &ret)
	return ret, err
}

// Peek is a convenience function for allocating a T and using the iterator to read into it with it.Peek
func Peek[T any](ctx context.Context, it Peekable[T]) (ret T, _ error) {
	err := it.Peek(ctx, &ret)
	return ret, err
}

// ForEach calls fn with elements from it.  The element passed to fn must not be retained after
// fn has returned.
func ForEach[T any](ctx context.Context, it Iterator[T], fn func(t T) error) error {
	var x T
	for {
		if err := it.Next(ctx, &x); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if err := fn(x); err != nil {
			return err
		}
	}
}

// Read fills buf with elements from the iterator and returns the number copied into buf.
func Read[T any](ctx context.Context, it Iterator[T], buf []T) (n int, _ error) {
	for i := range buf {
		if err := it.Next(ctx, &buf[i]); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

// Collect reads at most max from the iterator into a buffer and returns it.
func Collect[T any](ctx context.Context, it Iterator[T], max int) (ret []T, _ error) {
	err := ForEach(ctx, it, func(x T) error {
		if len(ret) >= max {
			return errors.New("iterator has more than max entries")
		}
		ret = append(ret, x)
		return nil
	})
	return ret, err
}

// Skip discards one element from the iterator.
func Skip[T any](ctx context.Context, it Iterator[T]) error {
	var x T
	return it.Next(ctx, &x)
}

type Slice[T any] struct {
	xs  []T
	pos int
}

func NewSlice[T any](xs []T) *Slice[T] {
	return &Slice[T]{
		xs: xs,
	}
}

func (s *Slice[T]) Next(ctx context.Context, dst *T) error {
	if s.pos >= len(s.xs) {
		return io.EOF
	}
	*dst = s.xs[s.pos]
	s.pos++
	return nil
}

func (s *Slice[T]) Peek(ctx context.Context, dst *T) error {
	if s.pos >= len(s.xs) {
		return io.EOF
	}
	*dst = s.xs[s.pos]
	return nil
}
