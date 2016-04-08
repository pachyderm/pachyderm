package obj

import (
	"io"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/cloud"
	"google.golang.org/cloud/storage"
)

type googleClient struct {
	ctx    context.Context
	bucket *storage.BucketHandle
}

func newGoogleClient(ctx context.Context, bucket string) (*googleClient, error) {
	client, err := storage.NewClient(
		ctx,
		cloud.WithTokenSource(google.ComputeTokenSource("")),
		cloud.WithScopes(storage.ScopeFullControl),
	)
	if err != nil {
		return nil, err
	}
	return &googleClient{ctx, client.Bucket(bucket)}, nil
}

func (c *googleClient) Writer(name string) (io.WriteCloser, error) {
	return c.bucket.Object(name).NewWriter(c.ctx), nil
}

func (c *googleClient) Walk(name string, fn func(name string) error) error {
	query := &storage.Query{Prefix: name}
	for query != nil {
		objectList, err := c.bucket.List(c.ctx, query)
		if err != nil {
			return err
		}
		query = objectList.Next
		for _, objectAttrs := range objectList.Results {
			if err := fn(objectAttrs.Name); err != nil {
				return err
			}
		}
	}
	return nil
}

//TODO size 0 means read all
func (c *googleClient) Reader(name string, offset uint64, size uint64) (io.ReadCloser, error) {
	return c.bucket.Object(name).NewReader(c.ctx)
}

func (c *googleClient) Delete(name string) error {
	return c.bucket.Object(name).Delete(c.ctx)
}
