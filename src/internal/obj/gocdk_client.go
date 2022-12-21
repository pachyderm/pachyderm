package obj

import (
	"context"
	"fmt"
	"io"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"gocloud.dev/blob"

	// Import the blob packages for the cloud backends we want to be able to open.
	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/memblob"
	_ "gocloud.dev/blob/s3blob"
)

// We no longer want to use this obj.Client interface as it limits us, so we will have to write a new interface
// when we do the full port.

type goCDKClient struct {
	url *ObjectStoreURL
}

func newGoCDKClient(url *ObjectStoreURL) *goCDKClient {
	return &goCDKClient{
		url: url,
	}
}

func (c *goCDKClient) Put(ctx context.Context, name string, r io.Reader) (retErr error) {
	bucket, err := blob.OpenBucket(ctx, c.bucketBaseURL())
	if err != nil {
		return errors.Wrapf(err, "error opening bucket %s", c.url.Bucket)
	}
	defer func() {
		if err := bucket.Close(); err != nil {
			retErr = multierror.Append(retErr, errors.Wrapf(err, "error closing bucket %s", c.url.Bucket))
		}
	}()
	exists, err := bucket.Exists(ctx, name)
	if err != nil {
		return errors.Wrapf(err, "error checking if key %s exists in bucket %s", name, c.url.Bucket)
	}
	if exists {
		return fmt.Errorf("key %s already exists in bucket %s", name, c.url.Bucket)
	}
	w, err := bucket.NewWriter(ctx, name, nil)
	if err != nil {
		return errors.Wrapf(err, "error creating writer for bucket %s", c.url.Bucket)
	}
	defer func() {
		if err := w.Close(); err != nil {
			retErr = multierror.Append(retErr, errors.Wrapf(err, "error closing writer for bucket %s", c.url.Bucket))
		}
	}()
	_, err = w.ReadFrom(r)
	if err != nil {
		return errors.Wrapf(err, "error writing to bucket %s", c.url.Bucket)
	}
	return nil
}

// Get writes the data for an object to w
// If `size == 0`, the reader should read from the offset till the end of the object.
// It should error if the object doesn't exist or we don't have sufficient
// permission to read it.
func (c *goCDKClient) Get(ctx context.Context, name string, w io.Writer) (retErr error) {
	bucket, err := blob.OpenBucket(ctx, c.url.String())
	if err != nil {
		return errors.Wrapf(err, "error opening bucket %s", c.url.Bucket)
	}
	defer func() {
		if err := bucket.Close(); err != nil {
			retErr = multierror.Append(retErr, errors.Wrapf(err, "error closing bucket %s", c.url.Bucket))
		}
	}()
	r, err := bucket.NewReader(ctx, name, nil)
	if err != nil {
		return errors.Wrapf(err, "error creating reader from bucket %s", c.url.Bucket)
	}
	defer func() {
		if err := r.Close(); err != nil {
			retErr = multierror.Append(retErr, errors.Wrapf(err, "error closing reader for  bucket %s", c.url.Bucket))
		}
	}()
	_, err = io.Copy(w, r)
	if err != nil {
		return errors.Wrapf(err, "error copying from reader to writer for bucket %s", c.url.Bucket)
	}
	return nil
}

// Delete deletes an object.
// It should error if the object doesn't exist or we don't have sufficient
// permission to delete it.
func (c *goCDKClient) Delete(ctx context.Context, name string) (retErr error) {
	bucket, err := blob.OpenBucket(ctx, c.bucketBaseURL())
	if err != nil {
		return errors.Wrapf(err, "error opening bucket %s", c.url.Bucket)
	}
	defer func() {
		if err := bucket.Close(); err != nil {
			retErr = multierror.Append(retErr, errors.Wrapf(err, "error closing bucket %s", c.url.Bucket))
		}
	}()
	err = bucket.Delete(ctx, name)
	if err != nil {
		return errors.Wrapf(err, "error deleting %s in bucket %s", name, c.url.Bucket)
	}
	return nil
}

// Walk calls `fn` with the names of objects which can be found under `prefix`.
func (c *goCDKClient) Walk(ctx context.Context, prefix string, fn func(name string) error) (retErr error) {
	bucket, err := blob.OpenBucket(ctx, c.bucketBaseURL())
	if err != nil {
		return errors.Wrapf(err, "error opening bucket %s", c.url.Bucket)
	}
	defer func() {
		if err := bucket.Close(); err != nil {
			retErr = multierror.Append(retErr, errors.Wrapf(err, "error closing bucket %s", c.url.Bucket))
		}
	}()
	list := bucket.List(&blob.ListOptions{Prefix: prefix})
	var obj *blob.ListObject
	for {
		obj, err = list.Next(ctx)
		if err != nil {
			if err == io.EOF {
				break
			}
			return errors.Wrapf(err, "error iterating through list for bucket %s", c.url.Bucket)
		}
		if err := fn(obj.Key); err != nil {
			return errors.Wrapf(err, "error from callback on key %s in bucket %s", obj.Key, c.url.Bucket)
		}
	}
	return nil
}

// Exists checks if a given object already exists
func (c *goCDKClient) Exists(ctx context.Context, name string) (exists bool, retErr error) {
	bucket, err := blob.OpenBucket(ctx, c.bucketBaseURL())
	if err != nil {
		return false, errors.Wrapf(err, "error opening bucket %s", c.url.Bucket)
	}
	defer func() {
		if err := bucket.Close(); err != nil {
			retErr = multierror.Append(retErr, errors.Wrapf(err, "error closing bucket %s", c.url.Bucket))
		}
	}()
	exists, err = bucket.Exists(ctx, name)
	if err != nil {
		return false, errors.Wrapf(err, "error checking if key %s exists in bucket %s", name, c.url.Bucket)
	}
	return exists, err
}

// BucketURL returns the URL of the bucket this client uses.
func (c *goCDKClient) BucketURL() ObjectStoreURL {
	return *c.url
}

func (c *goCDKClient) bucketBaseURL() string {
	return fmt.Sprintf("%s://%s", c.url.Scheme, c.url.Bucket)
}
