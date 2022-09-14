package fsutil

import (
	"os"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

func WithTmpFile(prefix string, cb func(*os.File) error, exts ...string) (retErr error) {
	name := prefix + "-*"
	for _, ext := range exts {
		name += "." + ext
	}
	f, err := os.CreateTemp(os.TempDir(), name)
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
