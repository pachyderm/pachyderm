// +build !windows

package pfssync

import (
	"io"
	"os"
	"syscall"
)

func (d *downloader) makePipe(path string, cb func(io.Writer) error) error {
	if err := syscall.Mkfifo(path, 0666); err != nil {
		return err
	}
	d.pipes[path] = struct{}{}
	d.eg.Go(func() (retErr error) {
		f, err := os.OpenFile(path, os.O_WRONLY, os.ModeNamedPipe)
		if err != nil {
			return err
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
