package obj

import (
	"io"

	"golang.org/x/net/context"
	"google.golang.org/cloud/storage"
)

type googleClient struct {
	ctx    context.Context
	bucket *storage.BucketHandle
}

func newGoogleClient(ctx context.Context, bucket string) (*googleClient, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return &googleClient{ctx, client.Bucket(bucket)}, nil
}

func (c *googleClient) Writer(name string) (io.WriteCloser, error) {
	return c.bucket.Object(name).NewWriter(c.ctx), nil
}

func (c *googleClient) Reader(name string) (io.ReadCloser, error) {
	return c.bucket.Object(name).NewReader(c.ctx)
}

func (c *googleClient) Delete(name string) error {
	return c.bucket.Object(name).Delete(c.ctx)
}
