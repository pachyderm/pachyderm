package cmdutil

import (
	"bytes"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
)

// IO defines the inputs and outputs for a command.
type IO struct {
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	Environ []string
}

// RunIO runs the command with the given IO and arguments.
func RunIO(ioObj IO, args ...string) error {
	return RunIODirPath(ioObj, "", args...)
}

// RunStdin runs the command with the given stdin and arguments.
func RunStdin(stdin io.Reader, args ...string) error {
	return RunIO(IO{Stdin: stdin}, args...)
}

// RunIODirPath runs the command with the given IO and arguments in the given directory specified by dirPath.
func RunIODirPath(ioObj IO, dirPath string, args ...string) error {
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
	if ioObj.Environ != nil {
		cmd.Env = ioObj.Environ
	}
	if err := cmd.Run(); err != nil {
		stderrContent, _ := ioutil.ReadAll(debugStderr)
		if len(stderrContent) > 0 {
			return errors.Wrapf(err, "%s\nStderr: %s\nError", strings.Join(args, " "), string(stderrContent))
		}
		return errors.Wrapf(err, "%s", strings.Join(args, " "))
	}
	return nil
}
