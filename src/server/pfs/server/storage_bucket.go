package server

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/hashicorp/go-multierror"
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
	case "as", "wasb":
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
	case "test-minio":
		parts := strings.SplitN(url.Bucket, "/", 2)
		if len(parts) < 2 {
			return nil, errors.Errorf("could not parse bucket %q from url", url.Bucket)
		}
		sess, err := session.NewSession(&aws.Config{
			Region:           aws.String("dummy-region"),
			Credentials:      credentials.NewStaticCredentials("minioadmin", "minioadmin", ""),
			Endpoint:         aws.String(parts[0]),
			DisableSSL:       aws.Bool(true),
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
	default:
		return nil, errors.Errorf("unrecognized object store: %s", url.Scheme)
	}
}

func importObj(ctx context.Context, w io.Writer, name, bucketName string, bucket *blob.Bucket) (retErr error) {
	r, err := bucket.NewReader(ctx, name, nil)
	if err != nil {
		return errors.Wrapf(err, "error creating reader from bucket %s", bucketName)
	}
	defer func() {
		if err := r.Close(); err != nil {
			retErr = multierror.Append(retErr, errors.Wrapf(err, "error closing reader for  bucket %s", bucketName))
		}
	}()
	_, err = io.Copy(w, r)
	if err != nil {
		return errors.Wrapf(err, "error copying from reader to writer for bucket %s", bucketName)
	}
	return nil
}

func exportObj(ctx context.Context, r io.Reader, name, bucketName string, bucket *blob.Bucket) (retErr error) {
	exists, err := bucket.Exists(ctx, name)
	if err != nil {
		return errors.Wrapf(err, "error checking if key %s exists in bucket %s", name, bucketName)
	}
	if exists {
		return errors.EnsureStack(fmt.Errorf("key %s already exists in bucket %s", name, bucketName))
	}
	w, err := bucket.NewWriter(ctx, name, nil)
	if err != nil {
		return errors.Wrapf(err, "error creating writer for bucket %s", bucketName)
	}
	defer func() {
		if err := w.Close(); err != nil {
			retErr = multierror.Append(retErr, errors.Wrapf(err, "error closing writer for bucket %s", bucketName))
		}
	}()
	_, err = w.ReadFrom(r)
	if err != nil {
		return errors.Wrapf(err, "error writing to bucket %s", bucketName)
	}
	return nil
}
