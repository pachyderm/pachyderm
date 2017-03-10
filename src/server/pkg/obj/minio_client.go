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

func (c *minioClient) Writer(name string) (io.WriteCloser, error) {
	reader, writer := io.Pipe()
	go func(reader *io.PipeReader) {
		_, err := c.PutObject(c.bucket, name, reader, "application/octet-stream")
		if err != nil {
			reader.CloseWithError(err)
			return
		}
	}(reader)
	return writer, nil
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
	if offset > 0 {
		// Seek to an offset to fetch the new reader.
		_, err = obj.Seek(int64(offset), 0)
		if err != nil {
			return nil, err
		}
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

func (c *minioClient) isRetryable(err error) bool {
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
	if errResp == sentinelErrResp {
		return false
	}
	// Treat both object not found and bucket not found as IsNotExist().
	return errResp.Code == "NoSuchKey" || errResp.Code == "NoSuchBucket"
}
