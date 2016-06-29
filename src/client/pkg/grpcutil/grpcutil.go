package grpcutil

import (
	"google.golang.org/grpc"
)

type Dialer interface {
	Dial(address string) (*grpc.ClientConn, error)
	CloseConns() error
}

func NewDialer(opts ...grpc.DialOption) Dialer {
	return newDialer(opts...)
}
