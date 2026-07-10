package obj

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/promutil"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/version"
	"go.uber.org/zap"
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

const (
	s3UserAgentProduct         = "pachyderm"
	s3UserAgentFallbackVersion = "dev"
)

var s3UserAgentFallbackLogOnce sync.Once

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
	rawVersion, uaVersion := userAgentVersion()
	if uaVersion != rawVersion {
		s3UserAgentFallbackLogOnce.Do(func() {
			log.Info(ctx, "using fallback S3 User-Agent version",
				zap.String("user_agent_product", s3UserAgentProduct),
				zap.String("user_agent_version", uaVersion),
				zap.String("raw_version", rawVersion))
		})
	}
	// Identify cluster-side S3 traffic by appending pachyderm/<version> to every
	// S3 request, including AWS and custom S3-compatible endpoints. This
	// intentionally discloses the exact Pachyderm version to the configured
	// object-store backend so provider-side access logs can attribute traffic.
	// PushBack appends to the SDK-built value rather than replacing it.
	sess.Handlers.Build.PushBack(
		request.MakeAddToUserAgentHandler(s3UserAgentProduct, uaVersion),
	)
	return sess, nil
}

// userAgentVersion returns the raw Pachyderm version resolved at build time and
// the version token to use in the S3 User-Agent. The token falls back to "dev"
// for unstamped builds or malformed/non-token version strings.
func userAgentVersion() (rawVersion, uaVersion string) {
	rawVersion = version.PrettyVersion()
	return rawVersion, userAgentVersionToken(rawVersion)
}

func userAgentVersionToken(v string) string {
	if v == version.UnstampedVersion || !isHTTPToken(v) {
		return s3UserAgentFallbackVersion
	}
	return v
}

// isHTTPToken reports whether v is a valid HTTP token per RFC 7230 Section
// 3.2.6 (tchar), safe to embed in a User-Agent product version without header
// injection or malformed-header risk.
func isHTTPToken(v string) bool {
	if v == "" {
		return false
	}
	for _, r := range v {
		switch {
		case r >= '0' && r <= '9':
		case r >= 'A' && r <= 'Z':
		case r >= 'a' && r <= 'z':
		case strings.ContainsRune("!#$%&'*+-.^_`|~", r):
		default:
			return false
		}
	}
	return true
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
