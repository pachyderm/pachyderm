package obj

import (
	"context"
	"crypto/tls"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/promutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"gocloud.dev/blob"
	"gocloud.dev/blob/fileblob"
	"gocloud.dev/blob/s3blob"
)

// Bucket represents access to a single object storage bucket.
// Azure calls this a container, because it's not like that word conflicts with anything.
type Bucket = blob.Bucket

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

// NewBucket creates a Bucket using the given backend and storage root (for
// local backends).
// TODO: Not sure if we want to keep the storage root configuration for
// non-local deployments. If so, we will need to connect it to the object path
// prefix for chunks.
func NewBucket(ctx context.Context, storageBackend, storageRoot string) (*Bucket, error) {
	var err error
	var bucket *Bucket
	objURL, err := ParseURL(os.Getenv("STORAGE_URL"))
	if err != nil {
		return nil, errors.Wrap(err, "new bucket")
	}
	switch storageBackend {
	case Amazon:
		bucket, err = NewAmazonBucket(ctx, objURL)
	case Google, Microsoft:
		bucket, err = blob.OpenBucket(ctx, objURL.BucketString())
	case Local:
		bucket, err = fileblob.OpenBucket(storageRoot, nil)
	default:
		return nil, errors.Errorf("unrecognized storage backend: %s", storageBackend)
	}
	if err != nil {
		return nil, errors.Wrap(err, "new bucket")
	}
	return bucket, nil
}
