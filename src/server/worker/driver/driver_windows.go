// +build windows

package driver

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	"github.com/pachyderm/pachyderm/src/client"
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

// WithActiveData is implemented differently in unix vs windows because of how
// symlinks work on windows. Here, we move inputs into place before the
// callback, then move them back to the scratch space before returning.
func (d *driver) WithActiveData(inputs []*common.Input, dir string, cb func() error) (retErr error) {
	d.activeDataMutex.Lock()
	defer d.activeDataMutex.Unlock()

	if err := d.moveData(inputs, dir); err != nil {
		return fmt.Errorf("error when linking active data directory: %v", err)
	}
	defer func() {
		if err := d.unmoveData(inputs, dir); err != nil && retErr == nil {
			retErr = fmt.Errorf("error when unlinking active data directory: %v", err)
		}
	}()

	return cb()
}

// os.Symlink requires additional privileges on windows, so just move the files instead
func (d *driver) moveData(inputs []*common.Input, dir string) error {
	// Make sure that the previous outputs are removed.
	if err := d.unlinkData(inputs); err != nil {
		return err
	}

	for _, input := range inputs {
		if input.S3 {
			continue
		}
		src := filepath.Join(dir, input.Name)
		dst := filepath.Join(d.InputDir(), input.Name)
		if err := os.Rename(src, dst); err != nil {
			return err
		}
	}

	if d.PipelineInfo().Spout != nil && d.PipelineInfo().Spout.Marker != "" {
		src := filepath.Join(dir, d.PipelineInfo().Spout.Marker)
		dst := filepath.Join(d.InputDir(), d.PipelineInfo().Spout.Marker)
		if err := os.Rename(src, dst); err != nil {
			return err
		}
	}

	return os.Rename(filepath.Join(dir, "out"), filepath.Join(d.InputDir(), "out"))
}

func (d *driver) unmoveData(inputs []*common.Input, dir string) error {
	entries, err := ioutil.ReadDir(d.InputDir())
	if err != nil {
		return fmt.Errorf("ioutil.ReadDir: %v", err)
	}
	for _, entry := range entries {
		if entry.Name() == client.PPSScratchSpace {
			continue // don't delete scratch space
		}
		if err := os.Rename(filepath.Join(d.InputDir(), entry.Name()), filepath.Join(dir, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}
