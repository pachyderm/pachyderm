package obj

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/client"
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

// NewGoogleClient creates a google client with the given bucket name.
func NewGoogleClient(bucket string, opts []option.ClientOption) (c Client, err error) {
	if c, err = newGoogleClient(bucket, opts); err != nil {
		return nil, err
	}
	return newUniformClient(c), nil
}

func secretFile(name string) string {
	return filepath.Join("/", client.StorageSecretName, name)
}

func readSecretFile(name string) (string, error) {
	bytes, err := os.ReadFile(secretFile(name))
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
		bucket, err = readSecretFile(fmt.Sprintf("/%s", GoogleBucketEnvVar))
		if err != nil {
			return nil, errors.Errorf("google-bucket not found")
		}
	}
	cred, err := readSecretFile(fmt.Sprintf("/%s", GoogleCredEnvVar))
	if err != nil {
		return nil, errors.Errorf("google-cred not found")
	}
	var opts []option.ClientOption
	if cred != "" {
		opts = append(opts, option.WithCredentialsFile(secretFile(fmt.Sprintf("/%s", GoogleCredEnvVar))))
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
		container, err = readSecretFile(fmt.Sprintf("/%s", MicrosoftContainerEnvVar))
		if err != nil {
			return nil, errors.Errorf("microsoft-container not found")
		}
	}
	id, err := readSecretFile(fmt.Sprintf("/%s", MicrosoftIDEnvVar))
	if err != nil {
		return nil, errors.Errorf("microsoft-id not found")
	}
	secret, err := readSecretFile(fmt.Sprintf("/%s", MicrosoftSecretEnvVar))
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
		bucket, err = readSecretFile(fmt.Sprintf("/%s", MinioBucketEnvVar))
		if err != nil {
			return nil, err
		}
	}
	endpoint, err := readSecretFile(fmt.Sprintf("/%s", MinioEndpointEnvVar))
	if err != nil {
		return nil, err
	}
	id, err := readSecretFile(fmt.Sprintf("/%s", MinioIDEnvVar))
	if err != nil {
		return nil, err
	}
	secret, err := readSecretFile(fmt.Sprintf("/%s", MinioSecretEnvVar))
	if err != nil {
		return nil, err
	}
	secure, err := readSecretFile(fmt.Sprintf("/%s", MinioSecureEnvVar))
	if err != nil {
		return nil, err
	}
	secureBool, err := strconv.ParseBool(secure)
	if err != nil {
		return nil, errors.Errorf("Failed to convert minio secure boolean setting")
	}
	isS3V2, err := readSecretFile(fmt.Sprintf("/%s", MinioSignatureEnvVar))
	if err != nil {
		return nil, err
	}
	return NewMinioClient(endpoint, bucket, id, secret, secureBool, isS3V2 == "1")
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
	secureBool, err := strconv.ParseBool(secure)
	if err != nil {
		return nil, errors.Errorf("Failed to convert minio secure boolean setting")
	}
	isS3V2, ok := os.LookupEnv(MinioSignatureEnvVar)
	if !ok {
		return nil, errors.Errorf("%s not found", MinioSignatureEnvVar)
	}
	return NewMinioClient(endpoint, bucket, id, secret, secureBool, isS3V2 == "1")
}

// NewAmazonClientFromSecret constructs an amazon client by reading credentials
// from a mounted AmazonSecret. You may pass "" for bucket in which case it
// will read the bucket from the secret.
func NewAmazonClientFromSecret(bucket string, reverse ...bool) (Client, error) {
	// Get AWS region (required for constructing an AWS client)
	region, err := readSecretFile(fmt.Sprintf("/%s", AmazonRegionEnvVar))
	if err != nil {
		return nil, errors.Errorf("amazon-region not found")
	}

	// Use or retrieve S3 bucket
	if bucket == "" {
		bucket, err = readSecretFile(fmt.Sprintf("/%s", AmazonBucketEnvVar))
		if err != nil {
			return nil, err
		}
	}

	// Retrieve static credentials; if not found,
	// use IAM roles (i.e. the EC2 metadata service)
	var creds AmazonCreds
	creds.ID, err = readSecretFile(fmt.Sprintf("/%s", AmazonIDEnvVar))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	creds.Secret, err = readSecretFile(fmt.Sprintf("/%s", AmazonSecretEnvVar))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	creds.Token, err = readSecretFile(fmt.Sprintf("/%s", AmazonTokenEnvVar))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	// Get Cloudfront distribution (not required, though we can log a warning)
	distribution, err := readSecretFile(fmt.Sprintf("/%s", AmazonDistributionEnvVar))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	// Get endpoint for custom deployment (optional).
	endpoint, err := readSecretFile(fmt.Sprintf("/%s", CustomEndpointEnvVar))
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
	switch url.Scheme {
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
	case "test-minio":
		parts := strings.SplitN(url.Bucket, "/", 2)
		if len(parts) < 2 {
			return nil, errors.Errorf("could not parse bucket %q from url", url.Bucket)
		}
		c, err = NewMinioClient(parts[0], parts[1], "minioadmin", "minioadmin", false, false)
	case "local":
		root := strings.ReplaceAll(url.Bucket, ".", "/")
		c, err = NewLocalClient("/" + root)
	}
	switch {
	case err != nil:
		return nil, err
	case c != nil:
		return TracingObjClient(url.Scheme, c), nil
	default:
		return nil, errors.Errorf("unrecognized object store: %s", url.Scheme)
	}
}

// ObjectStoreURL represents a parsed URL to an object in an object store.
type ObjectStoreURL struct {
	// The object store, e.g. s3, gcs, as...
	Scheme string
	// The "bucket" (in AWS parlance) or the "container" (in Azure parlance).
	Bucket string
	// The object itself.
	Object string
}

func (s ObjectStoreURL) String() string {
	return fmt.Sprintf("%s://%s/%s", s.Scheme, s.Bucket, s.Object)
}

// ParseURL parses an URL into ObjectStoreURL.
func ParseURL(urlStr string) (*ObjectStoreURL, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, errors.Wrapf(err, "error parsing url %v", urlStr)
	}
	switch u.Scheme {
	case "s3", "gcs", "gs", "local":
		return &ObjectStoreURL{
			Scheme: u.Scheme,
			Bucket: u.Host,
			Object: strings.Trim(u.Path, "/"),
		}, nil
	case "as", "wasb":
		// In Azure, the first part of the path is the container name.
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		if len(parts) < 1 {
			// return nil, errors.Errorf("malformed Azure URI: %v", urlStr)
			return nil, errors.Errorf("malformed Azure URI: %v", urlStr)
		}
		return &ObjectStoreURL{
			Scheme: u.Scheme,
			Bucket: parts[0],
			Object: strings.Trim(path.Join(parts[1:]...), "/"),
		}, nil
	case "minio", "test-minio":
		parts := strings.SplitN(strings.Trim(u.Path, "/"), "/", 2)
		var key string
		if len(parts) == 2 {
			key = parts[1]
		}
		return &ObjectStoreURL{
			Scheme: u.Scheme,
			Bucket: u.Host + "/" + parts[0],
			Object: key,
		}, nil
	}
	// return nil, errors.Errorf("unrecognized object store: %s", u.Scheme)
	return nil, errors.Errorf("unrecognized object store: %s", u.Scheme)
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
