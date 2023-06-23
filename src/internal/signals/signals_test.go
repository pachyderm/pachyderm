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
	var (
		c = make(chan os.Signal)
		g errgroup.Group
	)

	signal.Notify(c, signals.TerminationSignals...)

	g.Go(func() error {
		p, err := os.FindProcess(os.Getpid())
		if err != nil {
			return errors.Wrap(err, "could not find own process")
		}
		p.Signal(os.Interrupt)
		return nil
	})

	g.Go(func() error {
		if got, expected := <-c, os.Interrupt; got != expected {
			return errors.Errorf("unexpected signal %v (expected %v)", got, expected)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		t.Error(err)
	}
}
