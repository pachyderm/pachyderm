package obj

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

// Copy copys an object from src at srcPath to dst at dstPath
func Copy(ctx context.Context, src, dst Client, srcPath, dstPath string) (retErr error) {
	return miscutil.WithPipe(func(w io.Writer) error {
		return src.Get(ctx, srcPath, w)
	}, func(r io.Reader) error {
		return dst.Put(ctx, dstPath, r)
	})
}

// NewTestClient creates a obj.Client which is cleaned up after the test exists
func NewTestClient(t testing.TB) (Client, string) {
	dir := t.TempDir()
	objC, err := NewLocalClient(dir)
	require.NoError(t, err)
	return objC, strings.ReplaceAll(strings.Trim(dir, "/"), "/", ".")
}
