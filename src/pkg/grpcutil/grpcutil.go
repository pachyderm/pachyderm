package grpcutil

import "google.golang.org/grpc"

type Dialer interface {
	Dial(address string) (*grpc.ClientConn, error)
	Clean() error
}

func NewDialer(opts ...grpc.DialOption) Dialer {
	return newDialer(opts...)
}
