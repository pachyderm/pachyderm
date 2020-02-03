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

func makeCmdCredentials(uid uint32, gid uint32) *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: uid,
			Gid: gid,
		},
	}
}

// WithActiveData is implemented differently in unix vs windows because of how
// symlinks work on windows. Here, we create symlinks to the scratch space
// directory, then clean up before returning.
func (d *driver) WithActiveData(inputs []*common.Input, dir string, cb func() error) (retErr error) {
	d.activeDataMutex.Lock()
	defer d.activeDataMutex.Unlock()

	if err := d.linkData(inputs, dir); err != nil {
		return fmt.Errorf("error when linking active data directory: %v", err)
	}
	defer func() {
		if err := d.unlinkData(inputs); err != nil && retErr == nil {
			retErr = fmt.Errorf("error when unlinking active data directory: %v", err)
		}
	}()

	return cb()
}

func (d *driver) linkData(inputs []*Input, dir string) error {
	// Make sure that the previously-symlinked outputs are removed.
	if err := d.unlinkData(inputs); err != nil {
		return err
	}

	for _, input := range inputs {
		src := filepath.Join(dir, input.Name)
		dst := filepath.Join(d.InputDir(), input.Name)
		if err := os.Symlink(src, dst); err != nil {
			return err
		}
	}

	if d.PipelineInfo().Spout != nil && d.PipelineInfo().Spout.Marker != "" {
		if err = os.Symlink(
			filepath.Join(dir, d.PipelineInfo().Spout.Marker),
			filepath.Join(d.InputDir(), d.PipelineInfo().Spout.Marker),
		); err != nil {
			return err
		}
	}

	return os.Symlink(filepath.Join(dir, "out"), filepath.Join(d.InputDir(), "out"))
}
