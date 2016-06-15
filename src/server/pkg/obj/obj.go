package obj

import (
	"io"
	"time"

	"github.com/cenkalti/backoff"
	"go.pedge.io/lion/proto"
	"golang.org/x/net/context"
)

// Client is an interface to object storage.
type Client interface {
	// Writer returns a writer which writes to an object.
	// It should error if the object already exists or we don't have sufficient
	// permissions to write it.
	Writer(name string) (io.WriteCloser, error)
	// Reader returns a reader which reads from an object.
	// If `size == 0`, the reader should read from the offset till the end of the object.
	// It should error if the object doesn't exist or we don't have sufficient
	// permission to read it.
	Reader(name string, offset uint64, size uint64) (io.ReadCloser, error)
	// Delete deletes an object.
	// It should error if the object doesn't exist or we don't have sufficient
	// permission to delete it.
	Delete(name string) error
	// Walk calls `fn` with the names of objects which can be found under `prefix`.
	Walk(prefix string, fn func(name string) error) error
	// IsRetryable determines if an operation should be retried given an error
	IsRetryable(err error) bool
}

func NewGoogleClient(ctx context.Context, bucket string) (Client, error) {
	return newGoogleClient(ctx, bucket)
}

func NewAmazonClient(bucket string, id string, secret string, token string,
	region string) (Client, error) {
	return newAmazonClient(bucket, id, secret, token, region)
}

func newExponentialBackOffConfig() *backoff.ExponentialBackOff {
	config := backoff.NewExponentialBackOff()
	// We want to backoff more aggressively (i.e. wait longer) than the default
	config.InitialInterval = 1 * time.Second
	config.Multiplier = 2
	config.MaxElapsedTime = 5 * time.Minute
	return config
}

type RetryError struct {
	err               string
	timeTillNextRetry string
	bytesProcessed    int
}

// BackoffReadCloser retries with exponential backoff in the case of failures
type BackoffReadCloser struct {
	client        Client
	reader        io.ReadCloser
	backoffConfig *backoff.ExponentialBackOff
}

func newBackoffReadCloser(client Client, reader io.ReadCloser) io.ReadCloser {
	return &BackoffReadCloser{
		client:        client,
		reader:        reader,
		backoffConfig: newExponentialBackOffConfig(),
	}
}

func (b *BackoffReadCloser) Read(data []byte) (int, error) {
	bytesRead := 0
	var n int
	var err error
	backoff.RetryNotify(func() error {
		n, err = b.reader.Read(data[bytesRead:])
		bytesRead += n
		if b.client.IsRetryable(err) {
			return err
		}
		return nil
	}, b.backoffConfig, func(err error, d time.Duration) {
		protolion.Infof("Error reading (retrying): %#v", RetryError{
			err:               err.Error(),
			timeTillNextRetry: d.String(),
			bytesProcessed:    bytesRead,
		})
	})
	if err == nil && eof {
		err = io.EOF
	}
	return bytesRead, err
}

func (b *BackoffReadCloser) Close() error {
	return b.reader.Close()
}

// BackoffWriteCloser retries with exponential backoff in the case of failures
type BackoffWriteCloser struct {
	client        Client
	writer        io.WriteCloser
	backoffConfig *backoff.ExponentialBackOff
}

func newBackoffWriteCloser(client Client, writer io.WriteCloser) io.WriteCloser {
	return &BackoffWriteCloser{
		client:        client,
		writer:        writer,
		backoffConfig: newExponentialBackOffConfig(),
	}
}

func (b *BackoffWriteCloser) Write(data []byte) (int, error) {
	bytesWritten := 0
	var n int
	var err error
	backoff.RetryNotify(func() error {
		n, err = b.writer.Write(data[bytesWritten:])
		bytesWritten += n
		if b.client.IsRetryable(err) {
			return err
		}
		return nil
	}, b.backoffConfig, func(err error, d time.Duration) {
		protolion.Infof("Error writing (retrying): %#v", RetryError{
			err:               err.Error(),
			timeTillNextRetry: d.String(),
			bytesProcessed:    bytesWritten,
		})
	})
	return bytesWritten, err
}

func (b *BackoffWriteCloser) Close() error {
	return b.writer.Close()
}
