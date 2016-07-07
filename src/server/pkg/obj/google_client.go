package obj

import (
	"io"
	"reflect"

	"go.pedge.io/lion/proto"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/googleapi"
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

func (c *googleClient) Exists(name string) bool {
	_, err := c.bucket.Object(name).Attrs(c.ctx)
	return err == nil
}

func (c *googleClient) Writer(name string) (io.WriteCloser, error) {
	return newBackoffWriteCloser(c, c.bucket.Object(name).NewWriter(c.ctx)), nil
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

func (c *googleClient) Reader(name string, offset uint64, size uint64) (io.ReadCloser, error) {
	var reader io.ReadCloser
	var err error
	if size == 0 {
		// a negative length will cause the object to be read till the end
		reader, err = c.bucket.Object(name).NewRangeReader(c.ctx, int64(offset), -1)
	} else {
		reader, err = c.bucket.Object(name).NewRangeReader(c.ctx, int64(offset), int64(size))
	}
	if err != nil {
		return nil, err
	}
	return newBackoffReadCloser(c, reader), nil
}

func (c *googleClient) Delete(name string) error {
	return c.bucket.Object(name).Delete(c.ctx)
}

func (c *googleClient) IsRetryable(err error) (ret bool) {
	defer func() {
		protolion.Infof("retryable: %v; type of err: %s; err: %v", ret, reflect.TypeOf(err).String(), err)
	}()
	googleErr, ok := err.(*googleapi.Error)
	if !ok {
		return false
	}
	return googleErr.Code >= 500
}

func (c *googleClient) IsNotExist(err error) bool {
	googleErr, ok := err.(*googleapi.Error)
	if !ok {
		return false
	}
	return googleErr.Code == 404
}

func (c *googleClient) IsIgnorable(err error) bool {
	googleErr, ok := err.(*googleapi.Error)
	if !ok {
		return false
	}
	return googleErr.Code == 429
}
