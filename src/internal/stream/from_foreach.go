package stream

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
)

// fromForEach provides functionality for generic imperative iteration.
// TODO: Move file set merge and datum merge to this abstraction.
// TODO: Improve when we upgrade to a go version with generics.
type forEach[T any] struct {
	peek     *T
	dataChan chan T
	errChan  chan error
}

// NewIterator creates a new iterator from a forEachFunc
// Don't write new code that needs this.
func NewFromForEach[T any](ctx context.Context, forEachFunc func(func(T) error) error) Iterator[T] {
	dataChan := make(chan T)
	errChan := make(chan error, 1)
	go func() {
		if err := forEachFunc(func(data T) error {
			select {
			case dataChan <- data:
				return nil
			case <-ctx.Done():
				return errutil.ErrBreak
			}
		}); err != nil {
			errChan <- err
			return
		}
		close(dataChan)
	}()
	return &forEach[T]{
		dataChan: dataChan,
		errChan:  errChan,
	}
}

// Next returns the next item and progresses the iterator.
func (i *forEach[T]) Next(ctx context.Context, dst *T) error {
	if i.peek != nil {
		tmp := i.peek
		i.peek = nil
		*dst = *tmp
		return nil
	}
	select {
	case data, more := <-i.dataChan:
		if !more {
			return EOS
		}
		*dst = data
		return nil
	case err := <-i.errChan:
		return err
	}
}
