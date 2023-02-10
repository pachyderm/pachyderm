package grpcutil

import (
	"context"

	"github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
	"google.golang.org/grpc"
)

type ClientStream[T proto.Message] interface {
	Recv() (T, error)
	grpc.ClientStream
}

type iterator[T proto.Message] struct {
	cs ClientStream[T]
}

func newIterator[T proto.Message](cs ClientStream[T]) iterator[T] {
	return iterator[T]{cs: cs}
}

func (it iterator[T]) Next(ctx context.Context, dst *T) error {
	x, err := it.cs.Recv()
	if err != nil {
		return err
	}
	*dst = x
	return nil
}

// ForEach calls fn for each element in cs.
// fn must not retain the element passed to it.
func ForEach[T proto.Message](cs ClientStream[T], fn func(x T) error) error {
	return stream.ForEach[T](cs.Context(), newIterator(cs), fn)
}

// Read fills buf with received messages from cs and returns the number read.
func Read[T proto.Message](cs ClientStream[T], buf []T) (int, error) {
	return stream.Read[T](cs.Context(), newIterator(cs), buf)
}

// Collect reads at most max elements from cs, and returns them as a slice.
func Collect[T proto.Message](cs ClientStream[T], max int) (ret []T, _ error) {
	return stream.Collect[T](cs.Context(), newIterator(cs), max)
}
