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

// Page pages content to whichever pager is defined by the PAGER env-var
// (normally /usr/bin/less). If noop is true it's just a pass through.
func Page(noop bool, out io.Writer, run func(out io.Writer) error) error {
	if noop {
		return run(out)
	}
	var eg errgroup.Group
	r, w := io.Pipe()
	eg.Go(func() error {
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
