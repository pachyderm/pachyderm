package fsutil

import (
	"io/ioutil"
	"os"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

func WithTmpFile(prefix string, cb func(*os.File) error) (retErr error) {
	f, err := ioutil.TempFile(os.TempDir(), prefix)
	if err != nil {
		return errors.EnsureStack(err)
	}
	defer func() {
		if err := os.Remove(f.Name()); retErr == nil {
			retErr = errors.EnsureStack(err)
		}
		if err := f.Close(); retErr == nil {
			retErr = errors.EnsureStack(err)
		}
	}()
	return cb(f)
}
