package dockervolume

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/gengo/grpc-gateway/runtime"
	"github.com/golang/protobuf/proto"
	"go.pedge.io/proto/server"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	protocolTCP = iota
	protocolUnix
)

type server struct {
	protocol         int
	volumeDriver     VolumeDriver
	volumeDriverName string
	grpcPort         uint16
	groupOrAddress   string
	opts             ServerOptions
}

func newServer(
	protocol int,
	volumeDriver VolumeDriver,
	volumeDriverName string,
	grpcPort uint16,
	groupOrAddress string,
	opts ServerOptions,
) *server {
	return &server{
		protocol,
		volumeDriver,
		volumeDriverName,
		grpcPort,
		groupOrAddress,
		opts,
	}
}

func (s *server) Serve() (retErr error) {
	start := make(chan struct{})
	var listener net.Listener
	var spec string
	var err error
	var addr string
	switch s.protocol {
	case protocolTCP:
		listener, spec, err = newTCPListener(s.volumeDriverName, s.groupOrAddress, start)
		addr = s.groupOrAddress
	case protocolUnix:
		listener, spec, err = newUnixListener(s.volumeDriverName, s.groupOrAddress, start)
		addr = s.volumeDriverName
	default:
		return fmt.Errorf("unknown protocol: %d", s.protocol)
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
	return protoserver.Serve(
		s.grpcPort,
		func(grpcServer *grpc.Server) {
			RegisterAPIServer(grpcServer, newAPIServer(s.volumeDriver))
		},
		protoserver.ServeOptions{
			DebugPort: s.opts.GRPCDebugPort,
			HTTPRegisterFunc: func(ctx context.Context, mux *runtime.ServeMux, clientConn *grpc.ClientConn) error {
				return RegisterAPIHandler(ctx, mux, clientConn)
			},
			HTTPAddress:  addr,
			HTTPListener: listener,
			ServeMuxOptions: []runtime.ServeMuxOption{
				runtime.WithForwardResponseOption(
					func(_ context.Context, responseWriter http.ResponseWriter, _ proto.Message) error {
						responseWriter.Header().Set("Content-Type", "application/vnd.docker.plugins.v1.1+json")
						return nil
					},
				),
			},
		},
	)
}
