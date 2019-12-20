package cmdutil

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"
)

// IO defines the inputs and outputs for a command.
type IO struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
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
	if err := cmd.Run(); err != nil {
		data, _ := ioutil.ReadAll(debugStderr)
		if len(data) > 0 {
			return fmt.Errorf("%s: %s\n%s", strings.Join(args, " "), err.Error(), string(data))
		}
		return fmt.Errorf("%s: %s", strings.Join(args, " "), err.Error())
	}
	return nil
}
