//go:build unix
// +build unix

package driver

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/meters"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/common"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/logs"
)

func makeSysProcAttr(uid *uint32, gid *uint32) *syscall.SysProcAttr {
	attr := &syscall.SysProcAttr{
		// We create a process group so that we can later send SIGKILL to all child
		// processes created by the user code.
		Setpgid: true,
	}
	if uid != nil && gid != nil {
		attr.Credential = &syscall.Credential{
			Uid:         *uid,
			Gid:         *gid,
			NoSetGroups: true,
		}
	}
	return attr
}

// makeProcessGroupKiller creates a background goroutine that kills all processes assosciated with
// the provided process group when the root context expires or the returned cancellation function is
// called.  Even though we pass a negative number to syscall.Kill, pgid should be positive.
//
// For the driver specifically, we always run this no matter what.  If the context times out, we
// kill all subprocesses including the parent (the non-zero exit there will fail the job, as
// expected).  If user code exits 0, but has child processes in the background, we still kill those,
// but the job will succeed.  If user code wants to fail when its children hang, it will have to
// implement that logic itself (by calling waitpid on the children; "wait" in bash).
func makeProcessGroupKiller(rctx context.Context, l logs.TaggedLogger, pgid int) func() {
	ctx, c := pctx.WithCancel(rctx)
	go func() {
		<-ctx.Done()
		logRunningProcesses(l, pgid)
		if err := syscall.Kill(-pgid, syscall.SIGKILL); err != nil {
			// ESRCH means that the process group is gone, because everything
			// has already exited on its own (or there was only one process in
			// the process group, and it's already gone).
			if err == syscall.ESRCH {
				return
			}
			l.Logf("warning: problem killing user code process group #%v: %v", pgid, err)
		}
	}()
	return c
}

// WithActiveData is implemented differently in unix vs windows because of how
// symlinks work on windows. Here, we create symlinks to the scratch space
// directory, then clean up before returning.
func (d *driver) WithActiveData(inputs []*common.Input, dir string, cb func() error) (retErr error) {
	d.activeDataMutex.Lock()
	defer d.activeDataMutex.Unlock()

	// If a custom user is set, make sure the directory and its content are owned by them.
	if d.uid != nil && d.gid != nil {
		if err := filepath.Walk(dir, func(name string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			return errors.EnsureStack(os.Chown(name, int(*d.uid), int(*d.gid)))
		}); err != nil {
			return errors.EnsureStack(err)
		}
	}
	if err := d.linkData(inputs, dir); err != nil {
		return errors.Wrap(err, "error when linking active data directory")
	}
	defer func() {
		if !d.PipelineInfo().Details.S3Out {
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
	err := filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if (info.Mode() & os.ModeSymlink) == 0 {
			return nil
		}

		target, err := os.Readlink(path)
		if err != nil {
			return errors.EnsureStack(err)
		}
		if !filepath.IsAbs(target) {
			target, err = filepath.Abs(filepath.Join(filepath.Dir(path), target))
			if err != nil {
				return errors.EnsureStack(err)
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
			return errors.EnsureStack(err)
		}

		// Always overwrite the symlink at this point, in case it's relative
		return errors.EnsureStack(os.Symlink(filepath.Join(target), filepath.Join(path)))
	})
	return errors.EnsureStack(err)
}

func (d *driver) linkData(inputs []*common.Input, dir string) error {
	// Make sure that the previously-symlinked outputs are removed.
	if err := d.unlinkData(inputs); err != nil {
		return err
	}

	// link env file
	src := filepath.Join(dir, common.EnvFileName)
	dst := filepath.Join(d.InputDir(), common.EnvFileName)
	if err := os.Symlink(src, dst); err != nil {
		return errors.EnsureStack(err)
	}

	// sometimes for group inputs, this part may get run multiple times for the same file
	seen := make(map[string]bool)
	for _, input := range inputs {
		if input.Name == "" {
			return errors.New("input does not have a name")
		}
		if _, ok := seen[input.Name]; !ok {
			seen[input.Name] = true
			src := filepath.Join(dir, input.Name)
			dst := filepath.Join(d.InputDir(), input.Name)
			if err := os.Symlink(src, dst); err != nil {
				return errors.EnsureStack(err)
			}
		}
	}

	if !d.PipelineInfo().Details.S3Out {
		if err := os.Symlink(filepath.Join(dir, "out"), filepath.Join(d.InputDir(), "out")); err != nil {
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
	meters.Set(ctx, "cpu_time_seconds", state.UserTime().Seconds()+state.SystemTime().Seconds())
	rusage, ok := state.SysUsage().(*syscall.Rusage)
	if !ok {
		return
	}
	// Maxrss is reported in "kilobytes", which means KiB in the Linux kernel world.  (See
	// getrusage(2).)
	meters.Set(ctx, "resident_memory_bytes", rusage.Maxrss*1024)
}
