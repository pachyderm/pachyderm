//go:build linux

package driver

import (
	"runtime"
	"syscall"
	"unsafe"
)

const _P_PID = 1

// blockUntilWaitable attempts to block until a call to Wait4(pid) will succeed immediately, and
// reports whether it has done so.  It does not actually call Wait4.  Incorrect results will be
// returned if pid has not been started at the time of the call.  This code is modified slightly
// from os/wait_waitid.go in the Go source code.
func blockUntilWaitable(pid int) (bool, error) {
	// The waitid system call expects a pointer to a siginfo_t,
	// which is 128 bytes on all Linux systems.
	// On darwin/amd64, it requires 104 bytes.
	// We don't care about the values it returns.
	var siginfo [16]uint64
	psig := &siginfo[0]
	var e syscall.Errno
	for {
		_, _, e = syscall.Syscall6(syscall.SYS_WAITID, _P_PID, uintptr(pid), uintptr(unsafe.Pointer(psig)), syscall.WEXITED|syscall.WNOWAIT, 0, 0)
		if e != syscall.EINTR {
			break
		}
	}
	runtime.KeepAlive(pid)
	if e != 0 {
		if e == syscall.ECHILD {
			// This means there is no process with that ID, so we can assume it's gone.
			return true, e
		}
		return false, e
	}
	return true, nil
}
