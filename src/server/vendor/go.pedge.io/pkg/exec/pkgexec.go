package pkgexec //import "go.pedge.io/pkg/exec"

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

	"go.pedge.io/lion"
	"go.pedge.io/lion/proto"
)

var (
	// ErrNoArgs is the error returned if there are no arguments given.
	ErrNoArgs = errors.New("pkgexec: no arguments given")

	globalDebug = false
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
		lion.SetLogger(lion.GlobalLogger().AtLevel(lion.LevelDebug))
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

// RunDirPath runs the command with the given arguments in the given directory specified by dirPath.
func RunDirPath(dirPath string, args ...string) error {
	return RunIODirPath(IO{}, dirPath, args...)
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

// RunOutput runs the command with the given arguments and returns the output of stdout.
func RunOutput(args ...string) ([]byte, error) {
	stdout := bytes.NewBuffer(nil)
	err := RunStdout(stdout, args...)
	return stdout.Bytes(), err
}

// RunOutputDirPath runs the command with the given arguments in the given directory and returns the output of stdout.
func RunOutputDirPath(dirPath string, args ...string) ([]byte, error) {
	stdout := bytes.NewBuffer(nil)
	err := RunIODirPath(IO{Stdout: stdout}, dirPath, args...)
	return stdout.Bytes(), err
}

// RunIO runs the command with the given IO and arguments.
func RunIO(ioObj IO, args ...string) error {
	return RunIODirPath(ioObj, "", args...)
}

// RunIODirPath runs the command with the given IO and arguments in the given directory specified by dirPath.
func RunIODirPath(ioObj IO, dirPath string, args ...string) error {
	if len(args) == 0 {
		return ErrNoArgs
	}
	var debugStderr io.ReadWriter = bytes.NewBuffer(nil)
	var stderr io.Writer = debugStderr
	if ioObj.Stderr != nil {
		stderr = io.MultiWriter(debugStderr, ioObj.Stderr)
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = ioObj.Stdin
	cmd.Stdout = ioObj.Stdout
	cmd.Stderr = stderr
	cmd.Dir = dirPath
	if globalDebug {
		protolion.Debug(&RunningCommand{Args: strings.Join(args, " ")})
	}
	if err := cmd.Run(); err != nil {
		data, _ := ioutil.ReadAll(debugStderr)
		if data != nil && len(data) > 0 {
			return fmt.Errorf("%s: %s\n%s", strings.Join(args, " "), err.Error(), string(data))
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
