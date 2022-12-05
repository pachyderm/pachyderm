//go:build linux

package driver

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pachyderm/pachyderm/v2/src/server/worker/logs"
)

// logRunningProcesses will print the cmdline of processes that we'd kill if we killed the provided
// process group.  Looking at these logs will let users know that they might be doing something
// interesting accidentally.
func logRunningProcesses(l logs.TaggedLogger, pgid int) {
	err := filepath.WalkDir("/proc", func(path string, d fs.DirEntry, err error) error {
		if path == "/proc" {
			return nil
		}
		if !d.IsDir() {
			return filepath.SkipDir
		}
		if err != nil {
			return err
		}

		pid, err := strconv.Atoi(d.Name())
		if err != nil {
			// Not a PID directory.
			return filepath.SkipDir
		}
		stat, err := os.ReadFile(filepath.Join(path, "stat"))
		if err != nil {
			l.Logf("warning: unable to read stat: %v", err)
			return filepath.SkipDir
		}
		// From proc(5):
		// /proc/[pid]/stat
		//
		//     Status information about the process.  This is used by ps(1).  It is defined
		//     in the kernel source file fs/proc/array.c.
		//         (1) pid  %d
		//         The process ID.
		//         (2) comm  %s
		//         The  filename of the executable, in parentheses.
		//         ...
		//         (5) pgrp  %d
		//         The process group ID of the process.
		//
		// That man page is 1-indexed.
		var comm string
		statParts := bytes.SplitN(stat, []byte{' '}, 6)
		if len(statParts) > 4 {
			pgrp, err := strconv.Atoi(string(statParts[4]))
			if err != nil {
				l.Logf("warning: unable to parse %v/stat[4]: %v", path, err)
				return filepath.SkipDir
			}
			if pgrp != pgid {
				// Not something we're going to try and kill; ignore.
				return filepath.SkipDir
			}
			comm = string(statParts[1])
		}

		cmdline, err := os.ReadFile(filepath.Join(path, "cmdline"))
		if err != nil {
			l.Logf("warning: unable to read cmdline: %v", err)
		}
		cmdlineParts := bytes.Split(cmdline, []byte{0})

		l.Logf("note: about to kill unexpectedly-remaining subprocess of the user code: pid=%v comm=%v cmdline=%s", pid, comm, bytes.Join(cmdlineParts, []byte{' '}))
		return filepath.SkipDir
	})
	if err != nil {
		l.Logf("warning: unable to walk /proc (to provide debug information about orphaned child processes): %v", err)
	}
}
