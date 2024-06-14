package obj

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/promutil"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"gocloud.dev/blob"
	"gocloud.dev/blob/fileblob"
	"gocloud.dev/blob/s3blob"
)

// Bucket represents access to a single object storage bucket.
// Azure calls this a container, because it's not like that word conflicts with anything.
type Bucket = blob.Bucket

// Valid object storage backends
const (
	Minio     = "MINIO"
	Amazon    = "AMAZON"
	Google    = "GOOGLE"
	Microsoft = "MICROSOFT"
	Local     = "LOCAL"
)

// NewBucket creates a Bucket using the given backend and storage root (for
// local backends).
func NewBucket(ctx context.Context, storageBackend, storageRoot, storageURL string) (*Bucket, error) {
	// handle local differently since it is set without passing in an object storage url.
	if storageBackend == Local {
		if err := os.MkdirAll(storageRoot, 0777); err != nil {
			return nil, errors.Wrap(err, "new local bucket")
		}
		bucket, err := fileblob.OpenBucket(storageRoot, nil)
		if err != nil {
			return nil, errors.Wrap(err, "new local bucket")
		}
		return bucket, nil
	}
	objURL, err := ParseURL(storageURL)
	if err != nil {
		return nil, errors.Wrap(err, "new bucket")
	}
	var bucket *Bucket
	switch storageBackend {
	case Amazon:
		bucket, err = NewAmazonBucket(ctx, objURL)
	case Google, Microsoft:
		bucket, err = blob.OpenBucket(ctx, objURL.BucketString())
	default:
		return nil, errors.Errorf("unrecognized storage backend: %s", storageBackend)
	}
	if err != nil {
		return nil, errors.Wrap(err, "new bucket")
	}
	return bucket, nil
}

// NewAmazonBucket constructs an amazon client by reading credentials
// from a mounted AmazonSecret. You may pass "" for bucket in which case it
// will read the bucket from the secret.
func NewAmazonBucket(ctx context.Context, objURL *ObjectStoreURL) (*Bucket, error) {
	// Use or retrieve S3 bucket
	sess, err := amazonSession(ctx, objURL)
	if err != nil {
		return nil, errors.Wrap(err, "amazon bucket")
	}
	blobBucket, err := s3blob.OpenBucket(ctx, sess, objURL.Bucket, nil)
	if err != nil {
		return nil, errors.Wrap(err, "amazon bucket")
	}
	return blobBucket, nil
}

// This seems to be required in order to support disabling ssl verification -- which is needed for EDF testing.
func amazonSession(ctx context.Context, objURL *ObjectStoreURL) (*session.Session, error) {
	urlParams, err := url.ParseQuery(objURL.Params)
	if err != nil {
		return nil, errors.Wrap(err, "creating amazon session")
	}
	endpoint := urlParams.Get("endpoint")
	// if unset, disableSSL will be false.
	disableSSL, _ := strconv.ParseBool(urlParams.Get("disableSSL"))
	region := urlParams.Get("region")
	httpClient, retries, err := amazonHTTPClient()
	if err != nil {
		return nil, errors.Wrap(err, "creating amazon session")
	}
	awsConfig := &aws.Config{
		Region:     aws.String(region),
		MaxRetries: aws.Int(retries),
		HTTPClient: httpClient,
		DisableSSL: aws.Bool(disableSSL),
		Logger:     log.NewAmazonLogger(ctx),
	}
	// Set custom endpoint for a custom deployment.
	if endpoint != "" {
		awsConfig.Endpoint = aws.String(endpoint)
		awsConfig.S3ForcePathStyle = aws.Bool(true)
	}
	// Create new session using awsConfig
	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, errors.Wrap(err, "creating amazon session")
	}
	return sess, nil
}

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

func amazonHTTPClient() (*http.Client, int, error) {
	advancedConfig := &AmazonAdvancedConfiguration{}
	if err := cmdutil.Populate(advancedConfig); err != nil {
		return nil, -1, errors.Wrap(err, "creating amazon http client")
	}
	timeout, err := time.ParseDuration(advancedConfig.Timeout)
	if err != nil {
		return nil, -1, errors.Wrap(err, "creating amazon http client")
	}
	httpClient := &http.Client{Timeout: timeout}
	if advancedConfig.NoVerifySSL {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		httpClient.Transport = transport
	}
	httpClient.Transport = promutil.InstrumentRoundTripper("s3", httpClient.Transport)
	return httpClient, advancedConfig.Retries, nil
}
