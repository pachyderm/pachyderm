// Package miscutil provides an "Island of Misfit Toys", but for helper functions
package miscutil

import (
	"io"

	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
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
	return errors.EnsureStack(eg.Wait())
}
