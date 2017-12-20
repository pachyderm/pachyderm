package obj

import (
	"io"
	"strings"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type googleClient struct {
	ctx    context.Context
	bucket *storage.BucketHandle
}

func newGoogleClient(ctx context.Context, bucket string, credFile string) (*googleClient, error) {
	opts := []option.ClientOption{option.WithScopes(storage.ScopeFullControl)}
	if credFile == "" {
		opts = append(opts, option.WithTokenSource(google.ComputeTokenSource("")))
	} else {
		opts = append(opts, option.WithCredentialsFile(credFile))
	}
	client, err := storage.NewClient(ctx, opts...)
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
	objectIter := c.bucket.Objects(c.ctx, &storage.Query{Prefix: name})
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
