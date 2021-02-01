// +build windows

package pfssync

import (
	"io"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

func (d *downloader) makePipe(path string, cb func(io.Writer) error) error {
	return errors.Errorf("lazy file download through pipes is not supported on Windows")
}
