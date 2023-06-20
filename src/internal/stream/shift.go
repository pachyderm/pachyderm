package stream

import "context"

type blueshift[T any] struct {
	inner Iterator[[]T]
	cp    func(dst, src *T)
	buf   []T
	i     int
}

// Blueshift takes an Iterator[[]T] and produces an Iterator[T]
// The inner iterator is called to get a batch, and then the elements of that batch are emitted.
func Blueshift[T any](x Iterator[[]T], cp func(dst, src *T)) Iterator[T] {
	return &blueshift[T]{
		inner: x,
		cp:    cp,
	}
}

func (it *blueshift[T]) Next(ctx context.Context, dst *T) error {
	if it.i >= len(it.buf) {
		if err := it.inner.Next(ctx, &it.buf); err != nil {
			return err
		}
		it.i = 0
	}
	*dst = it.buf[it.i]
	it.i++
	return nil
}
