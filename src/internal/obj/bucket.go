package obj

import (
	"context"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	s3v2 "github.com/aws/aws-sdk-go-v2/service/s3"
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
