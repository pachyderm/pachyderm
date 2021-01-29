// +build windows

package pfssync

import (
	"io"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

func (i *importer) makePipe(path string, cb func(io.Writer) error) error {
	return errors.Errorf("lazy file import through pipes is not supported on Windows")
}
