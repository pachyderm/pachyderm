package util

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// we use ServerStreamWrapper to set the stream's Context with added values
type ServerStreamWrapper struct {
	Stream grpc.ServerStream
	Ctx    context.Context
}

func (s ServerStreamWrapper) Context() context.Context {
	return s.Ctx
}

func (s ServerStreamWrapper) SetHeader(md metadata.MD) error {
	return s.Stream.SetHeader(md)
}

func (s ServerStreamWrapper) SendHeader(md metadata.MD) error {
	return s.Stream.SendHeader(md)
}

func (s ServerStreamWrapper) SetTrailer(md metadata.MD) {
	s.Stream.SetTrailer(md)
}

func (s ServerStreamWrapper) SendMsg(m interface{}) error {
	return s.Stream.SendMsg(m)
}

func (s ServerStreamWrapper) RecvMsg(m interface{}) error {
	return s.Stream.RecvMsg(m)
}
