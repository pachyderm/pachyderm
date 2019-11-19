package obj

import (
	"context"
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

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/tracing"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

// Environment variables for determining storage backend and pathing
const (
	StorageBackendEnvVar = "STORAGE_BACKEND"
	PachRootEnvVar       = "PACH_ROOT"
)

// Valid object storage backends
const (
	Minio     = "MINIO"
	Amazon    = "AMAZON"
	Google    = "GOOGLE"
	Microsoft = "MICROSOFT"
	Local     = "LOCAL"
)

// Google environment variables
const (
	GoogleBucketEnvVar = "GOOGLE_BUCKET"
	GoogleCredEnvVar   = "GOOGLE_CRED"
)

// Microsoft environment variables
const (
	MicrosoftContainerEnvVar = "MICROSOFT_CONTAINER"
	MicrosoftIDEnvVar        = "MICROSOFT_ID"
	MicrosoftSecretEnvVar    = "MICROSOFT_SECRET"
)

// Minio environment variables
const (
	MinioBucketEnvVar    = "MINIO_BUCKET"
	MinioEndpointEnvVar  = "MINIO_ENDPOINT"
	MinioIDEnvVar        = "MINIO_ID"
	MinioSecretEnvVar    = "MINIO_SECRET"
	MinioSecureEnvVar    = "MINIO_SECURE"
	MinioSignatureEnvVar = "MINIO_SIGNATURE"
)

// Amazon environment variables
const (
	AmazonRegionEnvVar       = "AMAZON_REGION"
	AmazonBucketEnvVar       = "AMAZON_BUCKET"
	AmazonIDEnvVar           = "AMAZON_ID"
	AmazonSecretEnvVar       = "AMAZON_SECRET"
	AmazonTokenEnvVar        = "AMAZON_TOKEN"
	AmazonVaultAddrEnvVar    = "AMAZON_VAULT_ADDR"
	AmazonVaultRoleEnvVar    = "AMAZON_VAULT_ROLE"
	AmazonVaultTokenEnvVar   = "AMAZON_VAULT_TOKEN"
	AmazonDistributionEnvVar = "AMAZON_DISTRIBUTION"
	CustomEndpointEnvVar     = "CUSTOM_ENDPOINT"
)

// Advanced configuration environment variables
const (
	RetriesEnvVar        = "RETRIES"
	TimeoutEnvVar        = "TIMEOUT"
	UploadACLEnvVar      = "UPLOAD_ACL"
	ReverseEnvVar        = "REVERSE"
	PartSizeEnvVar       = "PART_SIZE"
	MaxUploadPartsEnvVar = "MAX_UPLOAD_PARTS"
)

const (
	// DefaultRetries is the default number of retries for object storage requests.
	DefaultRetries = 10
	// DefaultTimeout is the default timeout for object storage requests.
	DefaultTimeout = "5m"
	// DefaultUploadACL is the default upload ACL for object storage uploads.
	DefaultUploadACL = "bucket-owner-full-control"
	// DefaultReverse is the default for whether to reverse object storage paths or not.
	DefaultReverse = true
	// DefaultPartSize is the default part size for object storage uploads.
	DefaultPartSize = 5242880
	// DefaultMaxUploadParts is the default maximum number of upload parts.
	DefaultMaxUploadParts = 10000
)

// AmazonAdvancedConfiguration contains the advanced configuration for the amazon client.
type AmazonAdvancedConfiguration struct {
	Retries int    `env:"RETRIES, default=10"`
	Timeout string `env:"TIMEOUT, default=5m"`
	// By default, objects uploaded to a bucket are only accessible to the
	// uploader, and not the owner of the bucket. Using the default ensures that
	// the owner of the bucket can access the objects as well.
	UploadACL      string `env:"UPLOAD_ACL, default=bucket-owner-full-control"`
	Reverse        bool   `env:"REVERSE, default=true"`
	PartSize       int64  `env:"PART_SIZE, default=5242880"`
	MaxUploadParts int    `env:"MAX_UPLOAD_PARTS, default=10000"`
}

// EnvVarToSecretKey is an environment variable name to secret key mapping
// This is being used to temporarily bridge the gap as we transition to a model
// where object storage access in the workers is based on environment variables
// and a library rather than mounting a secret to a sidecar container which
// accesses object storage
var EnvVarToSecretKey = []struct {
	Key   string
	Value string
}{
	{Key: GoogleBucketEnvVar, Value: "google-bucket"},
	{Key: GoogleCredEnvVar, Value: "google-cred"},
	{Key: MicrosoftContainerEnvVar, Value: "microsoft-container"},
	{Key: MicrosoftIDEnvVar, Value: "microsoft-id"},
	{Key: MicrosoftSecretEnvVar, Value: "microsoft-secret"},
	{Key: MinioBucketEnvVar, Value: "minio-bucket"},
	{Key: MinioEndpointEnvVar, Value: "minio-endpoint"},
	{Key: MinioIDEnvVar, Value: "minio-id"},
	{Key: MinioSecretEnvVar, Value: "minio-secret"},
	{Key: MinioSecureEnvVar, Value: "minio-secure"},
	{Key: MinioSignatureEnvVar, Value: "minio-signature"},
	{Key: AmazonRegionEnvVar, Value: "amazon-region"},
	{Key: AmazonBucketEnvVar, Value: "amazon-bucket"},
	{Key: AmazonIDEnvVar, Value: "amazon-id"},
	{Key: AmazonSecretEnvVar, Value: "amazon-secret"},
	{Key: AmazonTokenEnvVar, Value: "amazon-token"},
	{Key: AmazonVaultAddrEnvVar, Value: "amazon-vault-addr"},
	{Key: AmazonVaultRoleEnvVar, Value: "amazon-vault-role"},
	{Key: AmazonVaultTokenEnvVar, Value: "amazon-vault-token"},
	{Key: AmazonDistributionEnvVar, Value: "amazon-distribution"},
	{Key: CustomEndpointEnvVar, Value: "custom-endpoint"},
	{Key: RetriesEnvVar, Value: "retries"},
	{Key: TimeoutEnvVar, Value: "timeout"},
	{Key: UploadACLEnvVar, Value: "upload-acl"},
	{Key: ReverseEnvVar, Value: "reverse"},
	{Key: PartSizeEnvVar, Value: "part-size"},
	{Key: MaxUploadPartsEnvVar, Value: "max-upload-parts"},
}

// StorageRootFromEnv gets the storage root based on environment variables.
func StorageRootFromEnv() (string, error) {
	storageRoot, ok := os.LookupEnv(PachRootEnvVar)
	if !ok {
		return "", fmt.Errorf("%s not found", PachRootEnvVar)
	}
	storageBackend, ok := os.LookupEnv(StorageBackendEnvVar)
	if !ok {
		return "", fmt.Errorf("%s not found", StorageBackendEnvVar)
	}
	// These storage backends do not like leading slashes
	switch storageBackend {
	case Amazon:
		fallthrough
	case Minio:
		if len(storageRoot) > 0 && storageRoot[0] == '/' {
			storageRoot = storageRoot[1:]
		}
	}
	return storageRoot, nil
}

// BlockPathFromEnv gets the path to an object storage block based on environment variables.
func BlockPathFromEnv(block *pfs.Block) (string, error) {
	storageRoot, err := StorageRootFromEnv()
	if err != nil {
		return "", err
	}
	return filepath.Join(storageRoot, "block", block.Hash), nil
}

// Client is an interface to object storage.
type Client interface {
	// Writer returns a writer which writes to an object.
	// It should error if the object already exists or we don't have sufficient
	// permissions to write it.
	Writer(ctx context.Context, name string) (io.WriteCloser, error)
	// Reader returns a reader which reads from an object.
	// If `size == 0`, the reader should read from the offset till the end of the object.
	// It should error if the object doesn't exist or we don't have sufficient
	// permission to read it.
	Reader(ctx context.Context, name string, offset uint64, size uint64) (io.ReadCloser, error)
	// Delete deletes an object.
	// It should error if the object doesn't exist or we don't have sufficient
	// permission to delete it.
	Delete(ctx context.Context, name string) error
	// Walk calls `fn` with the names of objects which can be found under `prefix`.
	Walk(ctx context.Context, prefix string, fn func(name string) error) error
	// Exsits checks if a given object already exists
	Exists(ctx context.Context, name string) bool
	// IsRetryable determines if an operation should be retried given an error
	IsRetryable(err error) bool
	// IsNotExist returns true if err is a non existence error
	IsNotExist(err error) bool
	// IsIgnorable returns true if the error can be ignored
	IsIgnorable(err error) bool
}

// NewGoogleClient creates a google client with the given bucket name.
func NewGoogleClient(bucket string, opts []option.ClientOption) (Client, error) {
	return newGoogleClient(bucket, opts)
}

func secretFile(name string) string {
	return filepath.Join("/", client.StorageSecretName, name)
}

func readSecretFile(name string) (string, error) {
	bytes, err := ioutil.ReadFile(secretFile(name))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(bytes)), nil
}

