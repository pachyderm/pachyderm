// Package miscutil provides an "Island of Misfit Toys", but for helper functions
package miscutil

import (
	"context"
	"io"

	"github.com/hashicorp/golang-lru/v2/simplelru"
	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
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
type Iterator struct {
	peek     interface{}
	dataChan chan interface{}
	errChan  chan error
}

// NewIterator creates a new iterator.
// DEPRECATED: use stream.NewFromForEach instead
func NewIterator(ctx context.Context, iterate func(func(interface{}) error) error) *Iterator {
	dataChan := make(chan interface{})
	errChan := make(chan error, 1)
	go func() {
		if err := iterate(func(data interface{}) error {
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
	return &Iterator{
		dataChan: dataChan,
		errChan:  errChan,
	}
}

// Next returns the next item and progresses the iterator.
func (i *Iterator) Next() (interface{}, error) {
	if i.peek != nil {
		tmp := i.peek
		i.peek = nil
		return tmp, nil
	}
	select {
	case data, more := <-i.dataChan:
		if !more {
			return nil, io.EOF
		}
		return data, nil
	case err := <-i.errChan:
		return nil, err
	}
}

// CacheFunc caches any function with a single input and output. Uses a LRU with the given size.
// The size defualts to 100 to avoid errors.
func CacheFunc[K comparable, V any](f func(K) V, size int) func(K) V {
	if size <= 0 {
		size = 100
	}
	cache, _ := simplelru.NewLRU[K, V](size, nil)
	return func(a K) V {
		if ent, ok := cache.Get(a); ok {
			return ent
		}
		v := f(a)
		cache.Add(a, v)
		return v
	}
}

// ReadInto reads into dst until the end of the stream.
// If the stream has more than len(dst) bytes, an io.ErrShortBuffer error is returned.
func ReadInto(dst []byte, r io.Reader) (int, error) {
	var n int
	for n < len(dst) {
		n2, err := r.Read(dst[n:])
		if errors.Is(err, io.EOF) || n2 == 0 {
			n += n2
			return n, nil
		}
		if err != nil {
			return 0, errors.EnsureStack(err)
		}
		n += n2
	}
	// If we got here, it means the buffer is full.
	// Now to check if there is anything more to read.
	// If we get io.EOF, and nothing else, then we don't need to error.
	extraBuf := [1]byte{}
	if n2, err := r.Read(extraBuf[:]); errors.Is(err, io.EOF) && n2 == 0 {
		return n, nil
	} else if err != nil {
		return n, errors.EnsureStack(err)
	}
	// We filled the buffer, and there was another byte afterwards, return a short buffer error
	return n, errors.Wrapf(io.ErrShortBuffer, "len(dst)=%d", io.ErrShortBuffer, len(dst))
}
