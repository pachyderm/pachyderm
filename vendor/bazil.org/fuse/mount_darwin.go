package fuse

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

var errNoAvail = errors.New("no available fuse devices")

var errNotLoaded = errors.New("osxfusefs is not loaded")

func loadOSXFUSE() error {
	cmd := exec.Command("/Library/Filesystems/osxfusefs.fs/Support/load_osxfusefs")
	cmd.Dir = "/"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

func openOSXFUSEDev() (*os.File, error) {
	var f *os.File
	var err error
	for i := uint64(0); ; i++ {
		path := "/dev/osxfuse" + strconv.FormatUint(i, 10)
		f, err = os.OpenFile(path, os.O_RDWR, 0000)
		if os.IsNotExist(err) {
			if i == 0 {
				// not even the first device was found -> fuse is not loaded
				return nil, errNotLoaded
			}

			// we've run out of kernel-provided devices
			return nil, errNoAvail
		}

		if err2, ok := err.(*os.PathError); ok && err2.Err == syscall.EBUSY {
			// try the next one
			continue
		}

		if err != nil {
			return nil, err
		}
		return f, nil
	}
}

func handleMountOSXFUSE(errCh chan<- error) func(line string) (ignore bool) {
	return func(line string) (ignore bool) {
		const (
			noMountpointPrefix = `mount_osxfusefs: `
			noMountpointSuffix = `: No such file or directory`
		)
		if strings.HasPrefix(line, noMountpointPrefix) && strings.HasSuffix(line, noMountpointSuffix) {
			// re-extract it from the error message in case some layer
			// changed the path
			mountpoint := line[len(noMountpointPrefix) : len(line)-len(noMountpointSuffix)]
			err := &MountpointDoesNotExistError{
				Path: mountpoint,
			}
			select {
			case errCh <- err:
				return true
			default:
				// not the first error; fall back to logging it
				return false
			}
		}

		return false
	}
}

// isBoringMountOSXFUSEError returns whether the Wait error is
// uninteresting; exit status 64 is.
func isBoringMountOSXFUSEError(err error) bool {
	if err, ok := err.(*exec.ExitError); ok && err.Exited() {
		if status, ok := err.Sys().(syscall.WaitStatus); ok && status.ExitStatus() == 64 {
			return true
		}
	}
	return false
}

func callMount(dir string, conf *mountConfig, f *os.File, ready chan<- struct{}, errp *error) error {
	bin := "/Library/Filesystems/osxfusefs.fs/Support/mount_osxfusefs"

	for k, v := range conf.options {
		if strings.Contains(k, ",") || strings.Contains(v, ",") {
			// Silly limitation but the mount helper does not
			// understand any escaping. See TestMountOptionCommaError.
			return fmt.Errorf("mount options cannot contain commas on darwin: %q=%q", k, v)
		}
	}
	cmd := exec.Command(
		bin,
		"-o", conf.getOptions(),
		// Tell osxfuse-kext how large our buffer is. It must split
		// writes larger than this into multiple writes.
		//
		// OSXFUSE seems to ignore InitResponse.MaxWrite, and uses
		// this instead.
		"-o", "iosize="+strconv.FormatUint(maxWrite, 10),
		// refers to fd passed in cmd.ExtraFiles
		"3",
		dir,
	)
	cmd.ExtraFiles = []*os.File{f}
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "MOUNT_FUSEFS_CALL_BY_LIB=")
	// TODO this is used for fs typenames etc, let app influence it
	cmd.Env = append(cmd.Env, "MOUNT_FUSEFS_DAEMON_PATH="+bin)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("setting up mount_osxfusefs stderr: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("setting up mount_osxfusefs stderr: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("mount_osxfusefs: %v", err)
	}
	helperErrCh := make(chan error, 1)
	go func() {
		var wg sync.WaitGroup
		wg.Add(2)
		go lineLogger(&wg, "mount helper output", neverIgnoreLine, stdout)
		go lineLogger(&wg, "mount helper error", handleMountOSXFUSE(helperErrCh), stderr)
		wg.Wait()
		if err := cmd.Wait(); err != nil {
			// see if we have a better error to report
			select {
			case helperErr := <-helperErrCh:
				// log the Wait error if it's not what we expected
				if !isBoringMountOSXFUSEError(err) {
					log.Printf("mount helper failed: %v", err)
				}
				// and now return what we grabbed from stderr as the real
				// error
				*errp = helperErr
				close(ready)
				return
			default:
				// nope, fall back to generic message
			}

			*errp = fmt.Errorf("mount_osxfusefs: %v", err)
			close(ready)
			return
		}

		*errp = nil
		close(ready)
	}()
	return nil
}

func mount(dir string, conf *mountConfig, ready chan<- struct{}, errp *error) (*os.File, error) {
	f, err := openOSXFUSEDev()
	if err == errNotLoaded {
		err = loadOSXFUSE()
		if err != nil {
			return nil, err
		}
		// try again
		f, err = openOSXFUSEDev()
	}
	if err != nil {
		return nil, err
	}
	err = callMount(dir, conf, f, ready, errp)
	if err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}
