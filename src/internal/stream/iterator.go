package stream

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

type errEndOfStream struct{}

func (e errEndOfStream) Error() string {
	return "end of stream"
}

// EOS returns a new end of stream error
func EOS() error {
	return errors.EnsureStack(errEndOfStream{})
}

// IsEOS returns true if the error is an end of stream error.
func IsEOS(err error) bool {
	return errors.Is(err, errEndOfStream{})
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
			if IsEOS(err) {
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
// End of iteration is signaled by returning (_, EOS)
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
	for {
		if len(ret) > max {
			return nil, errors.Errorf("stream.Collect: iterator produced too many elements. max=%d", max)
		}
		if err := appendNext(ctx, it, &ret); err != nil {
			if IsEOS(err) {
				break
			}
			return nil, err
		}
	}
	return ret, nil
}

// Skip discards one element from the iterator.
func Skip[T any](ctx context.Context, it Iterator[T]) error {
	var x T
	return it.Next(ctx, &x)
}

// Slice is an iterator backed by an in-memory slice
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
		return EOS()
	}
	*dst = s.xs[s.pos]
	s.pos++
	return nil
}

func (s *Slice[T]) Peek(ctx context.Context, dst *T) error {
	if s.pos >= len(s.xs) {
		return EOS()
	}
	*dst = s.xs[s.pos]
	return nil
}

// Reset resets the iterator to the beginning.
func (s *Slice[T]) Reset() {
	s.pos = 0
}

type peekable[T any] struct {
	inner Iterator[T]
	copy  func(dst, src *T)

	peek   T
	exists bool
}

func NewPeekable[T any](it Iterator[T], cp func(dst, src *T)) Peekable[T] {
	if p, ok := it.(Peekable[T]); ok {
		return p
	}
	return &peekable[T]{
		inner: it,
		copy:  cp,
	}
}

func (p *peekable[T]) Next(ctx context.Context, dst *T) error {
	if p.exists {
		p.copy(dst, &p.peek)
		p.exists = false
		return nil
	}
	return p.inner.Next(ctx, dst)
}

func (p *peekable[T]) Peek(ctx context.Context, dst *T) error {
	if !p.exists {
		if err := p.inner.Next(ctx, &p.peek); err != nil {
			return err
		}
		p.exists = true
	}
	p.copy(dst, &p.peek)
	return nil
}

// appendNext appends the next value from it to s
func appendNext[T any](ctx context.Context, it Iterator[T], s *[]T) error {
	var zero T
	*s = append(*s, zero)
	if err := it.Next(ctx, &(*s)[len(*s)-1]); err != nil {
		*s = (*s)[:len(*s)-1]
		return err
	}
	return nil
}
