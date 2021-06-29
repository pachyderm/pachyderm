package pager

import (
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"

	"golang.org/x/sync/errgroup"
)

var (
	defaultPager = []string{"less", "-RFX"}
)

// Page pages content to whichever pager is defined by the PAGER env-var
// (normally /usr/bin/less). If noop is true it's just a pass through.
func Page(noop bool, env cmdutil.Env, run func(out io.Writer) error) error {
	if noop {
		return run(env.Out())
	}
	var eg errgroup.Group
	r, w := io.Pipe()
	eg.Go(func() error {
		return w.CloseWithError(run(w))
	})
	eg.Go(func() error {
		pager := strings.Fields(os.Getenv("PAGER"))
		if len(pager) == 0 {
			pager = defaultPager
		}
		cmd := exec.Command(pager[0], pager[1:]...)
		cmd.Stdin = r
		cmd.Stdout = env.Out()
		cmd.Stderr = env.Err()
		return cmd.Run()
	})
	return eg.Wait()
}
