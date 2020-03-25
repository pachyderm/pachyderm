// +build windows

package sync

import (
	"io"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
)

func (p *Puller) makePipe(path string, f func(io.Writer) error) error {
	return errors.Errorf("lazy file sync through pipes is not supported on Windows")
}
