package obj

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	log "github.com/sirupsen/logrus"
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

func readSecretFile(name string) (string, error) {
	bytes, err := ioutil.ReadFile(filepath.Join("/", client.StorageSecretName, name))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(bytes)), nil
}

// NewGoogleClientFromSecret creates a google client by reading credentials
// from a mounted GoogleSecret. You may pass "" for bucket in which case it
// will read the bucket from the secret.
func NewGoogleClientFromSecret(ctx context.Context, bucket string) (Client, error) {
	var err error
	if bucket == "" {
		bucket, err = readSecretFile("/google-bucket")
		if err != nil {
			return nil, fmt.Errorf("google-bucket not found")
		}
	}
	return NewGoogleClient(ctx, bucket)
}

// NewMicrosoftClient creates a microsoft client:
//	container   - Azure Blob Container name
//	accountName - Azure Storage Account name
// 	accountKey  - Azure Storage Account key
func NewMicrosoftClient(container string, accountName string, accountKey string) (Client, error) {
	return newMicrosoftClient(container, accountName, accountKey)
}

// NewMicrosoftClientFromSecret creates a microsoft client by reading
// credentials from a mounted MicrosoftSecret. You may pass "" for container in
// which case it will read the container from the secret.
func NewMicrosoftClientFromSecret(container string) (Client, error) {
	var err error
	if container == "" {
		container, err = readSecretFile("/microsoft-container")
		if err != nil {
			return nil, fmt.Errorf("microsoft-container not found")
		}
	}
	id, err := readSecretFile("/microsoft-id")
	if err != nil {
		return nil, fmt.Errorf("microsoft-id not found")
	}
	secret, err := readSecretFile("/microsoft-secret")
	if err != nil {
		return nil, fmt.Errorf("microsoft-secret not found")
	}
	return NewMicrosoftClient(container, id, secret)
}

// NewMinioClient creates an s3 compatible client with the following credentials:
//   endpoint - S3 compatible endpoint
//   bucket - S3 bucket name
//   id     - AWS access key id
//   secret - AWS secret access key
//   secure - Set to true if connection is secure.
//   isS3V2 - Set to true if client follows S3V2
func NewMinioClient(endpoint, bucket, id, secret string, secure, isS3V2 bool) (Client, error) {
	if isS3V2 {
		return newMinioClientV2(endpoint, bucket, id, secret, secure)
	}
	return newMinioClient(endpoint, bucket, id, secret, secure)
}

// NewAmazonClient creates an amazon client with the following credentials:
//   bucket - S3 bucket name
//   distribution - cloudfront distribution ID
//   id     - AWS access key id
//   secret - AWS secret access key
//   token  - AWS access token
//   region - AWS region
func NewAmazonClient(bucket string, distribution string, id string, secret string, token string,
	region string) (Client, error) {
	return newAmazonClient(bucket, distribution, id, secret, token, region)
}

// NewMinioClientFromSecret constructs an s3 compatible client by reading
// credentials from a mounted AmazonSecret. You may pass "" for bucket in which case it
// will read the bucket from the secret.
func NewMinioClientFromSecret(bucket string) (Client, error) {
	var err error
	if bucket == "" {
		bucket, err = readSecretFile("/minio-bucket")
		if err != nil {
			return nil, err
		}
	}
	endpoint, err := readSecretFile("/minio-endpoint")
	if err != nil {
		return nil, err
	}
	id, err := readSecretFile("/minio-id")
	if err != nil {
		return nil, err
	}
	secret, err := readSecretFile("/minio-secret")
	if err != nil {
		return nil, err
	}
	secure, err := readSecretFile("/minio-secure")
	if err != nil {
		return nil, err
	}
	isS3V2, err := readSecretFile("/minio-signature")
	if err != nil {
		return nil, err
	}
	return NewMinioClient(endpoint, bucket, id, secret, secure == "1", isS3V2 == "1")
}

