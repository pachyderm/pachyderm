package grpcutil

import (
	"context"
	"io"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
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

// ForEach calls fn for each element in cs.
// fn must not retain the element passed to it.
func ForEach[T any](cs ClientStream[T], fn func(x *T) error) error {
	return stream.ForEach[T](cs.Context(), NewIterator(cs), func(x T) error {
		return fn(&x)
	})
}

// Collect reads at most max elements
func Collect[T any](cs ClientStream[T], max int) (ret []*T, _ error) {
	it := NewIterator(cs)
	for {
		x := new(T)
		if err := it.Next(cs.Context(), x); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		ret = append(ret, x)
	}
	return ret, nil
}
