package obj

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pacherr"
	"github.com/pachyderm/pachyderm/v2/src/internal/promutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"

	minio "github.com/minio/minio-go/v6"
)

// Represents minio client instance for any s3 compatible server.
type minioClient struct {
	*minio.Client
	bucket string
}

// Creates a new minioClient structure and returns
func newMinioClient(endpoint, bucket, id, secret string, secure bool) (*minioClient, error) {
	mclient, err := minio.New(endpoint, id, secret, secure)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	mclient.SetCustomTransport(promutil.InstrumentRoundTripper("minio", http.DefaultTransport))
	return &minioClient{
		bucket: bucket,
		Client: mclient,
	}, nil
}

// Creates a new minioClient S3V2 structure and returns
func newMinioClientV2(endpoint, bucket, id, secret string, secure bool) (*minioClient, error) {
	mclient, err := minio.NewV2(endpoint, id, secret, secure)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &minioClient{
		bucket: bucket,
		Client: mclient,
	}, nil
}

func (c *minioClient) Put(ctx context.Context, name string, r io.Reader) (retErr error) {
	defer func() { retErr = c.transformError(retErr, name) }()
	opts := minio.PutObjectOptions{
		ContentType: "application/octet-stream",
		PartSize:    uint64(8 * 1024 * 1024),
	}
	_, err := c.Client.PutObjectWithContext(ctx, c.bucket, name, r, -1, opts)
	return errors.EnsureStack(err)
}

// TODO: this should respect the context
func (c *minioClient) Walk(_ context.Context, name string, fn func(name string) error) (retErr error) {
	defer func() { retErr = c.transformError(retErr, name) }()
	recursive := true // Recursively walk by default.

	doneCh := make(chan struct{})
	defer close(doneCh)
	for objInfo := range c.ListObjectsV2(c.bucket, name, recursive, doneCh) {
		if objInfo.Err != nil {
			return objInfo.Err
		}
		if err := fn(objInfo.Key); err != nil {
			return err
		}
	}
	return nil
}

func (c *minioClient) Get(ctx context.Context, name string, w io.Writer) (retErr error) {
	defer func() { retErr = c.transformError(retErr, name) }()
	rc, err := c.GetObjectWithContext(ctx, c.bucket, name, minio.GetObjectOptions{})
	if err != nil {
		return errors.EnsureStack(err)
	}
	defer func() {
		if err := rc.Close(); retErr == nil {
			retErr = err
		}
	}()
	_, err = io.Copy(w, rc)
	return errors.EnsureStack(err)
}

// TODO: should respect context
func (c *minioClient) Delete(_ context.Context, name string) (retErr error) {
	defer func() { retErr = c.transformError(retErr, name) }()
	return errors.EnsureStack(c.RemoveObject(c.bucket, name))
}

func (c *minioClient) Exists(ctx context.Context, name string) (bool, error) {
	_, err := c.StatObjectWithContext(ctx, c.bucket, name, minio.StatObjectOptions{})
	tracing.TagAnySpan(ctx, "err", err)
	if err != nil {
		err = c.transformError(err, name)
		if pacherr.IsNotExist(err) {
			err = nil
		}
		return false, err
	}
	return true, nil
}

func (c *minioClient) BucketURL() ObjectStoreURL {
	u := c.Client.EndpointURL()
	return ObjectStoreURL{
		Scheme: "minio",
		Bucket: fmt.Sprintf("%s/%s", u.Host, c.bucket),
	}
}

func (c *minioClient) transformError(err error, objectPath string) error {
	if err == nil {
		return nil
	}
	errResp := minio.ErrorResponse{}
	if !errors.As(err, &errResp) {
		return err
	}
	if errResp.Code == sentinelErrResp.Code {
		return err
	}
	// Treat both object not found and bucket not found as IsNotExist().
	if errResp.Code == "NoSuchKey" || errResp.Code == "NoSuchBucket" {
		return pacherr.NewNotExist(c.bucket, objectPath)
	}
	return err
}

// Sentinel error response returned if err is not
// of type *minio.ErrorResponse.
var sentinelErrResp = minio.ErrorResponse{}
