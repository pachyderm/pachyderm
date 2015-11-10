package obj

import (
	"io"

	"golang.org/x/net/context"
	"google.golang.org/cloud/storage"
)

type googleClient struct {
	ctx    context.Context
	bucket string
}

func newGoogleClient(ctx context.Context, bucket string) *googleClient {
	return &googleClient{ctx, bucket}
}

func (c *googleClient) Writer(name string) (io.WriteCloser, error) {
	return storage.NewWriter(c.ctx, c.bucket, name), nil
}

func (c *googleClient) Reader(name string) (io.ReadCloser, error) {
	return storage.NewReader(c.ctx, c.bucket, name)
}

func (c *googleClient) Delete(name string) error {
	return storage.DeleteObject(c.ctx, c.bucket, name)
}
