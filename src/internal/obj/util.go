package obj

import (
	"context"
	"io"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"golang.org/x/sync/errgroup"
)

// Copy copys an object from src at srcPath to dst at dstPath
func Copy(ctx context.Context, src, dst Client, srcPath, dstPath string) (retErr error) {
	return WithPipe(func(w io.Writer) error {
		return src.Get(ctx, srcPath, w)
	}, func(r io.Reader) error {
		return dst.Put(ctx, dstPath, r)
	})
}

// WithPipe calls rcb with a reader and wcb with a writer
func WithPipe(wcb func(w io.Writer) error, rcb func(r io.Reader) error) error {
	pr, pw := io.Pipe()
	eg := errgroup.Group{}
	eg.Go(func() error {
		if err := wcb(pw); err != nil {
			return pw.CloseWithError(err)
		}
		return pw.Close()
	})
	eg.Go(func() error {
		if err := rcb(pr); err != nil {
			return pr.CloseWithError(err)
		}
		return pr.Close()
	})
	return eg.Wait()
}

// NewTestClient creates a obj.Client which is cleaned up after the test exists
func NewTestClient(t testing.TB) Client {
	dir := t.TempDir()
	objC, err := NewLocalClient(dir)
	require.NoError(t, err)
	return objC
}
