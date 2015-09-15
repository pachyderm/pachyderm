package dockervolume

import (
	"fmt"
	"net"
	"net/http"
	"os"
)

const (
	// ProtocolTCP denotes using TCP.
	ProtocolTCP Protocol = iota
	// ProtocolUnix denotes using Unix sockets.
	ProtocolUnix
)

// Protocol represents TCP or Unix.
type Protocol int

// VolumeDriver is the interface that should be implemented for custom volume drivers.
type VolumeDriver interface {
	// Create a volume with the given name
	Create(name string, opts map[string]string) (err error)
	// Remove the volume with the given name
	Remove(name string) (err error)
	// Get the mountpoint of the given volume
	Path(name string) (mountpoint string, err error)
	// Mount the given volume and return the mountpoint
	Mount(name string) (mountpoint string, err error)
	// Unmount the given volume
	Unmount(name string) (err error)
}

// Logger is a generic interface for logging requests to a VolumeDriver.
type Logger interface {
	LogCall(call *Call)
}

// Handler is the http.Handler used for the volume driver plugin, plus additional methods.
type Handler interface {
	http.Handler
}

// HandlerOptions are options for a new volume driver handler.
type HandlerOptions struct {
	// Logger specifies a customer logger.
	//
	// If not specified, the default Logger will be used.
	Logger Logger
	// Reset specifies to not use previous state.
	Reset bool
}

// NewHandler returns a new http.Handler.
func NewHandler(volumeDriver VolumeDriver, opts HandlerOptions) Handler {
	return newHandler(volumeDriver, opts)
}

// NewTCPListener returns a new net.Listener for TCP.
//
// The string returned is a file that should be removed when finished with the listener.
func NewTCPListener(
	volumeDriverName string,
	address string,
	start <-chan struct{},
) (net.Listener, string, error) {
	return newTCPListener(
		volumeDriverName,
		address,
		start,
	)
}

// NewUnixListener returns a new net.Listener for Unix.
//
// The string returned is a file that should be removed when finished with the listener.
func NewUnixListener(
	volumeDriverName string,
	group string,
	start <-chan struct{},
) (net.Listener, string, error) {
	return newUnixListener(
		volumeDriverName,
		group,
		start,
	)
}

// Serve serves the volume driver handler.
func Serve(
	handler Handler,
	protocol Protocol,
	volumeDriverName string,
	groupOrAddress string,
) (retErr error) {
	server := &http.Server{
		Handler: handler,
	}
	start := make(chan struct{})
	var listener net.Listener
	var spec string
	var err error
	switch protocol {
	case ProtocolTCP:
		listener, spec, err = NewTCPListener(volumeDriverName, groupOrAddress, start)
		server.Addr = groupOrAddress
	case ProtocolUnix:
		listener, spec, err = NewUnixListener(volumeDriverName, groupOrAddress, start)
		server.Addr = volumeDriverName
	default:
		return fmt.Errorf("unknown protocol: %v", protocol)
	}
	if spec != "" {
		defer func() {
			if err := os.Remove(spec); err != nil && retErr == nil {
				retErr = err
			}
		}()
	}
	if err != nil {
		return err
	}
	close(start)
	return server.Serve(listener)
}
