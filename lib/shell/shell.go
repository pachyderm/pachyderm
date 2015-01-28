package shell

import (
	"bytes"
	"log"
	"os/exec"
	"strings"
)

func RunStderr(c *exec.Cmd) error {
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
