package signals_test

import (
	"fmt"
	"os"
	"testing"

	"golang.org/x/sys/unix"
)

func TestUnixSignals(t *testing.T) {
	for _, s := range []os.Signal{unix.SIGHUP, unix.SIGTERM, unix.SIGQUIT, unix.SIGINT} {
		t.Run(fmt.Sprintf("signal: %v", s), func(t *testing.T) {
			if err := testSignal(s); err != nil {
				t.Error(err)
			}
		})
	}
}
