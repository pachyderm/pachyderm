// Package miscutil provides an "Island of Misfit Toys", but for helper functions
package miscutil

import (
	"io"

	"golang.org/x/sync/errgroup"
)

// WithPipe calls rcb with a reader and wcb with a writer
func WithPipe(wcb func(w io.Writer) error, rcb func(r io.Reader) error) error {
	pr, pw := io.Pipe()
	eg := errgroup.Group{}
	eg.Go(func() error {
		err := wcb(pw)
		pw.CloseWithError(err)
		return err
	})
	eg.Go(func() error {
		err := rcb(pr)
		pr.CloseWithError(err)
		return pr.Close()
	})
	return eg.Wait()
}
