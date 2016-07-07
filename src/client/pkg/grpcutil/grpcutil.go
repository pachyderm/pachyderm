package grpcutil

import (
	"google.golang.org/grpc"
)

// Dialer defines a grpc.ClientConn connection dialer.
type Dialer interface {
	Dial(address string) (*grpc.ClientConn, error)
	CloseConns() error
}

// NewDialer creates a Dialer.
func NewDialer(opts ...grpc.DialOption) Dialer {
	return newDialer(opts...)
}
