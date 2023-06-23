package signals

import "syscall"

func init() {
	TerminationSignals = append(TerminationSignals, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
}
