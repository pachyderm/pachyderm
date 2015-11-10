package obj

import (
	"io"

	"golang.org/x/net/context"
)

type Client interface {
	Writer(name string) (io.WriteCloser, error)
	Reader(name string) (io.ReadCloser, error)
	Delete(name string) error
}

func NewClientGoogleClient(ctx context.Context, bucket string) Client {
	return newGoogleClient(ctx, bucket)
}
