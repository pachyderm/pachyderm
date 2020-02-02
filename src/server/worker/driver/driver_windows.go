// +build windows

package driver

import (
	"os"
	"syscall"

	"github.com/otiai10/copy"
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
func (d *driver) linkData(dir string) error {
	// Make sure that previous symlink is removed.
	if err := d.unlinkData(); err != nil {
		return err
	}
	return copy.Copy(dir, d.InputDir())
}
