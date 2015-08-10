package grpcutil

import (
	"fmt"
	"math"
	"net"
	"net/http"

	"google.golang.org/grpc"
)

type Dialer interface {
	Dial(address string) (*grpc.ClientConn, error)
	Clean() error
}

func NewDialer(opts ...grpc.DialOption) Dialer {
	return newDialer(opts...)
}

func GrpcDo(
	port int,
	tracePort int,
	registerFunc func(*grpc.Server),
) error {
	s := grpc.NewServer(grpc.MaxConcurrentStreams(math.MaxUint32))
	registerFunc(s)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	errC := make(chan error)
	go func() { errC <- s.Serve(listener) }()
	if tracePort != 0 {
		go func() { errC <- http.ListenAndServe(fmt.Sprintf(":%d", tracePort), nil) }()
	}
	return <-errC
}
