package pager

import (
	"io"
	"os"
	"os/exec"
)

const (
	defaultPager = "less"
)

func Page(in io.Reader, out io.Writer) error {
	pager := os.Getenv("PAGER")
	if pager == "" {
		pager = defaultPager
	}
	cmd := exec.Command(pager)
	cmd.Stdin = in
	cmd.Stdout = out
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
