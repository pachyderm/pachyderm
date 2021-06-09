package obj

import (
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"

	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

// Environment variables for determining storage backend and pathing
const (
	StorageBackendEnvVar = "STORAGE_BACKEND"
)

// StorageRootFromEnv gets the storage root based on environment variables.
func StorageRootFromEnv(storageRoot string) (string, error) {
	storageBackend, ok := os.LookupEnv(StorageBackendEnvVar)
	if !ok {
		return "", errors.Errorf("%s not found", StorageBackendEnvVar)
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
	DisableSSLEnvVar     = "DISABLE_SSL"
	NoVerifySSLEnvVar    = "NO_VERIFY_SSL"
	LogOptionsEnvVar     = "OBJ_LOG_OPTS"
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
	// DefaultDisableSSL is the default for whether SSL should be disabled.
	DefaultDisableSSL = false
	// DefaultNoVerifySSL is the default for whether SSL certificate verification should be disabled.
	DefaultNoVerifySSL = false
	// DefaultAwsLogOptions is the default set of enabled S3 client log options
	DefaultAwsLogOptions = ""
)

// AmazonAdvancedConfiguration contains the advanced configuration for the amazon client.
type AmazonAdvancedConfiguration struct {
	Retries int    `env:"RETRIES, default=10"`
	Timeout string `env:"TIMEOUT, default=5m"`
	// By default, objects uploaded to a bucket are only accessible to the
	// uploader, and not the owner of the bucket. Using the default ensures that
	// the owner of the bucket can access the objects as well.
	UploadACL      string `env:"UPLOAD_ACL, default=bucket-owner-full-control"`
	PartSize       int64  `env:"PART_SIZE, default=5242880"`
	MaxUploadParts int    `env:"MAX_UPLOAD_PARTS, default=10000"`
	DisableSSL     bool   `env:"DISABLE_SSL, default=false"`
	NoVerifySSL    bool   `env:"NO_VERIFY_SSL, default=false"`
	LogOptions     string `env:"OBJ_LOG_OPTS, default="`
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
	{Key: AmazonDistributionEnvVar, Value: "amazon-distribution"},
	{Key: CustomEndpointEnvVar, Value: "custom-endpoint"},
	{Key: RetriesEnvVar, Value: "retries"},
	{Key: TimeoutEnvVar, Value: "timeout"},
	{Key: UploadACLEnvVar, Value: "upload-acl"},
	{Key: ReverseEnvVar, Value: "reverse"},
	{Key: PartSizeEnvVar, Value: "part-size"},
	{Key: MaxUploadPartsEnvVar, Value: "max-upload-parts"},
	{Key: DisableSSLEnvVar, Value: "disable-ssl"},
	{Key: NoVerifySSLEnvVar, Value: "no-verify-ssl"},
	{Key: LogOptionsEnvVar, Value: "log-options"},
}

// NewGoogleClient creates a google client with the given bucket name.
func NewGoogleClient(bucket string, opts []option.ClientOption) (c Client, err error) {
	if c, err = newGoogleClient(bucket, opts); err != nil {
		return nil, err
	}
	return newUniformClient(c), nil
}

func secretFile(name string) string {
	return filepath.Join("/", "pachyderm-storage-secret", name)
}

func readSecretFile(name string) (string, error) {
	bytes, err := ioutil.ReadFile(secretFile(name))
	if err != nil {
		return "", errors.EnsureStack(err)
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
			return nil, errors.Errorf("google-bucket not found")
		}
	}
	cred, err := readSecretFile("/google-cred")
	if err != nil {
		return nil, errors.Errorf("google-cred not found")
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
		return nil, errors.Errorf("%s not found", GoogleBucketEnvVar)
	}
	creds, ok := os.LookupEnv(GoogleCredEnvVar)
	if !ok {
		return nil, errors.Errorf("%s not found", GoogleCredEnvVar)
	}
	opts := []option.ClientOption{option.WithCredentialsJSON([]byte(creds))}
	return NewGoogleClient(bucket, opts)
}

// NewMicrosoftClient creates a microsoft client:
//	container   - Azure Blob Container name
//	accountName - Azure Storage Account name
// 	accountKey  - Azure Storage Account key
func NewMicrosoftClient(container string, accountName string, accountKey string) (c Client, err error) {
	c, err = newMicrosoftClient(container, accountName, accountKey)
	if err != nil {
		return nil, err
	}
	return newUniformClient(c), nil
}

// NewMicrosoftClientFromSecret creates a microsoft client by reading
// credentials from a mounted MicrosoftSecret. You may pass "" for container in
// which case it will read the container from the secret.
func NewMicrosoftClientFromSecret(container string) (Client, error) {
	var err error
	if container == "" {
		container, err = readSecretFile("/microsoft-container")
		if err != nil {
			return nil, errors.Errorf("microsoft-container not found")
		}
	}
	id, err := readSecretFile("/microsoft-id")
	if err != nil {
		return nil, errors.Errorf("microsoft-id not found")
	}
	secret, err := readSecretFile("/microsoft-secret")
	if err != nil {
		return nil, errors.Errorf("microsoft-secret not found")
	}
	return NewMicrosoftClient(container, id, secret)
}

// NewMicrosoftClientFromEnv creates a Microsoft client based on environment variables.
func NewMicrosoftClientFromEnv() (Client, error) {
	container, ok := os.LookupEnv(MicrosoftContainerEnvVar)
	if !ok {
		return nil, errors.Errorf("%s not found", MicrosoftContainerEnvVar)
	}
	id, ok := os.LookupEnv(MicrosoftIDEnvVar)
	if !ok {
		return nil, errors.Errorf("%s not found", MicrosoftIDEnvVar)
	}
	secret, ok := os.LookupEnv(MicrosoftSecretEnvVar)
	if !ok {
		return nil, errors.Errorf("%s not found", MicrosoftSecretEnvVar)
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
func NewMinioClient(endpoint, bucket, id, secret string, secure, isS3V2 bool) (c Client, err error) {
	log.Warnf("DEPRECATED: Support for the S3V2 option is being deprecated. It will be removed in a future version")
	if isS3V2 {
		return newMinioClientV2(endpoint, bucket, id, secret, secure)
	}
	c, err = newMinioClient(endpoint, bucket, id, secret, secure)
	if err != nil {
		return nil, err
	}
	return newUniformClient(c), nil
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
func NewAmazonClient(region, bucket string, creds *AmazonCreds, distribution string, endpoint string) (c Client, err error) {
	advancedConfig := &AmazonAdvancedConfiguration{}
	if err := cmdutil.Populate(advancedConfig); err != nil {
		return nil, errors.EnsureStack(err)
	}
	c, err = newAmazonClient(region, bucket, creds, distribution, endpoint, advancedConfig)
	if err != nil {
		return nil, err
	}
	return newUniformClient(c), nil
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
		return nil, errors.Errorf("%s not found", MinioBucketEnvVar)
	}
	endpoint, ok := os.LookupEnv(MinioEndpointEnvVar)
	if !ok {
		return nil, errors.Errorf("%s not found", MinioEndpointEnvVar)
	}
	id, ok := os.LookupEnv(MinioIDEnvVar)
	if !ok {
		return nil, errors.Errorf("%s not found", MinioIDEnvVar)
	}
	secret, ok := os.LookupEnv(MinioSecretEnvVar)
	if !ok {
		return nil, errors.Errorf("%s not found", MinioSecretEnvVar)
	}
	secure, ok := os.LookupEnv(MinioSecureEnvVar)
	if !ok {
		return nil, errors.Errorf("%s not found", MinioSecureEnvVar)
	}
	isS3V2, ok := os.LookupEnv(MinioSignatureEnvVar)
	if !ok {
		return nil, errors.Errorf("%s not found", MinioSignatureEnvVar)
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
		return nil, errors.Errorf("amazon-region not found")
	}

	// Use or retrieve S3 bucket
	if bucket == "" {
		bucket, err = readSecretFile("/amazon-bucket")
		if err != nil {
			return nil, err
		}
	}

	// Retrieve static credentials; if not found,
	// use IAM roles (i.e. the EC2 metadata service)
	var creds AmazonCreds
	creds.ID, err = readSecretFile("/amazon-id")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	creds.Secret, err = readSecretFile("/amazon-secret")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	creds.Token, err = readSecretFile("/amazon-token")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	// Get Cloudfront distribution (not required, though we can log a warning)
	distribution, err := readSecretFile("/amazon-distribution")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	// Get endpoint for custom deployment (optional).
	endpoint, err := readSecretFile("/custom-endpoint")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	return NewAmazonClient(region, bucket, &creds, distribution, endpoint)
}

// NewAmazonClientFromEnv creates a Amazon client based on environment variables.
func NewAmazonClientFromEnv() (Client, error) {
	region, ok := os.LookupEnv(AmazonRegionEnvVar)
	if !ok {
		return nil, errors.Errorf("%s not found", AmazonRegionEnvVar)
	}
	bucket, ok := os.LookupEnv(AmazonBucketEnvVar)
	if !ok {
		return nil, errors.Errorf("%s not found", AmazonBucketEnvVar)
	}

	var creds AmazonCreds
	creds.ID, _ = os.LookupEnv(AmazonIDEnvVar)
	creds.Secret, _ = os.LookupEnv(AmazonSecretEnvVar)
	creds.Token, _ = os.LookupEnv(AmazonTokenEnvVar)

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
		root := strings.ReplaceAll(url.Bucket, ".", "/")
		c, err = NewLocalClient("/" + root)
	}
	switch {
	case err != nil:
		return nil, err
	case c != nil:
		return TracingObjClient(url.Store, c), nil
	default:
		return nil, errors.Errorf("unrecognized object store: %s", url.Store)
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
		return nil, errors.Wrapf(err, "error parsing url %v", urlStr)
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
			return nil, errors.Errorf("malformed Azure URI: %v", urlStr)
		}
		return &ObjectStoreURL{
			Store:  url.Scheme,
			Bucket: parts[0],
			Object: strings.Trim(path.Join(parts[1:]...), "/"),
		}, nil
	}
	return nil, errors.Errorf("unrecognized object store: %s", url.Scheme)
}

// NewClientFromEnv creates a client based on environment variables.
func NewClientFromEnv(storageRoot string) (c Client, err error) {
	storageBackend, ok := os.LookupEnv(StorageBackendEnvVar)
	if !ok {
		return nil, errors.Errorf("storage backend environment variable not found")
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
		return nil, errors.Errorf("unrecognized storage backend: %s", storageBackend)
	}
}

// NewClientFromSecret creates a client based on mounted secret files.
func NewClientFromSecret(storageRoot string) (c Client, err error) {
	storageBackend, ok := os.LookupEnv(StorageBackendEnvVar)
	if !ok {
		return nil, errors.Errorf("storage backend environment variable not found")
	}
	return NewClient(storageBackend, storageRoot)
}

// NewClient creates an obj.Client using the given backend and storage root (for
// local backends).
// TODO: Not sure if we want to keep the storage root configuration for
// non-local deployments. If so, we will need to connect it to the object path
// prefix for chunks.
func NewClient(storageBackend string, storageRoot string) (Client, error) {
	var c Client
	var err error
	switch storageBackend {
	case Minio:
		// S3 compatible doesn't like leading slashes
		// TODO: Readd?
		//if len(dir) > 0 && dir[0] == '/' {
		//	dir = dir[1:]
		//}
		c, err = NewMinioClientFromSecret("")
	case Amazon:
		// amazon doesn't like leading slashes
		// TODO: Readd?
		//if len(dir) > 0 && dir[0] == '/' {
		//	dir = dir[1:]
		//}
		c, err = NewAmazonClientFromSecret("")
	case Google:
		// TODO figure out if google likes leading slashses
		c, err = NewGoogleClientFromSecret("")
	case Microsoft:
		c, err = NewMicrosoftClientFromSecret("")
	case Local:
		c, err = NewLocalClient(storageRoot)
	}
	switch {
	case err != nil:
		return nil, err
	case c != nil:
		return TracingObjClient(storageBackend, c), nil
	default:
		return nil, errors.Errorf("unrecognized storage backend: %s", storageBackend)
	}
}
