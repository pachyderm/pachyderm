// +build windows

package worker

import (
	"fmt"
	"syscall"
)

// Note that these functions are stubs for windows and they are not meant to be used outside of tests

func createSpoutFifo(path string) error {
	return fmt.Errorf("unimplemented on windows")
}

func makeCmdCredentials(uid uint32, gid uint32) *syscall.SysProcAttr {
	return nil
}
