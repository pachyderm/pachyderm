package stream

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
)

// fromForEach provides functionality for generic imperative iteration.
// TODO: Move file set merge and datum merge to this abstraction.
type forEach[T any] struct {
	dataChan chan forEachMsg[T]
	errChan  chan error
	prev     chan struct{}
}

// NewFromForEach creates a new iterator from a forEachFunc
// Don't write new code that needs this.
func NewFromForEach[T any](ctx context.Context, forEachFunc func(func(T) error) error) Iterator[T] {
	dataChan := make(chan forEachMsg[T])
	errChan := make(chan error, 1)
	go func() {
		if err := forEachFunc(func(data T) error {
			msg := forEachMsg[T]{
				data: data,
				done: make(chan struct{}),
			}
			select {
			case dataChan <- msg:
				<-msg.done // do not allow the iterator to advance until the next call to Next.
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
func (it *forEach[T]) Next(ctx context.Context, dst *T) error {
	if it.prev != nil {
		close(it.prev)
		it.prev = nil
	}
	select {
	case msg, isOpen := <-it.dataChan:
		if !isOpen {
			return EOS
		}
		*dst = msg.data
		it.prev = msg.done
		return nil
	case err := <-it.errChan:
		return err
	}
}

type forEachMsg[T any] struct {
	data T
	done chan struct{}
}
