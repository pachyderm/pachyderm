package directory

import (
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/zeebo/blake3"
)

func Hash(path string) ([]byte, error) {
	hash := blake3.New()
	dir := os.DirFS(path)
	err := fs.WalkDir(dir, ".", func(path string, d fs.DirEntry, err error) (retErr error) {
		// TODO: Handle gitignore.
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			fmt.Fprintf(hash, "%v mode=%v\n", path, uint32(d.Type()))
			return nil
		}
		fh, err := dir.Open(path)
		if err != nil {
			return errors.Wrapf(err, "open %v", path)
		}
		defer errors.Close(&retErr, fh, "close %v", path)
		if n, err := io.Copy(hash, fh); err != nil {
			return errors.Wrapf(err, "read content")
		} else if n == 0 {
			fmt.Fprintf(hash, "empty file %v mode=%v\n", path, uint32(d.Type()))
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "walk")
	}
	return hash.Sum(nil), nil
}
