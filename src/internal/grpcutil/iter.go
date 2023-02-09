package grpcutil

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
	"google.golang.org/grpc"
)

type ClientStream[T any] interface {
	Recv() (*T, error)
	grpc.ClientStream
}

type Iterator[T any] struct {
	cs ClientStream[T]
}

func NewIterator[T any](cs ClientStream[T]) Iterator[T] {
	return Iterator[T]{cs: cs}
}

func (it Iterator[T]) Next(ctx context.Context, x *T) error {
	return it.cs.RecvMsg(x)
}

func ForEach[T any](cs ClientStream[T], fn func(x T) error) error {
	return stream.ForEach[T](cs.Context(), NewIterator(cs), fn)
}

func Read[T any](cs ClientStream[T], buf []T) (int, error) {
	return stream.Read[T](cs.Context(), NewIterator(cs), buf)
}
