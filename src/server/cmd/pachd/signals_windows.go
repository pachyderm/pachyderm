//go:build windows
// +build windows

package main

import "golang.org/x/sys/windows"

func init() {
	// use windows.SIG{TERM,INT} instead of syscall.SIG{TERM,INT} because syscall is deprecated
	notifySignals = append(notifySignals, windows.SIGTERM, windows.SIGINT)
}
