// Package miscutil provides an "Island of Misfit Toys", but for helper functions
package miscutil

import (
	"context"
	"io"

	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
)

// WithPipe calls rcb with a reader and wcb with a writer
func WithPipe(wcb func(w io.Writer) error, rcb func(r io.Reader) error) error {
	pr, pw := io.Pipe()
	eg := errgroup.Group{}
	eg.Go(func() error {
		err := wcb(pw)
		pw.CloseWithError(err)
		return errors.EnsureStack(err)
	})
	eg.Go(func() error {
		err := rcb(pr)
		pr.CloseWithError(err)
		return errors.EnsureStack(err)
	})
	if err := eg.Wait(); err != nil {
		return errors.EnsureStack(err)
	}
	return nil
}

// Iterator provides functionality for generic imperative iteration.
// TODO: Move file set merge and datum merge to this abstraction.
// TODO: Improve when we upgrade to a go version with generics.
type Iterator[T any] struct {
	peek     *T
	dataChan chan T
	errChan  chan error
}

// NewIterator creates a new iterator.
func NewIterator[T any](ctx context.Context, iterate func(func(T) error) error) stream.Iterator[T] {
	dataChan := make(chan T)
	errChan := make(chan error, 1)
	go func() {
		if err := iterate(func(data T) error {
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
	return &Iterator[T]{
		dataChan: dataChan,
		errChan:  errChan,
	}
}

// Next returns the next item and progresses the iterator.
func (i *Iterator[T]) Next(ctx context.Context, dst *T) error {
	if i.peek != nil {
		tmp := i.peek
		i.peek = nil
		*dst = *tmp
		return nil
	}
	select {
	case data, more := <-i.dataChan:
		if !more {
			return stream.EOS
		}
		*dst = data
		return nil
	case err := <-i.errChan:
		return err
	}
}
