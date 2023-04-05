//go:build linux

package driver

import (
	"github.com/pachyderm/pachyderm/v2/src/server/worker/logs"
	"github.com/prometheus/procfs"
)

// logRunningProcesses will print the cmdline of processes that we'd kill if we killed the provided
// process group.  Looking at these logs will let users know that they might be doing something
// interesting accidentally.
func logRunningProcesses(l logs.TaggedLogger, pgid int) {
	fs, err := procfs.NewFS("/proc")
	if err != nil {
		l.Logf("warning: unable to process /proc (to provide debug information about orphaned child processes): %v", err)
	}
	pp, err := fs.AllProcs()
	if err != nil {
		l.Logf("warning: unable to get stats from /proc (to provide debug information about orphaned child processes): %v", err)
	}
	for _, p := range pp {
		l.Logf("note: about to kill unexpectedly-remaining subprocess of the user code: pid=%v comm=%v cmdline=%s", p.PID, p.Comm, p.CmdLine)
	}
}
