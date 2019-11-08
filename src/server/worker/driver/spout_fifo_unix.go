// +build !windows

package driver

import (
	"syscall"
)

func createSpoutFifo(path string) error {
	return syscall.Mkfifo(path, 0666)
}