// NewAmazonClientFromSecret constructs an amazon client by reading credentials
// from a mounted AmazonSecret. You may pass "" for bucket in which case it
// will read the bucket from the secret.
func NewAmazonClientFromSecret(bucket string) (Client, error) {
	var distribution string
	var err error
	if bucket == "" {
		bucket, err = readSecretFile("/amazon-bucket")
		if err != nil {
			return nil, err
		}
		distribution, err = readSecretFile("/amazon-distribution")
		if err != nil {
			// Distribution is not required, but we can log a warning
			log.Warnln("AWS deployed without cloudfront distribution\n")
		} else {
			log.Infof("AWS deployed with cloudfront distribution at %v\n", string(distribution))
		}
	}
	// It's ok if we can't find static credentials; we will use IAM roles
	// in that case.
	id, err := readSecretFile("/amazon-id")
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	secret, err := readSecretFile("/amazon-secret")
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	token, err := readSecretFile("/amazon-token")
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	// region is required for constructing an AWS client
	region, err := readSecretFile("/amazon-region")
	if err != nil {
		return nil, fmt.Errorf("amazon-region not found")
	}
	return NewAmazonClient(bucket, distribution, id, secret, token, region)
}

// NewClientFromURLAndSecret constructs a client by parsing `URL` and then
// constructing the correct client for that URL using secrets.
func NewClientFromURLAndSecret(ctx context.Context, url *ObjectStoreURL) (Client, error) {
	switch url.Store {
	case "s3":
		return NewAmazonClientFromSecret(url.Bucket)
	case "gcs":
		fallthrough
	case "gs":
		return NewGoogleClientFromSecret(ctx, url.Bucket)
	case "as":
		fallthrough
	case "wasb":
		// In Azure, the first part of the path is the container name.
		return NewMicrosoftClientFromSecret(url.Bucket)
	}
	return nil, fmt.Errorf("unrecognized object store: %s", url.Bucket)
}

// ObjectStoreURL represents a parsed URL to an object in an object store.
type ObjectStoreURL struct {
	// The object store, e.g. s3, gcs, as...
	Store string
	// The "bucket" (in AWS parlance) or the "container" (in Azure parlance).
	Bucket string
	// The object itself.
	Object string
}

// ParseURL parses an URL into ObjectStoreURL.
func ParseURL(urlStr string) (*ObjectStoreURL, error) {
	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing url %v: %v", urlStr, err)
	}
	switch url.Scheme {
	case "s3", "gcs", "gs":
		return &ObjectStoreURL{
			Store:  url.Scheme,
			Bucket: url.Host,
			Object: strings.Trim(url.Path, "/"),
		}, nil
	case "as", "wasb":
		// In Azure, the first part of the path is the container name.
		parts := strings.Split(strings.Trim(url.Path, "/"), "/")
		if len(parts) < 1 {
			return nil, fmt.Errorf("malformed Azure URI: %v", urlStr)
		}
		return &ObjectStoreURL{
			Store:  url.Scheme,
			Bucket: parts[0],
			Object: strings.Trim(path.Join(parts[1:]...), "/"),
		}, nil
	}
	return nil, fmt.Errorf("unrecognized object store: %s", url.Scheme)
}

// NewExponentialBackOffConfig creates an exponential back-off config with
// longer wait times than the default.
func NewExponentialBackOffConfig() *backoff.ExponentialBackOff {
	config := backoff.NewExponentialBackOff()
	// We want to backoff more aggressively (i.e. wait longer) than the default
	config.InitialInterval = 1 * time.Second
	config.Multiplier = 2
	config.MaxInterval = 15 * time.Minute
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
		if err != nil && IsRetryable(b.client, err) {
			return err
		}
		return nil
	}, b.backoffConfig, func(err error, d time.Duration) {
		log.Infof("Error reading; retrying in %s: %#v", d, RetryError{
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
		if err != nil && IsRetryable(b.client, err) {
			return err
		}
		return nil
	}, b.backoffConfig, func(err error, d time.Duration) {
		log.Infof("Error writing; retrying in %s: %#v", d, RetryError{
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

// IsRetryable determines if an operation should be retried given an error
func IsRetryable(client Client, err error) bool {
	return isNetRetryable(err) || client.IsRetryable(err)
}

func byteRange(offset uint64, size uint64) string {
	if offset == 0 && size == 0 {
		return ""
	} else if size == 0 {
		return fmt.Sprintf("%d-", offset)
	}
	return fmt.Sprintf("%d-%d", offset, offset+size-1)
}

func isNetRetryable(err error) bool {
	netErr, ok := err.(net.Error)
	return ok && netErr.Temporary()
}

// TestIsNotExist is a defensive method for checking to make sure IsNotExist is
// satisfying its semantics.
func TestIsNotExist(c Client) error {
	_, err := c.Reader(uuid.NewWithoutDashes(), 0, 0)
	if !c.IsNotExist(err) {
		return fmt.Errorf("storage is unable to discern NotExist errors, \"%s\" should count as NotExist", err.Error())
	}
	return nil
}
