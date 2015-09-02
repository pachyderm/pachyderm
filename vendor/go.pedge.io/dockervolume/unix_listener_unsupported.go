// +build !linux,!freebsd

package dockervolume

import "net"

func newUnixListener(
	volumeDriverName string,
	group string,
	start <-chan struct{},
) (net.Listener, error) {
	return nil, nil
}
