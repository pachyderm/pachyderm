package obj

import (
	"io"
	"strings"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type googleClient struct {
	bucket *storage.BucketHandle
}

func newGoogleClient(bucket string, opts []option.ClientOption) (*googleClient, error) {
	opts = append(opts, option.WithScopes(storage.ScopeFullControl))
	client, err := storage.NewClient(context.Background(), opts...)
	if err != nil {
		return nil, err
	}
	return &googleClient{client.Bucket(bucket)}, nil
}

func (c *googleClient) Exists(ctx context.Context, name string) bool {
	_, err := c.bucket.Object(name).Attrs(ctx)
	return err == nil
}

func (c *googleClient) Writer(ctx context.Context, name string) (io.WriteCloser, error) {
	return newBackoffWriteCloser(ctx, c, c.bucket.Object(name).NewWriter(ctx)), nil
}

func (c *googleClient) Walk(ctx context.Context, name string, fn func(name string) error) error {
	objectIter := c.bucket.Objects(ctx, &storage.Query{Prefix: name})
	for {
		objectAttrs, err := objectIter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return err
		}
		if err := fn(objectAttrs.Name); err != nil {
			return err
		}
	}
	return nil
}

func (c *googleClient) Reader(ctx context.Context, name string, offset uint64, size uint64) (io.ReadCloser, error) {
	var reader io.ReadCloser
	var err error
	if size == 0 {
		// a negative length will cause the object to be read till the end
		reader, err = c.bucket.Object(name).NewRangeReader(ctx, int64(offset), -1)
	} else {
		reader, err = c.bucket.Object(name).NewRangeReader(ctx, int64(offset), int64(size))
	}
	if err != nil {
		return nil, err
	}
	return newBackoffReadCloser(ctx, c, reader), nil
}

func (c *googleClient) Delete(ctx context.Context, name string) error {
	return c.bucket.Object(name).Delete(ctx)
}

func (c *googleClient) IsRetryable(err error) (ret bool) {
	googleErr, ok := err.(*googleapi.Error)
	if !ok {
		return false
	}
	// https://github.com/pachyderm/pachyderm/issues/912
	return googleErr.Code >= 500 || strings.Contains(err.Error(), "Parse Error")
}

func (c *googleClient) IsNotExist(err error) (result bool) {
	return err == storage.ErrObjectNotExist
}

func (c *googleClient) IsIgnorable(err error) bool {
	googleErr, ok := err.(*googleapi.Error)
	if !ok {
		return false
	}
	return googleErr.Code == 429
}
