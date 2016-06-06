package grpcutil

import (
	"google.golang.org/grpc"
)

type Dialer interface {
	Dial(address string) (*grpc.ClientConn, error)
}

func NewDialer(opts ...grpc.DialOption) Dialer {
	return newDialer(opts...)
}
