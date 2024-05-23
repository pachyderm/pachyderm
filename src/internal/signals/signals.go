// Package signals implements cross-platform signal-handling.
package signals

import "os"

// TerminationSignals contains all signals which may terminate a process.  It
// may be passed to signal.NotifyContext or signal.Notify.
var TerminationSignals = []os.Signal{
	os.Interrupt,
	os.Kill, // not notifiable, but included for completeness
}
