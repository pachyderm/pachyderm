// +build windows

package driver

import (
	"os"
	"path/filepath"
	"syscall"

	"github.com/otiai10/copy"

	"github.com/pachyderm/pachyderm/src/server/worker/common"
)

// Note: these are stubs only meant for tests - the worker does not run on windows

func makeCmdCredentials(uid uint32, gid uint32) *syscall.SysProcAttr {
	return nil
}

// Note: this function only exists for tests, the real system uses a fifo for
// this (which does not exist in the normal filesystem on Windows)
func createSpoutFifo(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	return file.Close()
}

// os.Symlink requires additional privileges on windows, so just copy the files instead
func (d *driver) linkData(inputs []*common.Input, dir string) error {
	// Make sure that previously symlinked outputs are removed.
	if err := d.unlinkData(inputs); err != nil {
		return err
	}
	for _, input := range inputs {
		src := filepath.Join(dir, input.Name)
		dst := filepath.Join(d.inputDir, input.Name)
		if err := copy.Copy(src, dst); err != nil {
			return err
		}
	}
	return copy.Copy(filepath.Join(dir, "out"), filepath.Join(d.inputDir, "out"))
}
