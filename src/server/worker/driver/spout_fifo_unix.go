// +build !windows

package driver

import (
	"os"
	"path/filepath"
	"syscall"

	"github.com/pachyderm/pachyderm/src/server/worker/common"
)

// Mkfifo does not exist on Windows, so this is left unimplemented there, except for tests
func createSpoutFifo(path string) error {
	return syscall.Mkfifo(path, 0666)
}

// os.Symlink requires additional privileges on windows, so this is left
// unimplemented there, except for tests
func (d *driver) linkData(inputs []*common.Input, dir string) error {
	// Make sure that previously symlinked outputs are removed.
	if err := d.unlinkData(inputs); err != nil {
		return err
	}
	for _, input := range inputs {
		src := filepath.Join(dir, input.Name)
		dst := filepath.Join(d.inputDir, input.Name)
		if err := os.Symlink(src, dst); err != nil {
			return err
		}
	}
	return os.Symlink(filepath.Join(dir, "out"), filepath.Join(d.inputDir, "out"))
}
