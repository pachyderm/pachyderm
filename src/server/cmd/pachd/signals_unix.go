//go:build unix

package main

import "golang.org/x/sys/unix"

func init() {
	// use unix.SIG{TERM,INT} instead of syscall.SIG{TERM,INT} because syscall is deprecated
	notifySignals = append(notifySignals, unix.SIGTERM, unix.SIGINT)
}
