package signals

import "golang.org/x/sys/unix"

func init() {
	TerminationSignals = append(TerminationSignals, unix.SIGTERM, unix.SIGQUIT, unix.SIGHUP)
}
