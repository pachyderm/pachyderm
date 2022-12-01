//go:build !linux

package driver

import "github.com/pachyderm/pachyderm/v2/src/server/worker/logs"

func logRunningProcesses(l logs.TaggedLogger, pgid int) {
	l.Logf("warning: listing processes on this OS is not supported; you won't see debug information about which child processes we're about to kill")
}
