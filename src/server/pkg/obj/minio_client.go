package obj

import (
	"io"

	minio "github.com/minio/minio-go"
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

// Represents minio writer structure with pipe and the error channel
type minioWriter struct {
	errChan chan error
	pipe    *io.PipeWriter
}

// Creates a new minio writer and a go routine to upload objects to minio server
func newMinioWriter(client *minioClient, name string) *minioWriter {
	reader, writer := io.Pipe()
	w := &minioWriter{
		errChan: make(chan error),
		pipe:    writer,
	}
	go func() {
		_, err := client.PutObject(client.bucket, name, reader, "application/octet-stream")
		if err != nil {
			reader.CloseWithError(err)
		}
		w.errChan <- err
	}()
	return w
}

func (w *minioWriter) Write(p []byte) (int, error) {
	return w.pipe.Write(p)
}

// This will block till upload is done
func (w *minioWriter) Close() error {
	if err := w.pipe.Close(); err != nil {
		return err
	}
	return <-w.errChan
}

func (c *minioClient) Writer(name string) (io.WriteCloser, error) {
	return newMinioWriter(c, name), nil
}

func (c *minioClient) Walk(name string, fn func(name string) error) error {
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

// limitReadCloser implements a closer compatible wrapper
// for a size limited reader.
type limitReadCloser struct {
	io.Reader
	mObj *minio.Object
}

func (l *limitReadCloser) Close() (err error) {
	return l.mObj.Close()
}

func (c *minioClient) Reader(name string, offset uint64, size uint64) (io.ReadCloser, error) {
	obj, err := c.GetObject(c.bucket, name)
	if err != nil {
		return nil, err
	}
	// Seek to an offset to fetch the new reader.
	_, err = obj.Seek(int64(offset), 0)
	if err != nil {
		return nil, err
	}
	if size > 0 {
		return &limitReadCloser{io.LimitReader(obj, int64(size)), obj}, nil
	}
	return obj, nil

}

func (c *minioClient) Delete(name string) error {
	return c.RemoveObject(c.bucket, name)
}

func (c *minioClient) Exists(name string) bool {
	_, err := c.StatObject(c.bucket, name)
	return err == nil
}

func (c *minioClient) IsRetryable(err error) bool {
	// Minio client already implements retrying, no
	// need for a caller retry.
	return false
}

func (c *minioClient) IsIgnorable(err error) bool {
	return false
}

// Sentinel error response returned if err is not
// of type *minio.ErrorResponse.
var sentinelErrResp = minio.ErrorResponse{}

func (c *minioClient) IsNotExist(err error) bool {
	errResp := minio.ToErrorResponse(err)
	if errResp.Code == sentinelErrResp.Code {
		return false
	}
	// Treat both object not found and bucket not found as IsNotExist().
	return errResp.Code == "NoSuchKey" || errResp.Code == "NoSuchBucket"
}
