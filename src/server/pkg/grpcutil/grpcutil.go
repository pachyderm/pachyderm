package grpcutil

import (
	"path/filepath"

	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"google.golang.org/grpc"
)

type Dialer interface {
	Dial(address string) (*grpc.ClientConn, error)
	Clean() error
}

func NewDialer(opts ...grpc.DialOption) Dialer {
	return newDialer(opts...)
}

type LocalServer interface {
	Server() *grpc.Server
	Serve() error
	Dial() (*grpc.ClientConn, error)
}

func NewLocalServer() LocalServer {
	return &localServer{
		server: grpc.NewServer(),
		path:   filepath.Join("/tmp", uuid.NewWithoutDashes()),
	}
}
