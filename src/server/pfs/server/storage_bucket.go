package server

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"gocloud.dev/blob/s3blob"

	"gocloud.dev/blob"
	// Import the blob packages for the cloud backends we want to be able to open.
	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/memblob"
	_ "gocloud.dev/blob/s3blob"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
)

// openBucket swaps out client backend implementation for s3, gcp, and azure with go-cloud-sdk.
func openBucket(ctx context.Context, url *obj.ObjectStoreURL) (bucket *blob.Bucket, err error) {
	if os.Getenv("CUSTOM_ENDPOINT") != "" {
		if url.Params != "" {
			url.Params += "&"
		}
		url.Params += "endpoint=" + os.Getenv("CUSTOM_ENDPOINT")
	}
	if os.Getenv("DISABLE_SSL") != "" {
		if url.Params != "" {
			url.Params += "&"
		}
		url.Params += "disableSSL=" + os.Getenv("DISABLE_SSL")
	}
	switch url.Scheme {
	case "wasb":
		url.Scheme = "azblob"
	case "gcs": // assuming 'gcs' is an alias for 'gs'
		url.Scheme = "gs"
	}
	switch url.Scheme {
	case "s3", "gs", "azblob":
		bucket, err = blob.OpenBucket(ctx, url.BucketString())
		if err != nil {
			return nil, errors.EnsureStack(errors.Wrapf(err, "error opening bucket %s", url.Bucket))
		}
		return bucket, nil
	case "test-minio", "minio":
		return handleMinio(ctx, url)
	default:
		return nil, errors.Errorf("unrecognized object store: %s", url.Scheme)
	}
}

func handleMinio(ctx context.Context, url *obj.ObjectStoreURL) (bucket *blob.Bucket, err error) {
	endpoint := ""
	id := "minioadmin"
	secret := "minioadmin"
	disableSSL := true
	if os.Getenv("MINIO_ENDPOINT") != "" {
		endpoint = os.Getenv("MINIO_ENDPOINT")
	}
	parts := strings.SplitN(url.Bucket, "/", 2)
	if len(parts) < 2 && endpoint == "" {
		return nil, errors.Errorf("could not parse bucket %q from url", url.Bucket)
	}
	if endpoint == "" {
		endpoint = parts[0]
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
			return nil, errors.EnsureStack(errors.Wrap(err, "error parsing MINIO_SECURE to bool"))
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
		return nil, errors.EnsureStack(errors.Wrapf(err, "error creating session s", url.Bucket))
	}
	bucket, err = s3blob.OpenBucket(ctx, sess, parts[1], nil)
	if err != nil {
		return nil, errors.EnsureStack(errors.Wrapf(err, "error opening bucket %s", url.Bucket))
	}
	return bucket, nil
}
