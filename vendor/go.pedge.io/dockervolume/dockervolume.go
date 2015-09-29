package dockervolume

import "google.golang.org/grpc"

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

// VolumeDriverClient can call VolumeDrivers, along with additional functionality.
type VolumeDriverClient interface {
	VolumeDriver
}

// NewVolumeDriverClient creates a new VolumeDriverClient for the given *grpc.ClientConn.
func NewVolumeDriverClient(clientConn *grpc.ClientConn) VolumeDriverClient {
	return newVolumeDriverClient(clientConn)
}

// Server serves a VolumeDriver.
type Server interface {
	Serve() error
}

// ServerOptions are options for a Server.
type ServerOptions struct {
	GRPCDebugPort uint16
}

// NewTCPServer returns a new Server for TCP.
func NewTCPServer(
	volumeDriver VolumeDriver,
	volumeDriverName string,
	grpcPort uint16,
	address string,
	opts ServerOptions,
) Server {
	return newServer(
		protocolTCP,
		volumeDriver,
		volumeDriverName,
		grpcPort,
		address,
		opts,
	)
}

// NewUnixServer returns a new Server for Unix sockets.
func NewUnixServer(
	volumeDriver VolumeDriver,
	volumeDriverName string,
	grpcPort uint16,
	group string,
	opts ServerOptions,
) Server {
	return newServer(
		protocolUnix,
		volumeDriver,
		volumeDriverName,
		grpcPort,
		group,
		opts,
	)
}
