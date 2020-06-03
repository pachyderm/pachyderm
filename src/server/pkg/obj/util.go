package obj

import (
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
