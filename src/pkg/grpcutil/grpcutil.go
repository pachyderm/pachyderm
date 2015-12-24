package grpcutil

import (
	"net"
	"time"

	"google.golang.org/grpc"
)

type Dialer interface {
	Dial(address string) (*grpc.ClientConn, error)
	Clean() error
}

func NewDialer(opts ...grpc.DialOption) Dialer {
	return newDialer(opts...)
}

type DialerFunc func(addr string, timeout time.Duration) (net.Conn, error)

func ListenerClientConnPair() (net.Listener, *grpc.ClientConn) {
	left, right := net.Pipe()
	return newStaticListener(left), staticClienConn(right)
}
