// +build linux freebsd

package dockervolume

import (
	"net"
	"os"
	"path/filepath"

	"github.com/docker/docker/pkg/sockets"
)

const (
	pluginSockDir = "/run/docker/plugins"
)

func newUnixListener(
	volumeDriverName string,
	group string,
	start <-chan struct{},
) (net.Listener, error) {
	path, err := fullSocketAddress(volumeDriverName)
	if err != nil {
		return nil, err
	}
	return sockets.NewUnixSocket(path, group, start)
}

func fullSocketAddress(address string) (string, error) {
	if err := os.MkdirAll(pluginSockDir, 0755); err != nil {
		return "", err
	}
	if filepath.IsAbs(address) {
		return address, nil
	}
	return filepath.Join(pluginSockDir, address+".sock"), nil
}
