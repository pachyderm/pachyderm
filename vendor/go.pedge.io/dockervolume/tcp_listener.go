package dockervolume

import (
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	"github.com/docker/docker/pkg/sockets"
)

const (
	pluginSpecDir = "/etc/docker/plugins"
)

func newTCPListener(
	volumeDriverName string,
	address string,
	start <-chan struct{},
) (net.Listener, error) {
	listener, err := sockets.NewTCPSocket(address, nil, start)
	if err != nil {
		return nil, err
	}
	if err := writeSpec(volumeDriverName, listener.Addr().String()); err != nil {
		return nil, err
	}
	return listener, nil
}

func writeSpec(name string, address string) error {
	if err := os.MkdirAll(pluginSpecDir, 0755); err != nil {
		return err
	}
	spec := filepath.Join(pluginSpecDir, name+".spec")
	url := "tcp://" + address
	return ioutil.WriteFile(spec, []byte(url), 0644)
}
