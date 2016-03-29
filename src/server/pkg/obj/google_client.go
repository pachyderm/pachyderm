package obj

import (
	"fmt"
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

func (c *googleClient) Walk(name string, fn func(name string) error) error {
	return fmt.Errorf("googleClient.Walk: not implemented")
}

//TODO size 0 means read all
func (c *googleClient) Reader(name string, offset uint64, size uint64) (io.ReadCloser, error) {
	return c.bucket.Object(name).NewReader(c.ctx)
}

func (c *googleClient) Delete(name string) error {
	return c.bucket.Object(name).Delete(c.ctx)
}
