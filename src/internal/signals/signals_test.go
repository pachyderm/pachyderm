package signals_test

import (
	"os"
	"os/signal"
	"testing"

	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/signals"
)

func TestSignals(t *testing.T) {
	if err := testSignal(os.Interrupt); err != nil {
		t.Error(err)
	}
}

func testSignal(s os.Signal) error {
	var (
		c = make(chan os.Signal, 1)
		g errgroup.Group
	)
	signal.Notify(c, signals.TerminationSignals...)
	defer signal.Stop(c)
	defer close(c)

	g.Go(func() error {
		p, err := os.FindProcess(os.Getpid())
		if err != nil {
			return errors.Wrap(err, "could not find own process")
		}
		return errors.Wrap(p.Signal(s), "could not signal self")
	})

	g.Go(func() error {
		if got, expected := <-c, s; got != expected {
			return errors.Errorf("unexpected signal %v (expected %v)", got, expected)
		}
		return nil
	})

	return errors.Wrapf(g.Wait(), "could not signal & handle %v", s)
}
