package stream

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
)

// fromForEach provides functionality for generic imperative iteration.
// TODO: Move file set merge and datum merge to this abstraction.
type forEach[T any] struct {
	dataChan chan T
	errChan  chan error
	copy     func(dst, src *T)
}

// NewFromForEach creates a new iterator from a forEachFunc
// Don't write new code that needs this.
func NewFromForEach[T any](ctx context.Context, cp func(dst, src *T), forEachFunc func(func(T) error) error) Iterator[T] {
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
		copy:     cp,
	}
}

// Next returns the next item and progresses the iterator.
func (i *forEach[T]) Next(ctx context.Context, dst *T) error {
	select {
	case data, more := <-i.dataChan:
		if !more {
			return EOS()
		}
		i.copy(dst, &data)
		return nil
	case err := <-i.errChan:
		return err
	}
}
