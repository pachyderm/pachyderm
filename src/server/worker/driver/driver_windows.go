//go:build windows
// +build windows

package driver

import (
	"context"
	"os"
	"path/filepath"
	"syscall"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/common"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/logs"
)

// Note: these are stubs only meant for tests - the worker does not run on windows

func makeSysProcAttr(uid *uint32, gid *uint32) *syscall.SysProcAttr {
	return nil
}

func makeProcessGroupKiller(rctx context.Context, l logs.TaggedLogger, p *os.Process) func() {
	return func() {
		l.Logf("warning: not killing user code's children: unsupported OS")
	}
}

// WithActiveData is implemented differently in unix vs windows because of how
// symlinks work on windows. Here, we move inputs into place before the
// callback, then move them back to the scratch space before returning.
func (d *driver) WithActiveData(inputs []*common.Input, dir string, cb func() error) (retErr error) {
	d.activeDataMutex.Lock()
	defer d.activeDataMutex.Unlock()

	if err := d.moveData(inputs, dir); err != nil {
		return errors.Wrap(err, "error when linking active data directory")
	}
	defer func() {
		if err := d.unmoveData(inputs, dir); err != nil && retErr == nil {
			retErr = errors.Wrap(err, "error when unlinking active data directory")
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

	// rename env file
	src := filepath.Join(dir, common.EnvFileName)
	dst := filepath.Join(d.InputDir(), common.EnvFileName)
	if err := os.Rename(src, dst); err != nil {
		return errors.EnsureStack(err)
	}

	// sometimes for group inputs, this part may get run multiple times for the same file
	seen := make(map[string]bool)
	for _, input := range inputs {
		if _, ok := seen[input.Name]; !ok {
			seen[input.Name] = true
			src := filepath.Join(dir, input.Name)
			dst := filepath.Join(d.InputDir(), input.Name)
			if err := os.Rename(src, dst); err != nil {
				return errors.EnsureStack(err)
			}
		}
	}

	err := os.Rename(filepath.Join(dir, "out"), filepath.Join(d.InputDir(), "out"))
	return errors.EnsureStack(err)
}

func (d *driver) unmoveData(inputs []*common.Input, dir string) error {
	entries, err := os.ReadDir(d.InputDir())
	if err != nil {
		return errors.Wrap(err, "os.ReadDir")
	}
	for _, entry := range entries {
		if entry.Name() == client.PPSScratchSpace {
			continue // don't delete scratch space
		}
		if err := os.Rename(filepath.Join(d.InputDir(), entry.Name()), filepath.Join(dir, entry.Name())); err != nil {
			return errors.EnsureStack(err)
		}
	}
	return nil
}

func printRusage(ctx context.Context, state *os.ProcessState) {
	if state == nil {
		log.Info(ctx, "no process state information after user code exited")
		return
	}
	m.Set(ctx, "cpu_time_seconds", state.UserTime().Seconds()+state.SystemTime().Seconds())
}
