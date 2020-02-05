// +build !windows

package server

import (
	"syscall"
)

// Mkfifo does not exist on Windows, so this is left unimplemented there, except for tests
func createSpoutFifo(path string) error {
	return syscall.Mkfifo(path, 0666)
}

func makeCmdCredentials(uid uint32, gid uint32) *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: uid,
			Gid: gid,
		},
	}
}
