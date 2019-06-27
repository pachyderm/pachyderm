package pager

import (
	"io"
	"os"
	"os/exec"

	"golang.org/x/sync/errgroup"
)

const (
	defaultPager = "less"
)

func Page(noop bool, out io.Writer, run func(out io.Writer) error) error {
	if noop {
		return run(out)
	}
	var eg errgroup.Group
	r, w := io.Pipe()
	eg.Go(func() (retErr error) {
		return w.CloseWithError(run(w))
	})
	eg.Go(func() error {
		pager := os.Getenv("PAGER")
		if pager == "" {
			pager = defaultPager
		}
		cmd := exec.Command(pager)
		cmd.Stdin = r
		cmd.Stdout = out
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})
	return eg.Wait()
}
