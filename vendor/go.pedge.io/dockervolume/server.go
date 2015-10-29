package dockervolume

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/gengo/grpc-gateway/runtime"
	"github.com/golang/protobuf/proto"
	"go.pedge.io/google-protobuf"
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
	apiServer        *apiServer
	volumeDriverName string
	groupOrAddress   string
	opts             ServerOptions
}

func newServer(
	protocol int,
	volumeDriver VolumeDriver,
	volumeDriverName string,
	groupOrAddress string,
	opts ServerOptions,
) *server {
	return &server{
		protocol,
		newAPIServer(
			volumeDriver,
			volumeDriverName,
			opts.NoEvents,
		),
		volumeDriverName,
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
	if err != nil {
		return err
	}
	grpcPort := s.opts.GRPCPort
	if grpcPort == 0 {
		grpcPort = DefaultGRPCPort
	}
	return protoserver.Serve(
		grpcPort,
		func(grpcServer *grpc.Server) {
			RegisterAPIServer(grpcServer, s.apiServer)
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
			HTTPBeforeShutdown: func() { _ = s.cleanup() },
			HTTPShutdownInitiated: func() {
				if spec != "" {
					_ = os.Remove(spec)
				}
			},
			HTTPStart: start,
		},
	)
}

func (s *server) cleanup() error {
	if s.opts.CleanupOnShutdown {
		_, err := s.apiServer.Cleanup(
			context.Background(),
			google_protobuf.EmptyInstance,
		)
		return err
	}
	return nil
}
