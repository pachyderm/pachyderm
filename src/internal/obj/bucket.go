package obj

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/promutil"
	"go.uber.org/zap"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3v2 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"gocloud.dev/blob"
	"gocloud.dev/blob/azureblob"
	"gocloud.dev/blob/fileblob"
	"gocloud.dev/blob/gcsblob"
	"gocloud.dev/blob/s3blob"
	"gocloud.dev/gcp"
	"golang.org/x/oauth2/google"
)

// Bucket represents access to a single object storage bucket.
// Azure calls this a container, because it's not like that word conflicts with anything.
type Bucket = blob.Bucket

func NewMinioBucketFromSecret(ctx context.Context) (*Bucket, error) {
	fmt.Println("fahad: this code is running :)")
	var id, secret string
	disableSSL := true
	bucketName := os.Getenv("MINIO_BUCKET")
	if bucketName == "" {
		return nil, errors.New("minio bucket not found")
	}
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		return nil, errors.New("minio endpoint not found")
	}
	if os.Getenv("MINIO_ID") != "" {
		id = os.Getenv("MINIO_ID")
	}
	if os.Getenv("MINIO_SECRET") != "" {
		secret = os.Getenv("MINIO_SECRET")
	}
	if os.Getenv("MINIO_SECURE") != "" {
		ssl, err := strconv.ParseBool(os.Getenv("MINIO_SECURE"))
		if err != nil {
			return nil, errors.EnsureStack(err)
		}
		disableSSL = !ssl
	}
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String("dummy-region"),
		Credentials:      credentials.NewStaticCredentials(id, secret, ""),
		Endpoint:         aws.String(endpoint),
		DisableSSL:       aws.Bool(disableSSL),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	bucket, err := s3blob.OpenBucket(ctx, sess, bucketName, nil)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return bucket, nil
}

// NewLocalBucket creates a Bucket backed by the local filesystem.
// Data will be stored at dir.
func NewLocalBucket(dir string) (*Bucket, error) {
	return fileblob.OpenBucket(dir, nil)
}

func NewAmazonBucket(ctx context.Context, s3Opts s3v2.Options, bucketName string) (*Bucket, error) {
	s3c := s3v2.New(s3Opts)
	return s3blob.OpenBucketV2(ctx, s3c, bucketName, nil)
}

func NewGoogleBucket(ctx context.Context, creds *google.Credentials, bucketName string) (*Bucket, error) {
	tokSrc := gcp.CredentialsTokenSource(creds)
	gcpClient, err := gcp.NewHTTPClient(http.DefaultClient.Transport, tokSrc)
	if err != nil {
		return nil, err
	}
	return gcsblob.OpenBucket(ctx, gcpClient, bucketName, nil)
}

func NewMicrosoftBucket(ctx context.Context, svcURL string, cred azcore.TokenCredential, bucketName string) (*Bucket, error) {
	svcClient, err := azblob.NewServiceClient(svcURL, cred, nil)
	if err != nil {
		return nil, err
	}
	return azureblob.OpenBucket(ctx, svcClient, bucketName, nil)
}

func NewAmazonBucketFromEnv(ctx context.Context) (*Bucket, error) {
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

	// Get endpoint for custom deployment (optional).
	endpoint, _ := os.LookupEnv(CustomEndpointEnvVar)
	return NewAmazonBucket(ctx, s3.Options{
		Region:           region,
		EndpointResolver: s3.EndpointResolverFromURL(endpoint),
	}, bucket)
}

// This seems to be required in order to support disabling ssl verification -- which is needed for EDF testing.
func amazonSession(ctx context.Context) (*session.Session, error) {
	endpoint, err := readSecretFile(fmt.Sprintf("/%s", CustomEndpointEnvVar))
	if err != nil {
		return nil, err
	}
	advancedConfig := &AmazonAdvancedConfiguration{}
	if err := cmdutil.Populate(advancedConfig); err != nil {
		return nil, errors.EnsureStack(err)
	}
	timeout, err := time.ParseDuration(advancedConfig.Timeout)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	httpClient := &http.Client{Timeout: timeout}
	if advancedConfig.NoVerifySSL {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		httpClient.Transport = transport
	}
	region, err := readSecretFile(fmt.Sprintf("/%s", AmazonRegionEnvVar))
	if err != nil {
		return nil, errors.Errorf("amazon-region not found")
	}
	httpClient.Transport = promutil.InstrumentRoundTripper("s3", httpClient.Transport)
	awsConfig := &aws.Config{
		Region:     aws.String(region),
		MaxRetries: aws.Int(advancedConfig.Retries),
		HTTPClient: httpClient,
		DisableSSL: aws.Bool(advancedConfig.DisableSSL),
		Logger:     log.NewAmazonLogger(ctx),
	}
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
	if creds.ID != "" {
		awsConfig.Credentials = credentials.NewStaticCredentials(creds.ID, creds.Secret, creds.Token)
	}
	// Set custom endpoint for a custom deployment.
	if endpoint != "" {
		awsConfig.Endpoint = aws.String(endpoint)
		awsConfig.S3ForcePathStyle = aws.Bool(true)
	}
	// Create new session using awsConfig
	return session.NewSession(awsConfig)
}

// NewAmazonBucketFromSecret constructs an amazon client by reading credentials
// from a mounted AmazonSecret. You may pass "" for bucket in which case it
// will read the bucket from the secret.
func NewAmazonBucketFromSecret(ctx context.Context) (*Bucket, error) {
	// Use or retrieve S3 bucket
	sess, err := amazonSession(ctx)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	bucket, err := readSecretFile(fmt.Sprintf("/%s", AmazonBucketEnvVar))
	if err != nil {
		return nil, errors.Errorf("amazon bucket not found")
	}
	blobBucket, err := s3blob.OpenBucket(ctx, sess, bucket, nil)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return blobBucket, nil
}

func NewGoogleBucketFromEnv(ctx context.Context) (*Bucket, error) {
	bucket, ok := os.LookupEnv(GoogleBucketEnvVar)
	if !ok {
		return nil, errors.Errorf("%s not found", GoogleBucketEnvVar)
	}
	credData, ok := os.LookupEnv(GoogleCredEnvVar)
	if !ok {
		return nil, errors.Errorf("%s not found", GoogleCredEnvVar)
	}
	creds, err := google.CredentialsFromJSON(ctx, []byte(credData))
	if err != nil {
		return nil, err
	}
	return NewGoogleBucket(ctx, creds, bucket)
}

// NewMicrosoftBucketFromEnv creates a Microsoft client based on environment variables.
func NewMicrosoftBucketFromEnv(ctx context.Context) (*Bucket, error) {
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
	log.Debug(ctx, "", zap.String("id", id), zap.String("secret", secret))
	return NewMicrosoftBucket(ctx, "", nil, container)
}
