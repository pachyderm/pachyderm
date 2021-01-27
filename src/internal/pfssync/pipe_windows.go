// +build windows

package pfssync

import (
	"io"

	"github.com/pachyderm/pachyderm/src/internal/errors"
)

func (p *Puller) makePipe(path string, f func(io.Writer) error) error {
	return errors.Errorf("lazy file sync through pipes is not supported on Windows")
}
