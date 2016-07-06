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
	// Exsits checks if a given object already exists
	Exists(name string) bool
	// IsRetryable determines if an operation should be retried given an error
	IsRetryable(err error) bool
	// IsNotExist returns true if err is a non existence error
	IsNotExist(err error) bool
	// IsIgnorable returns true if the error can be ignored
	IsIgnorable(err error) bool
}

// NewGoogleClient creates a google client with the given bucket name.
func NewGoogleClient(ctx context.Context, bucket string) (Client, error) {
	return newGoogleClient(ctx, bucket)
}

// NewAmazonClient creates an amazon client with the following credentials:
//   bucket - S3 bucket name
//   id     - AWS access key id
//   secret - AWS secret access key
//   token  - AWS access token
//   region - AWS region
func NewAmazonClient(bucket string, id string, secret string, token string,
	region string) (Client, error) {
	return newAmazonClient(bucket, id, secret, token, region)
}

// NewExponentialBackOffConfig creates an exponential back-off config with
// longer wait times than the default.
func NewExponentialBackOffConfig() *backoff.ExponentialBackOff {
	config := backoff.NewExponentialBackOff()
	// We want to backoff more aggressively (i.e. wait longer) than the default
	config.InitialInterval = 1 * time.Second
	config.Multiplier = 2
	config.MaxElapsedTime = 5 * time.Minute
	return config
}

// RetryError is used to log retry attempts.
type RetryError struct {
	Err               string
	TimeTillNextRetry string
	BytesProcessed    int
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
		backoffConfig: NewExponentialBackOffConfig(),
	}
}

func (b *BackoffReadCloser) Read(data []byte) (int, error) {
	bytesRead := 0
	var n int
	var err error
	backoff.RetryNotify(func() error {
		n, err = b.reader.Read(data[bytesRead:])
		bytesRead += n
		if err != nil && b.client.IsRetryable(err) {
			return err
		}
		return nil
	}, b.backoffConfig, func(err error, d time.Duration) {
		protolion.Infof("Error reading; retrying in %s: %#v", d, RetryError{
			Err:               err.Error(),
			TimeTillNextRetry: d.String(),
			BytesProcessed:    bytesRead,
		})
	})
	return bytesRead, err
}

// Close closes the ReaderCloser contained in b.
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
		backoffConfig: NewExponentialBackOffConfig(),
	}
}

func (b *BackoffWriteCloser) Write(data []byte) (int, error) {
	bytesWritten := 0
	var n int
	var err error
	backoff.RetryNotify(func() error {
		n, err = b.writer.Write(data[bytesWritten:])
		bytesWritten += n
		if err != nil && b.client.IsRetryable(err) {
			return err
		}
		return nil
	}, b.backoffConfig, func(err error, d time.Duration) {
		protolion.Infof("Error writing; retrying in %s: %#v", d, RetryError{
			Err:               err.Error(),
			TimeTillNextRetry: d.String(),
			BytesProcessed:    bytesWritten,
		})
	})
	return bytesWritten, err
}

// Close closes the WriteCloser contained in b.
func (b *BackoffWriteCloser) Close() error {
	err := b.writer.Close()
	if b.client.IsIgnorable(err) {
		return nil
	}
	return err
}