// NewGoogleClientFromSecret creates a google client by reading credentials
// from a mounted GoogleSecret. You may pass "" for bucket in which case it
// will read the bucket from the secret.
func NewGoogleClientFromSecret(bucket string) (Client, error) {
	var err error
	if bucket == "" {
		bucket, err = readSecretFile("/google-bucket")
		if err != nil {
			return nil, fmt.Errorf("google-bucket not found")
		}
	}
	cred, err := readSecretFile("/google-cred")
	if err != nil {
		return nil, fmt.Errorf("google-cred not found")
	}
	var opts []option.ClientOption
	if cred != "" {
		opts = append(opts, option.WithCredentialsFile(secretFile("/google-cred")))
	} else {
		opts = append(opts, option.WithTokenSource(google.ComputeTokenSource("")))
	}
	return NewGoogleClient(bucket, opts)
}

// NewGoogleClientFromEnv creates a Google client based on environment variables.
func NewGoogleClientFromEnv() (Client, error) {
	bucket, ok := os.LookupEnv(GoogleBucketEnvVar)
	if !ok {
		return nil, fmt.Errorf("%s not found", GoogleBucketEnvVar)
	}
	creds, ok := os.LookupEnv(GoogleCredEnvVar)
	if !ok {
		return nil, fmt.Errorf("%s not found", GoogleCredEnvVar)
	}
	opts := []option.ClientOption{option.WithCredentialsJSON([]byte(creds))}
	return NewGoogleClient(bucket, opts)
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

// NewMicrosoftClientFromEnv creates a Microsoft client based on environment variables.
func NewMicrosoftClientFromEnv() (Client, error) {
	container, ok := os.LookupEnv(MicrosoftContainerEnvVar)
	if !ok {
		return nil, fmt.Errorf("%s not found", MicrosoftContainerEnvVar)
	}
	id, ok := os.LookupEnv(MicrosoftIDEnvVar)
	if !ok {
		return nil, fmt.Errorf("%s not found", MicrosoftIDEnvVar)
	}
	secret, ok := os.LookupEnv(MicrosoftSecretEnvVar)
	if !ok {
		return nil, fmt.Errorf("%s not found", MicrosoftSecretEnvVar)
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
//   endpoint - Custom endpoint (generally used for S3 compatible object stores)
//   reverse - Reverse object storage paths (overwrites configured value)
func NewAmazonClient(region, bucket string, creds *AmazonCreds, distribution string, endpoint string, reverse ...bool) (Client, error) {
	advancedConfig := &AmazonAdvancedConfiguration{}
	if err := cmdutil.Populate(advancedConfig); err != nil {
		return nil, err
	}
	if len(reverse) > 0 {
		advancedConfig.Reverse = reverse[0]
	}
	return newAmazonClient(region, bucket, creds, distribution, endpoint, advancedConfig)
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

// NewMinioClientFromEnv creates a Minio client based on environment variables.
func NewMinioClientFromEnv() (Client, error) {
	bucket, ok := os.LookupEnv(MinioBucketEnvVar)
	if !ok {
		return nil, fmt.Errorf("%s not found", MinioBucketEnvVar)
	}
	endpoint, ok := os.LookupEnv(MinioEndpointEnvVar)
	if !ok {
		return nil, fmt.Errorf("%s not found", MinioEndpointEnvVar)
	}
	id, ok := os.LookupEnv(MinioIDEnvVar)
	if !ok {
		return nil, fmt.Errorf("%s not found", MinioIDEnvVar)
	}
	secret, ok := os.LookupEnv(MinioSecretEnvVar)
	if !ok {
		return nil, fmt.Errorf("%s not found", MinioSecretEnvVar)
	}
	secure, ok := os.LookupEnv(MinioSecureEnvVar)
	if !ok {
		return nil, fmt.Errorf("%s not found", MinioSecureEnvVar)
	}
	isS3V2, ok := os.LookupEnv(MinioSignatureEnvVar)
	if !ok {
		return nil, fmt.Errorf("%s not found", MinioSignatureEnvVar)
	}
	return NewMinioClient(endpoint, bucket, id, secret, secure == "1", isS3V2 == "1")
}

// NewAmazonClientFromSecret constructs an amazon client by reading credentials
// from a mounted AmazonSecret. You may pass "" for bucket in which case it
// will read the bucket from the secret.
func NewAmazonClientFromSecret(bucket string, reverse ...bool) (Client, error) {
	// Get AWS region (required for constructing an AWS client)
	region, err := readSecretFile("/amazon-region")
	if err != nil {
		return nil, fmt.Errorf("amazon-region not found")
	}

	// Use or retrieve S3 bucket
	if bucket == "" {
		bucket, err = readSecretFile("/amazon-bucket")
		if err != nil {
			return nil, err
		}
	}

	// Retrieve either static or vault credentials; if neither are found, we will
	// use IAM roles (i.e. the EC2 metadata service)
	var creds AmazonCreds
	creds.ID, err = readSecretFile("/amazon-id")
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	creds.Secret, err = readSecretFile("/amazon-secret")
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	creds.Token, err = readSecretFile("/amazon-token")
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	creds.VaultAddress, err = readSecretFile("/amazon-vault-addr")
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	creds.VaultRole, err = readSecretFile("/amazon-vault-role")
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	creds.VaultToken, err = readSecretFile("/amazon-vault-token")
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	// Get Cloudfront distribution (not required, though we can log a warning)
	distribution, err := readSecretFile("/amazon-distribution")
	// Get endpoint for custom deployment (optional).
	endpoint, err := readSecretFile("/custom-endpoint")
	return NewAmazonClient(region, bucket, &creds, distribution, endpoint, reverse...)
}

// NewAmazonClientFromEnv creates a Amazon client based on environment variables.
func NewAmazonClientFromEnv() (Client, error) {
	region, ok := os.LookupEnv(AmazonRegionEnvVar)
	if !ok {
		return nil, fmt.Errorf("%s not found", AmazonRegionEnvVar)
	}
	bucket, ok := os.LookupEnv(AmazonBucketEnvVar)
	if !ok {
		return nil, fmt.Errorf("%s not found", AmazonBucketEnvVar)
	}

	var creds AmazonCreds
	creds.ID, _ = os.LookupEnv(AmazonIDEnvVar)
	creds.Secret, _ = os.LookupEnv(AmazonSecretEnvVar)
	creds.Token, _ = os.LookupEnv(AmazonTokenEnvVar)
	creds.VaultAddress, _ = os.LookupEnv(AmazonVaultAddrEnvVar)
	creds.VaultRole, _ = os.LookupEnv(AmazonVaultRoleEnvVar)
	creds.VaultToken, _ = os.LookupEnv(AmazonVaultTokenEnvVar)

	distribution, _ := os.LookupEnv(AmazonDistributionEnvVar)
	// Get endpoint for custom deployment (optional).
	endpoint, _ := os.LookupEnv(CustomEndpointEnvVar)
	return NewAmazonClient(region, bucket, &creds, distribution, endpoint)
}

// NewClientFromURLAndSecret constructs a client by parsing `URL` and then
// constructing the correct client for that URL using secrets.
func NewClientFromURLAndSecret(url *ObjectStoreURL, reverse ...bool) (c Client, err error) {
	switch url.Store {
	case "s3":
		c, err = NewAmazonClientFromSecret(url.Bucket, reverse...)
	case "gcs":
		fallthrough
	case "gs":
		c, err = NewGoogleClientFromSecret(url.Bucket)
	case "as":
		fallthrough
	case "wasb":
		// In Azure, the first part of the path is the container name.
		c, err = NewMicrosoftClientFromSecret(url.Bucket)
	case "local":
		c, err = NewLocalClient("/" + url.Bucket)
	}
	switch {
	case err != nil:
		return nil, err
	case c != nil:
		return TracingObjClient(url.Store, c), nil
	default:
		return nil, fmt.Errorf("unrecognized object store: %s", url.Bucket)
	}
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
	case "s3", "gcs", "gs", "local":
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

// NewClientFromEnv creates a client based on environment variables.
func NewClientFromEnv(storageRoot string) (c Client, err error) {
	storageBackend, ok := os.LookupEnv(StorageBackendEnvVar)
	if !ok {
		return nil, fmt.Errorf("storage backend environment variable not found")
	}
	switch storageBackend {
	case Amazon:
		c, err = NewAmazonClientFromEnv()
	case Google:
		c, err = NewGoogleClientFromEnv()
	case Microsoft:
		c, err = NewMicrosoftClientFromEnv()
	case Minio:
		c, err = NewMinioClientFromEnv()
	case Local:
		c, err = NewLocalClient(storageRoot)
	}
	switch {
	case err != nil:
		return nil, err
	case c != nil:
		return TracingObjClient(storageBackend, c), nil
	default:
		return nil, fmt.Errorf("unrecognized storage backend: %s", storageBackend)
	}
}

// NewClientFromSecret creates a client based on mounted secret files.
func NewClientFromSecret(storageRoot string) (c Client, err error) {
	storageBackend, ok := os.LookupEnv(StorageBackendEnvVar)
	if !ok {
		return nil, fmt.Errorf("storage backend environment variable not found")
	}
	switch storageBackend {
	case Amazon:
		c, err = NewAmazonClientFromSecret("")
	case Google:
		c, err = NewGoogleClientFromSecret("")
	case Microsoft:
		c, err = NewMicrosoftClientFromSecret("")
	case Minio:
		c, err = NewMinioClientFromSecret("")
	case Local:
		c, err = NewLocalClient(storageRoot)
	}
	switch {
	case err != nil:
		return nil, err
	case c != nil:
		return TracingObjClient(storageBackend, c), nil
	default:
		return nil, fmt.Errorf("unrecognized storage backend: %s", storageBackend)
	}
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
	ctx           context.Context
	client        Client
	reader        io.ReadCloser
	backoffConfig *backoff.ExponentialBackOff
	object        string
}

func newBackoffReadCloser(ctx context.Context, client Client, reader io.ReadCloser) io.ReadCloser {
	return &BackoffReadCloser{
		ctx:           ctx,
		client:        client,
		reader:        reader,
		backoffConfig: NewExponentialBackOffConfig(),
	}
}

func (b *BackoffReadCloser) Read(data []byte) (retN int, retErr error) {
	span, _ := tracing.AddSpanToAnyExisting(b.ctx, "/obj.BackoffReadCloser/Read")
	defer func() {
		tracing.FinishAnySpan(span, "bytes", retN, "err", retErr)
	}()
	bytesRead := 0
	var n int
	var err error
	backoff.RetryNotify(func() error {
		start := time.Now()
		fmt.Printf("reading %d bytes from object: %q @ %v\n", len(data[bytesRead:]), b.object, start)
		n, err = b.reader.Read(data[bytesRead:])
		fmt.Printf("finished reading from object: %q after %v with error: %v\n", b.object, time.Since(start), err)
		bytesRead += n
		if err != nil && IsRetryable(b.client, err) {
			return err
		}
		return nil
	}, b.backoffConfig, func(err error, d time.Duration) error {
		log.Infof("Error reading; retrying in %s: %#v", d, RetryError{
			Err:               err.Error(),
			TimeTillNextRetry: d.String(),
			BytesProcessed:    bytesRead,
		})
		return nil
	})
	return bytesRead, err
}

// Close closes the ReaderCloser contained in b.
func (b *BackoffReadCloser) Close() (retErr error) {
	span, _ := tracing.AddSpanToAnyExisting(b.ctx, "/obj.BackoffReadCloser/Close")
	defer func() {
		tracing.FinishAnySpan(span, "err", retErr)
	}()
	return b.reader.Close()
}

// BackoffWriteCloser retries with exponential backoff in the case of failures
type BackoffWriteCloser struct {
	ctx           context.Context
	client        Client
	writer        io.WriteCloser
	backoffConfig *backoff.ExponentialBackOff
}

func newBackoffWriteCloser(ctx context.Context, client Client, writer io.WriteCloser) io.WriteCloser {
	return &BackoffWriteCloser{
		ctx:           ctx,
		client:        client,
		writer:        writer,
		backoffConfig: NewExponentialBackOffConfig(),
	}
}

func (b *BackoffWriteCloser) Write(data []byte) (retN int, retErr error) {
	span, _ := tracing.AddSpanToAnyExisting(b.ctx, "/obj.BackoffWriteCloser/Write")
	defer func() {
		tracing.FinishAnySpan(span, "bytes", retN, "err", retErr)
	}()
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
	}, b.backoffConfig, func(err error, d time.Duration) error {
		log.Infof("Error writing; retrying in %s: %#v", d, RetryError{
			Err:               err.Error(),
			TimeTillNextRetry: d.String(),
			BytesProcessed:    bytesWritten,
		})
		return nil
	})
	return bytesWritten, err
}

// Close closes the WriteCloser contained in b.
func (b *BackoffWriteCloser) Close() (retErr error) {
	span, _ := tracing.AddSpanToAnyExisting(b.ctx, "/obj.BackoffWriteCloser/Close")
	defer func() {
		tracing.FinishAnySpan(span, "err", retErr)
	}()
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

// TestStorage is a defensive method for checking to make sure that storage is
// properly configured.
func TestStorage(ctx context.Context, c Client) error {
	testObj := uuid.NewWithoutDashes()
	if err := func() (retErr error) {
		w, err := c.Writer(ctx, testObj)
		if err != nil {
			return err
		}
		defer func() {
			if err := w.Close(); err != nil && retErr == nil {
				retErr = err
			}
		}()
		_, err = w.Write([]byte("test"))
		return err
	}(); err != nil {
		return fmt.Errorf("unable to write to object storage: %v", err)
	}
	if err := func() (retErr error) {
		r, err := c.Reader(ctx, testObj, 0, 0)
		if err != nil {
			return err
		}
		defer func() {
			if err := r.Close(); err != nil && retErr == nil {
				retErr = err
			}
		}()
		_, err = ioutil.ReadAll(r)
		return err
	}(); err != nil {
		return fmt.Errorf("unable to read from object storage: %v", err)
	}
	if err := c.Delete(ctx, testObj); err != nil {
		return fmt.Errorf("unable to delete from object storage: %v", err)
	}
	// Try reading a non-existant object to make sure our IsNotExist function
	// works.
	_, err := c.Reader(ctx, uuid.NewWithoutDashes(), 0, 0)
	if !c.IsNotExist(err) {
		return fmt.Errorf("storage is unable to discern NotExist errors, \"%s\" should count as NotExist", err.Error())
	}
	return nil
}
