// +build !windows

package pfssync

import (
	"io"
	"os"
	"syscall"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

func (d *downloader) makePipe(path string, cb func(io.Writer) error) error {
	if err := syscall.Mkfifo(path, 0666); err != nil {
		return errors.EnsureStack(err)
	}
	d.pipes[path] = struct{}{}
	d.eg.Go(func() (retErr error) {
		f, err := os.OpenFile(path, os.O_WRONLY, os.ModeNamedPipe)
		if err != nil {
			return errors.EnsureStack(err)
		}
		defer func() {
			if err := f.Close(); retErr == nil {
				retErr = err
			}
		}()
		if d.done {
			return nil
		}
		return cb(f)
	})
	return nil
}
