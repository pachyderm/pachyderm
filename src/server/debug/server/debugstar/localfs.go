package debugstar

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing/fstest"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

type LocalDumpFS struct {
	Root  string
	files fstest.MapFS
}

var _ DumpFS = (*LocalDumpFS)(nil)

func (fs *LocalDumpFS) Write(file string, f func(w io.Writer) error) error {
	buf := new(bytes.Buffer)
	if err := f(buf); err != nil {
		return errors.Wrap(err, "write.f")
	}
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return errors.Wrap(err, "mkdirall")
	}
	if err := os.WriteFile(file, buf.Bytes(), 0o644); err != nil {
		return errors.Wrap(err, "write file")
	}
	return nil
}
