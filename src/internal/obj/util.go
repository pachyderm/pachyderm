package obj

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

// NewTestClient creates a Client which is cleaned up after the test exists
func NewTestClient(t testing.TB) (Client, string) {
	dirBase := path.Join(os.TempDir(), "pachyderm_test")
	require.NoError(t, os.MkdirAll(dirBase, 0700))
	dir, err := ioutil.TempDir(dirBase, "")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.RemoveAll(dir))
	})
	objC, err := NewLocalClient(dir)
	require.NoError(t, err)
	return objC, strings.ReplaceAll(strings.Trim(dir, "/"), "/", ".")
}

// Copy copys an object from src at srcPath to dst at dstPath
func Copy(ctx context.Context, src, dst Client, srcPath, dstPath string) (retErr error) {
	rc, err := src.Reader(ctx, srcPath, 0, 0)
	if err != nil {
		return err
	}
	defer func() {
		if err := rc.Close(); retErr == nil {
			retErr = err
		}
	}()
	wc, err := dst.Writer(ctx, dstPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := wc.Close(); retErr == nil {
			retErr = err
		}
	}()
	_, err = io.Copy(wc, rc)
	return err
}
