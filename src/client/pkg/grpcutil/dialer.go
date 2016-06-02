package grpcutil

import (
	"google.golang.org/grpc"
)

type dialer struct {
	opts []grpc.DialOption
}

func newDialer(opts ...grpc.DialOption) *dialer {
	return &dialer{opts}
}

func (d *dialer) Dial(address string) (*grpc.ClientConn, error) {
	return grpc.Dial(address, d.opts...)
}
