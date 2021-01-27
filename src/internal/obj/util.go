package obj

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

// NewTestClient creates a Client which is cleaned up after the test exists
func NewTestClient(t testing.TB) Client {
	dirBase := path.Join(os.TempDir(), "pachyderm_test")
	require.NoError(t, os.MkdirAll(dirBase, 0700))
	dir, err := ioutil.TempDir(dirBase, "")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.RemoveAll(dir))
	})
	objC, err := NewLocalClient(dir)
	require.Nil(t, err)
	return objC
}

// WithLocalClient constructs a local object storage client for testing during the lifetime of
// the callback.
// DEPRECATED: WithLocalClient implements a testing pattern deprecated since go 1.14. consider switching to NewTestClient
// TODO: delete this function
func WithLocalClient(f func(objC Client) error) (retErr error) {
	dirBase := path.Join(os.TempDir(), "pachyderm_test")
	if err := os.MkdirAll(dirBase, 0700); err != nil {
		return err
	}
	dir, err := ioutil.TempDir(dirBase, "")
	if err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(dir); retErr == nil {
			retErr = err
		}
	}()
	objC, err := NewLocalClient(dir)
	if err != nil {
		return err
	}
	return f(objC)
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
