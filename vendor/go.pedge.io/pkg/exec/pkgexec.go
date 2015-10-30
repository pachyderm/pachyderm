package pkgexec

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"

	"go.pedge.io/protolog"
)

var (
	// ErrNoArgs is the error returned if there are no arguments given.
	ErrNoArgs = errors.New("pkgexec: no arguments given")

	globalDebug = true
	lock        = &sync.Mutex{}
)

// SetDebug sets debug mode for  This will log commands at the debug level, and
// include the output of a command in the resulting error if the command fails.
func SetDebug(debug bool) {
	// sort of unnecessary but hey, standards
	lock.Lock()
	defer lock.Unlock()
	globalDebug = debug
	if debug {
		protolog.SetLogger(protolog.GlobalLogger().AtLevel(protolog.Level_LEVEL_DEBUG))
	}
}

// IO defines the inputs and outputs for a command.
type IO struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// Run runs the command with the given arguments.
func Run(args ...string) error {
	return RunIO(IO{}, args...)
}

// RunStdin runs the command with the given stdin and arguments.
func RunStdin(stdin io.Reader, args ...string) error {
	return RunIO(IO{Stdin: stdin}, args...)
}

// RunStdout runs the command with the given stdout and arguments.
func RunStdout(stdout io.Writer, args ...string) error {
	return RunIO(IO{Stdout: stdout}, args...)
}

// RunStderr runs the command with the given stderr and arguments.
func RunStderr(stderr io.Writer, args ...string) error {
	return RunIO(IO{Stderr: stderr}, args...)
}

// RunIO runs the command with the given IO and arguments.
func RunIO(ioObj IO, args ...string) error {
	if len(args) == 0 {
		return ErrNoArgs
	}
	var debugStderr io.ReadWriter
	stderr := ioObj.Stderr
	if globalDebug {
		debugStderr = bytes.NewBuffer(nil)
		if stderr == nil {
			stderr = debugStderr
		} else {
			stderr = io.MultiWriter(stderr, debugStderr)
		}
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = ioObj.Stdin
	cmd.Stdout = ioObj.Stdout
	cmd.Stderr = stderr
	if globalDebug {
		protolog.Debug(&RunningCommand{Args: strings.Join(args, " ")})
	}
	if err := cmd.Run(); err != nil {
		if debugStderr != nil {
			data, _ := ioutil.ReadAll(debugStderr)
			if data != nil && len(data) > 0 {
				return fmt.Errorf("%s: %s\n\t%s", strings.Join(args, " "), err.Error(), string(data))
			}
		}
		return fmt.Errorf("%s: %s", strings.Join(args, " "), err.Error())
	}
	return nil
}

// Chown changes the filepath to be owned by the current user and group.
// It uses 'sudo chown' and 'sudo chgrp'.
//
// TODO(pedge): how to escalate privileges from go?
func Chown(filePath string) error {
	uid := fmt.Sprintf("%d", os.Getuid())
	gid := fmt.Sprintf("%d", os.Getgid())
	if err := RunIO(IO{Stdout: os.Stderr, Stderr: os.Stderr}, "sudo", "chown", uid, filePath); err != nil {
		return err
	}
	if err := RunIO(IO{Stdout: os.Stderr, Stderr: os.Stderr}, "sudo", "chgrp", gid, filePath); err != nil {
		return err
	}
	return nil
}
