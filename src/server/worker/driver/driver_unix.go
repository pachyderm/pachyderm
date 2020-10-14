// +build !windows

package driver

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
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
		return errors.Wrap(err, "error when linking active data directory")
	}
	defer func() {
		if !d.PipelineInfo().S3Out {
			if err := d.rewriteSymlinks(dir); err != nil && retErr == nil {
				retErr = errors.Wrap(err, "error when redirecting symlinks in the active data directory")
			}
		}
		if err := d.unlinkData(inputs); err != nil && retErr == nil {
			retErr = errors.Wrap(err, "error when unlinking active data directory")
		}
	}()

	return cb()
}

// When deactivating a data directory, there may be active symlinks from the
// output dir to an input dir. The paths used in these symlinks may be
// invalidated when we deactivate the output directory, so walk the output
// directory and rewrite any such links.
func (d *driver) rewriteSymlinks(scratchSubdir string) error {
	outputDir := filepath.Join(scratchSubdir, "out")
	inputDirFields := strings.Split(filepath.Clean(d.InputDir()), string(filepath.Separator))
	return filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if (info.Mode() & os.ModeSymlink) == 0 {
			return nil
		}

		target, err := os.Readlink(path)
		if err != nil {
			return err
		}
		if !filepath.IsAbs(target) {
			target, err = filepath.Abs(filepath.Join(filepath.Dir(path), target))
			if err != nil {
				return err
			}
		}

		// Filter out any symlinks that aren't pointing to files in the active data
		targetFields := strings.Split(filepath.Clean(target), string(filepath.Separator))
		for i, dirname := range inputDirFields {
			if targetFields[i] != dirname {
				return nil
			}
		}

		// If it's not pointing into the scratch space, we need to change the target
		if targetFields[len(inputDirFields)] != client.PPSScratchSpace {
			targetFields = append([]string{scratchSubdir}, targetFields[len(inputDirFields):]...)
			target = filepath.Join(targetFields...)
		}

		if err := os.Remove(path); err != nil {
			return err
		}

		// Always overwrite the symlink at this point, in case it's relative or some dumb shit
		return os.Symlink(filepath.Join(target), filepath.Join(path))
	})
}

func (d *driver) linkData(inputs []*common.Input, dir string) error {
	// Make sure that the previously-symlinked outputs are removed.
	if err := d.unlinkData(inputs); err != nil {
		return err
	}

	for _, input := range inputs {
		if input.S3 {
			continue // S3 data is not downloaded
		}
		if input.Name == "" {
			return errors.New("input does not have a name")
		}

		src := filepath.Join(dir, input.Name)
		dst := filepath.Join(d.InputDir(), input.Name)

		if err := os.Symlink(src, dst); err != nil {
			return err
		}
	}

	if d.PipelineInfo().Spout != nil && d.PipelineInfo().Spout.Marker != "" {
		if err := os.Symlink(
			filepath.Join(dir, d.PipelineInfo().Spout.Marker),
			filepath.Join(d.InputDir(), d.PipelineInfo().Spout.Marker),
		); err != nil {
			return err
		}
	}

	if !d.PipelineInfo().S3Out {
		if err := os.Symlink(filepath.Join(dir, "out"), filepath.Join(d.InputDir(), "out")); err != nil {
			return err
		}
	}

	return nil
}
