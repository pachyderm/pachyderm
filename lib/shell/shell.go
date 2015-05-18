package shell

import (
	"bytes"
	"io"
	"log"
	"os/exec"
	"path"
	"runtime"
	"strings"
)

func RunStderr(c *exec.Cmd) error {
	_, callerFile, callerLine, _ := runtime.Caller(1)
	log.Printf("%15s:%.3d -> %s", path.Base(callerFile), callerLine, strings.Join(c.Args, " "))
	stderr, err := c.StderrPipe()
	if err != nil {
		return err
	}
	err = c.Start()
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(stderr)
	if buf.Len() != 0 {
		log.Print("Command had output on stderr.\n Cmd: ", strings.Join(c.Args, " "), "\nstderr: ", buf)
	}
	return c.Wait()
}

func CallCont(c *exec.Cmd, cont func(io.Reader) error) error {
	_, callerFile, callerLine, _ := runtime.Caller(1)
	log.Printf("%15s:%.3d -> %s", path.Base(callerFile), callerLine, strings.Join(c.Args, " "))
	reader, err := c.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := c.StderrPipe()
	if err != nil {
		return err
	}
	err = c.Start()
	if err != nil {
		return err
	}
	err = cont(reader)
	reader.Close()
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(stderr)
	if buf.Len() != 0 {
		log.Print("Command had output on stderr.\n Cmd: ", strings.Join(c.Args, " "), "\nstderr: ", buf)
	}

	return c.Wait()
}
