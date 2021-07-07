package obj

import (
	"context"
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
		return nil, err
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
		return nil, err
	}
	return &minioClient{
		bucket: bucket,
		Client: mclient,
	}, nil
}

func (c *minioClient) Put(ctx context.Context, name string, r io.Reader) (retErr error) {
	defer func() { retErr = c.transformError(retErr, name) }()
	wc := newMinioWriter(ctx, c, name)
	if _, err := io.Copy(wc, r); err != nil {
		return err
	}
	return wc.Close()
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
		return err
	}
	defer func() {
		if err := rc.Close(); retErr == nil {
			retErr = err
		}
	}()
	_, err = io.Copy(w, rc)
	return err
}

// TODO: should respect context
func (c *minioClient) Delete(_ context.Context, name string) (retErr error) {
	defer func() { retErr = c.transformError(retErr, name) }()
	return c.RemoveObject(c.bucket, name)
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

// Represents minio writer structure with pipe and the error channel
type minioWriter struct {
	ctx     context.Context
	errChan chan error
	pipe    *io.PipeWriter
}

// Creates a new minio writer and a go routine to upload objects to minio server
func newMinioWriter(ctx context.Context, client *minioClient, name string) *minioWriter {
	reader, writer := io.Pipe()
	w := &minioWriter{
		ctx:     ctx,
		errChan: make(chan error),
		pipe:    writer,
	}
	go func() {
		opts := minio.PutObjectOptions{
			ContentType: "application/octet-stream",
			PartSize:    uint64(8 * 1024 * 1024),
		}
		_, err := client.PutObject(client.bucket, name, reader, -1, opts)
		if err != nil {
			reader.CloseWithError(err)
		}
		w.errChan <- err
	}()
	return w
}

func (w *minioWriter) Write(p []byte) (retN int, retErr error) {
	span, _ := tracing.AddSpanToAnyExisting(w.ctx, "/Minio.Writer/Write")
	defer func() {
		tracing.FinishAnySpan(span, "bytes", retN, "err", retErr)
	}()
	return w.pipe.Write(p)
}

// This will block till upload is done
func (w *minioWriter) Close() (retErr error) {
	span, _ := tracing.AddSpanToAnyExisting(w.ctx, "/Minio.Writer/Close")
	defer func() {
		tracing.FinishAnySpan(span, "err", retErr)
	}()
	if err := w.pipe.Close(); err != nil {
		return err
	}
	return <-w.errChan
}
