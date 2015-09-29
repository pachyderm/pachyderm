package dockervolume

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/pkg/sockets"
)

const (
	pluginSpecDir = "/etc/docker/plugins"
)

func newTCPListener(
	volumeDriverName string,
	address string,
	start <-chan struct{},
) (net.Listener, string, error) {
	listener, err := sockets.NewTCPSocket(address, nil, start)
	if err != nil {
		return nil, "", err
	}
	spec, err := writeSpec(volumeDriverName, listener.Addr().String())
	if err != nil {
		return nil, "", err
	}
	return listener, spec, nil
}

func writeSpec(name string, address string) (string, error) {
	dir := filepath.Join(pluginSpecDir, name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	spec := filepath.Join(dir, name+".spec")
	if strings.HasPrefix(address, "[::]:") {
		address = fmt.Sprintf("0.0.0.0:%s", strings.TrimPrefix(address, "[::]:"))
	}
	url := "tcp://" + address
	if err := ioutil.WriteFile(spec, []byte(url), 0644); err != nil {
		return "", err
	}
	return spec, nil
}
