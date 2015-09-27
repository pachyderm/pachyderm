package executil

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"
	"sync"

	"go.pedge.io/protolog"
)

var (
	debug = true
	lock  = &sync.Mutex{}
)

func SetDebug(d bool) {
	// sort of unnecessary but hey, standards
	lock.Lock()
	defer lock.Unlock()
	debug = d
	protolog.SetLogger(protolog.GlobalLogger().AtLevel(protolog.Level_LEVEL_DEBUG))
}

type RunOptions struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func Run(args ...string) error {
	return RunWithOptions(RunOptions{}, args...)
}

func RunStdin(stdin io.Reader, args ...string) error {
	return RunWithOptions(RunOptions{stdin: stdin}, args...)
}

func RunStdout(stdout io.Writer, args ...string) error {
	return RunWithOptions(RunOptions{stdout: stdout}, args...)
}

func RunWithOptions(runOptions RunOptions, args ...string) error {
	if len(args) == 0 {
		return errors.New("run called with no args")
	}
	var debugStderr io.ReadWriter
	stderr := runOptions.stderr
	if debug {
		debugStderr = bytes.NewBuffer(nil)
		if stderr == nil {
			stderr = debugStderr
		} else {
			stderr = io.MultiWriter(stderr, debugStderr)
		}
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = runOptions.stdin
	cmd.Stdout = runOptions.stdout
	cmd.Stderr = stderr
	argsString := strings.Join(args, " ")
	protolog.Debug(&RunningCommand{Args: argsString})
	if err := cmd.Run(); err != nil {
		if debugStderr != nil {
			data, _ := ioutil.ReadAll(debugStderr)
			if data != nil && len(data) > 0 {
				return fmt.Errorf("%s: %s\n\t%s", argsString, err.Error(), string(data))
			}
		}
		return fmt.Errorf("%s: %s", argsString, err.Error())
	}
	return nil
}
