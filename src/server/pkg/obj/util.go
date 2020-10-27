package obj

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"path"
)

// WithLocalClient constructs a local object storage client for testing during the lifetime of
// the callback.
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
