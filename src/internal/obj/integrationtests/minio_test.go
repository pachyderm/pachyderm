package integrationtests

import (
	"fmt"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

// NOTE: these tests require object storage credentials to be loaded in your
// environment (see util.go for where they are loaded).
func TestMinioClient(t *testing.T) {
	t.Parallel()
	minioTests := func(t *testing.T, backend BackendType, endpoint string, bucket string, id string, secret string) {
		t.Run("S3v2", func(t *testing.T) {
			if backend == AmazonBackend || backend == ECSBackend {
				t.Skip("Minio client running S3v2 does not handle empty writes properly on S3 and ECS") // try upgrading to minio-go/v7?
			}
			t.Parallel()
			obj.TestSuite(t, func(t testing.TB) obj.Client {
				client, err := obj.NewMinioClient(endpoint, bucket, id, secret, true, true)
				require.NoError(t, err)
				return client
			})
		})

		t.Run("S3v4", func(t *testing.T) {
			t.Parallel()
			obj.TestSuite(t, func(t testing.TB) obj.Client {
				client, err := obj.NewMinioClient(endpoint, bucket, id, secret, true, true)
				require.NoError(t, err)
				return client
			})
		})
	}

	// Test the Minio client against S3 using the S3v2 and S3v4 APIs
	t.Run("AmazonObjectStorage", func(t *testing.T) {
		t.Parallel()
		id, secret, bucket, region := LoadAmazonParameters(t)
		endpoint := fmt.Sprintf("s3.%s.amazonaws.com", region) // Note that not all AWS regions support both http/https or both S3v2/S3v4
		minioTests(t, AmazonBackend, endpoint, bucket, id, secret)
	})

	// Test the Minio client against ECS using the S3v2 and S3v4 APIs
	t.Run("ECSObjectStorage", func(t *testing.T) {
		t.Parallel()
		id, secret, bucket, _, endpoint := LoadECSParameters(t)
		minioTests(t, ECSBackend, endpoint, bucket, id, secret)
	})

	// Test the Minio client against GCS using the S3v2 and S3v4 APIs
	t.Run("GoogleObjectStorage", func(t *testing.T) {
		t.Parallel()
		id, secret, bucket, _, endpoint := LoadGoogleHMACParameters(t)
		minioTests(t, GoogleBackend, endpoint, bucket, id, secret)
	})
}
