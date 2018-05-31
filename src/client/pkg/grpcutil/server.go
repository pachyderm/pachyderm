package grpcutil

import (
	"errors"
	"fmt"
	"math"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var (
	// ErrMustSpecifyRegisterFunc is used when a register func is nil.
	ErrMustSpecifyRegisterFunc = errors.New("must specify registerFunc")

	// ErrMustSpecifyPort is used when a port is 0
	ErrMustSpecifyPort = errors.New("must specify port on which to serve")
)

// ServerSpec represent optional fields for serving.
type ServerSpec struct {
	Port         uint16
	MaxMsgSize   int
	Cancel       chan struct{}
	RegisterFunc func(*grpc.Server) error
}

// Serve serves stuff.
func Serve(
	servers ...ServerSpec,
) (retErr error) {
	for _, server := range servers {
		if server.RegisterFunc == nil {
			return ErrMustSpecifyRegisterFunc
		}
		if server.Port == 0 {
			return ErrMustSpecifyPort
		}
		opts := []grpc.ServerOption{
			grpc.MaxConcurrentStreams(math.MaxUint32),
			grpc.MaxRecvMsgSize(server.MaxMsgSize),
			grpc.MaxSendMsgSize(server.MaxMsgSize),
			grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
				MinTime:             5 * time.Second,
				PermitWithoutStream: true,
			}),
		}

		GRPCServer := grpc.NewServer(opts...)
		if err := server.RegisterFunc(GRPCServer); err != nil {
			return err
		}
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", server.Port))
		if err != nil {
			return err
		}
		if server.Cancel != nil {
			go func() {
				<-server.Cancel
				if err := listener.Close(); err != nil {
					fmt.Printf("listener.Close(): %v\n", err)
				}
			}()
		}
		if err := GRPCServer.Serve(listener); err != nil {
			return err
		}
	}
	return nil
}
