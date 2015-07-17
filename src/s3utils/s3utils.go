// TODO(pedge): the public vs non-public needs to go away, we need to put
// creds somewhere managed, and consistently call out to s3
package s3utils

import (
	"io"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	Private           = ACL("private")
	PublicRead        = ACL("public-read")
	PublicReadWrite   = ACL("public-read-write")
	AuthenticatedRead = ACL("authenticated-read")
	BucketOwnerRead   = ACL("bucket-owner-read")
	BucketOwnerFull   = ACL("bucket-owner-full-control")
)

type ACL string

// An s3 input looks like: s3://bucket/dir
// Where dir can be a path

// getBucket extracts the bucket from an s3 input
func GetBucket(input string) (string, error) {
	return strings.Split(strings.TrimPrefix(input, "s3://"), "/")[0], nil
}

// getPath extracts the path from an s3 input
func GetPath(input string) (string, error) {
	return path.Join(strings.Split(strings.TrimPrefix(input, "s3://"), "/")[1:]...), nil
}

func NewClient(public bool) *s3.S3 {
	if public {
		return s3.New(
			&aws.Config{
				Credentials: credentials.AnonymousCredentials,
			},
		)
	}
	return s3.New(nil)
}

// PutMulti is like a smart bucket.Put in that it will automatically do a
// multiput if the input reader has enough data that it makes sense to do so.
func PutMulti(bucketUri string, path string, r io.Reader, contType string, perm ACL) error {
	bucket, err := GetBucket(bucketUri)

	if err != nil {
		return err
	}

	uploader := s3manager.NewUploader(nil)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: &bucket,
		Key:    &path,
		Body:   r,
	})
	return err
}

// Files calls `cont` on each file found at `uri` starting at marker.
// Pass `marker=""` to start from the beginning.
// Returns the marker that should be passed to pick-up where this call left off.
func ForEachFile(uri string, public bool, marker string, cont func(file string, modtime time.Time) error) error {
	inPath, err := GetPath(uri)
	if err != nil {
		return err
	}

	bucket, err := GetBucket(uri)
	if err != nil {
		return err
	}

	client := NewClient(public)
	nextMarker := aws.String(marker)

	for {
		lr, err := client.ListObjects(&s3.ListObjectsInput{
			Bucket:  aws.String(bucket),
			Prefix:  aws.String(inPath),
			Marker:  nextMarker,
			MaxKeys: aws.Long(1000),
		})

		if err != nil {
			return err
		}

		for _, key := range lr.Contents {
			err = cont(*key.Key, *key.LastModified)

			if err != nil {
				return err
			}
		}

		if !*lr.IsTruncated {
			// We've exhausted the output
			break
		}

		nextMarker = lr.NextMarker
	}

	return nil
}
