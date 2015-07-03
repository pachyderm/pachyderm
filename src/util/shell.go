package util

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
	"path"
	"runtime"
	"strings"

	"github.com/pachyderm/pachyderm/src/log"
)

func RunStderr(c *exec.Cmd) error {
	return callCont(c, nil)
}

func CallCont(c *exec.Cmd, cont func(io.Reader) error) error {
	if cont == nil {
		return errors.New("cont was nil")
	}
	return callCont(c, cont)
}

func callCont(c *exec.Cmd, cont func(io.Reader) error) error {
	_, callerFile, callerLine, _ := runtime.Caller(1)
	log.Printf("%15s:%.3d -> %s", path.Base(callerFile), callerLine, strings.Join(c.Args, " "))
	var reader io.Reader
	var err error
	if cont != nil {
		reader, err = c.StdoutPipe()
		if err != nil {
			return err
		}
	}
	stderr, err := c.StderrPipe()
	if err != nil {
		return err
	}
	if err = c.Start(); err != nil {
		return err
	}
	if cont != nil {
		if err = cont(reader); err != nil {
			return err
		}
	}
	buffer := bytes.NewBuffer(nil)
	buffer.ReadFrom(stderr)
	if buffer.Len() != 0 {
		log.Print("Command had output on stderr.\n Cmd: ", strings.Join(c.Args, " "), "\nstderr: ", buffer)
	}
	return c.Wait()
}
