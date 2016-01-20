package obj

import (
	"io"

	"golang.org/x/net/context"
)

type Client interface {
	Writer(name string) (io.WriteCloser, error)
	//TODO size 0 means read all
	Reader(name string, offset uint64, size uint64) (io.ReadCloser, error)
	Delete(name string) error
	Walk(name string, fn func(name string) error) error
}

func NewClientGoogleClient(ctx context.Context, bucket string) (Client, error) {
	return newGoogleClient(ctx, bucket)
}
