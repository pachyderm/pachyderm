package obj

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"

	"google.golang.org/api/option"
)

func uniqueObjectName(prefix string) string {
	return prefix + uuid.NewWithoutDashes()[0:12]
}

func requireExists(t *testing.T, client Client, object string, expected bool) {
	exists := client.Exists(context.Background(), object)
	require.Equal(t, expected, exists)
}

func doWriteTest(t *testing.T, client Client, object string, writes []string) {
	requireExists(t, client, object, false)
	defer requireExists(t, client, object, false)

	w, err := client.Writer(context.Background(), object)
	require.NoError(t, err)
	for _, s := range writes {
		size, err := w.Write([]byte(s))
		require.NoError(t, err)
		require.Equal(t, len(s), size)
	}
	require.NoError(t, w.Close())

	defer func() {
		require.NoError(t, client.Delete(context.Background(), object))
	}()

	requireExists(t, client, object, true)

	r, err := client.Reader(context.Background(), object, 0, 0)
	require.NoError(t, err)

	expected := ""
	for _, s := range writes {
		expected += s
	}

	buf := make([]byte, len(expected)+1)
	size, err := r.Read(buf)
	require.Equal(t, len(expected), size)
	require.True(t, err == nil || errors.Is(err, io.EOF))

	size, err = r.Read(buf)
	require.Equal(t, 0, size)
	require.True(t, err == nil || errors.Is(err, io.EOF))

	// TODO: test offset reads

	require.NoError(t, r.Close())
}

func runTests(t *testing.T, client Client) {
	t.Run("TestMissingObject", func(t *testing.T) {
		t.Parallel()
		object := uniqueObjectName("test-missing-object-")
		requireExists(t, client, object, false)

		r, err := client.Reader(context.Background(), object, 0, 0)
		require.Nil(t, r)
		require.YesError(t, err)
		require.True(t, client.IsNotExist(err))
	})

	t.Run("TestEmptyWrite", func(t *testing.T) {
		t.Parallel()
		doWriteTest(t, client, uniqueObjectName("test-empty-write-"), []string{})
	})

	t.Run("TestSingleWrite", func(t *testing.T) {
		t.Parallel()
		doWriteTest(t, client, uniqueObjectName("test-single-write-"), []string{"foo"})
	})

	t.Run("TestMultiWrite", func(t *testing.T) {
		t.Parallel()
		doWriteTest(t, client, uniqueObjectName("test-multi-write-"), []string{"foo", "bar"})
	})

	t.Run("TestSubdirectory", func(t *testing.T) {
		t.Parallel()
		object := path.Join(uniqueObjectName("test-subdirectory-"), "object")
		doWriteTest(t, client, object, []string{"foo", "bar"})
	})

	t.Run("TestWalk", func(t *testing.T) {
		t.Parallel()
	})

	t.Run("TestInterruption", func(t *testing.T) {
		t.Parallel()

		// Make a canceled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		object := uniqueObjectName("test-interruption-")
		defer requireExists(t, client, object, false)

		w, err := client.Writer(ctx, object)
		require.NoError(t, err)
		// Some clients return an error now and some when the stream is closed
		if w != nil {
			err = w.Close()
		}
		require.YesError(t, err)
		require.True(t, errors.Is(err, context.Canceled))

		r, err := client.Reader(ctx, object, 0, 0)
		// Some clients return an error now and some when the stream is closed
		if r != nil {
			err = r.Close()
		}
		require.YesError(t, err)
		require.True(t, errors.Is(err, context.Canceled))

		err = client.Delete(ctx, object)
		require.YesError(t, err)
		require.True(t, errors.Is(err, context.Canceled))

		err = client.Walk(ctx, object, func(name string) error {
			require.False(t, true)
			return nil
		})
		require.YesError(t, err)
		require.True(t, errors.Is(err, context.Canceled))

		exists := client.Exists(ctx, object)
		require.False(t, exists)
	})
}

func newDefaultAmazonConfig() *AmazonAdvancedConfiguration {
	return &AmazonAdvancedConfiguration{
		Retries:        DefaultRetries,
		Timeout:        DefaultTimeout,
		UploadACL:      DefaultUploadACL,
		Reverse:        DefaultReverse,
		PartSize:       DefaultPartSize,
		MaxUploadParts: DefaultMaxUploadParts,
		DisableSSL:     DefaultDisableSSL,
		NoVerifySSL:    DefaultNoVerifySSL,
		LogOptions:     DefaultAwsLogOptions,
	}
}

func loadAmazonParameters(t *testing.T) (string, string, string, string) {
	id := os.Getenv("AMAZON_DEPLOYMENT_ID")
	secret := os.Getenv("AMAZON_DEPLOYMENT_SECRET")
	bucket := os.Getenv("AMAZON_DEPLOYMENT_BUCKET")
	region := os.Getenv("AMAZON_DEPLOYMENT_REGION")
	require.NotEqual(t, "", id)
	require.NotEqual(t, "", secret)
	require.NotEqual(t, "", bucket)
	require.NotEqual(t, "", region)

	return id, secret, bucket, region
}

