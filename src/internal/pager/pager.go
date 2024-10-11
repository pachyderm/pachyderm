// Package pager implements paging based on the PAGER environment variable.
package pager

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

var (
	defaultPager = []string{"less", "-RFX"}
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
		return errors.EnsureStack(w.CloseWithError(run(w)))
	})
	eg.Go(func() error {
		pager := strings.Fields(os.Getenv("PAGER"))
		if len(pager) == 0 {
			pager = defaultPager
		}
		cmd := exec.Command(pager[0], pager[1:]...)
		cmd.Stdin = r
		cmd.Stdout = out
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "unable to run pager %q: %v\n", pager[0], err)
			// Fall back to what amounts to `cat`.
			if _, err := io.Copy(out, r); err != nil {
				return errors.EnsureStack(err)
			}
		}
		return nil
	})
	return errors.EnsureStack(eg.Wait())
}