func loadECSParameters(t *testing.T) (string, string, string, string, string) {
	id := os.Getenv("ECS_DEPLOYMENT_ID")
	secret := os.Getenv("ECS_DEPLOYMENT_SECRET")
	bucket := os.Getenv("ECS_DEPLOYMENT_BUCKET")
	region := os.Getenv("ECS_DEPLOYMENT_REGION") // The region is unused but needed by the object client
	endpoint := os.Getenv("ECS_DEPLOYMENT_CUSTOM_ENDPOINT")
	require.NotEqual(t, "", id)
	require.NotEqual(t, "", secret)
	require.NotEqual(t, "", bucket)
	require.NotEqual(t, "", region)
	require.NotEqual(t, "", endpoint)

	return id, secret, bucket, region, endpoint
}

func loadGoogleParameters(t *testing.T) (string, string) {
	bucket := os.Getenv("GOOGLE_DEPLOYMENT_BUCKET")
	creds := os.Getenv("GOOGLE_DEPLOYMENT_CREDS")
	require.NotEqual(t, "", bucket)
	require.NotEqual(t, "", creds)

	return bucket, creds
}

func loadGoogleHMACParameters(t *testing.T) (string, string, string) {
	id := os.Getenv("GOOGLE_DEPLOYMENT_HMAC_ID")
	secret := os.Getenv("GOOGLE_DEPLOYMENT_HMAC_SECRET")
	bucket := os.Getenv("GOOGLE_DEPLOYMENT_BUCKET")
	require.NotEqual(t, "", id)
	require.NotEqual(t, "", secret)
	require.NotEqual(t, "", bucket)

	return id, secret, bucket
}

func loadMicrosoftParameters(t *testing.T) (string, string, string) {
	id := os.Getenv("MICROSOFT_DEPLOYMENT_ID")
	secret := os.Getenv("MICROSOFT_DEPLOYMENT_SECRET")
	container := os.Getenv("MICROSOFT_DEPLOYMENT_CONTAINER")
	require.NotEqual(t, "", id)
	require.NotEqual(t, "", secret)
	require.NotEqual(t, "", container)

	return id, secret, container
}

func TestAmazonClient(t *testing.T) {
	advancedConfig := newDefaultAmazonConfig()

	// Test the Amazon client against S3
	t.Run("AmazonObjectStorage", func(t *testing.T) {
		t.Parallel()
		id, secret, bucket, region := loadAmazonParameters(t)
		creds := &AmazonCreds{ID: id, Secret: secret}

		client, err := newAmazonClient(region, bucket, creds, "", "", advancedConfig)
		require.NoError(t, err)
		runTests(t, client)
	})

	// Test the Amazon client against ECS
	t.Run("ECSObjectStorage", func(t *testing.T) {
		t.Parallel()
		id, secret, bucket, region, endpoint := loadECSParameters(t)
		creds := &AmazonCreds{ID: id, Secret: secret}

		client, err := newAmazonClient(region, bucket, creds, "", endpoint, advancedConfig)
		require.NoError(t, err)
		runTests(t, client)
	})
}

func TestMinioClient(t *testing.T) {
	minioTests := func(t *testing.T, endpoint string, bucket string, id string, secret string) {
		t.Run("S3v2", func(t *testing.T) {
			t.Parallel()
			client, err := newMinioClientV2(endpoint, bucket, id, secret, true)
			require.NoError(t, err)
			runTests(t, client)
		})

		t.Run("S3v4", func(t *testing.T) {
			t.Parallel()
			client, err := newMinioClient(endpoint, bucket, id, secret, true)
			require.NoError(t, err)
			runTests(t, client)
		})
	}

	// Test the Minio client against S3 using the S3v2 and S3v4 APIs
	t.Run("AmazonObjectStorage", func(t *testing.T) {
		id, secret, bucket, region := loadAmazonParameters(t)
		endpoint := fmt.Sprintf("s3.%s.amazonaws.com", region) // Note that not all AWS regions support both http/https or both S3v2/S3v4
		minioTests(t, endpoint, bucket, id, secret)
	})

	// Test the Minio client against ECS using the S3v2 and S3v4 APIs
	t.Run("ECSObjectStorage", func(t *testing.T) {
		id, secret, bucket, _, endpoint := loadECSParameters(t)
		minioTests(t, endpoint, bucket, id, secret)
	})

	// Test the Minio client against GCP using the S3v2 and S3v4 APIs
	t.Run("GoogleObjectStorage", func(t *testing.T) {
		id, secret, bucket := loadGoogleHMACParameters(t)
		minioTests(t, "storage.googleapis.com", bucket, id, secret)
	})
}

func TestGoogleClient(t *testing.T) {
	t.Parallel()
	bucket, credData := loadGoogleParameters(t)
	opts := []option.ClientOption{option.WithCredentialsJSON([]byte(credData))}
	client, err := newGoogleClient(bucket, opts)
	require.NoError(t, err)
	runTests(t, client)
}

func TestMicrosoftClient(t *testing.T) {
	t.Parallel()
	id, secret, container := loadMicrosoftParameters(t)
	client, err := newMicrosoftClient(container, id, secret)
	require.NoError(t, err)
	runTests(t, client)
}
