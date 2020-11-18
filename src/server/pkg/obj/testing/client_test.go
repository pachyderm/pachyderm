package testing

import (
	"context"
	"fmt"
	"io"
	"path"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"

	"google.golang.org/api/option"
)

func requireExists(t *testing.T, client obj.Client, object string, expected bool) {
	exists := client.Exists(context.Background(), object)
	require.Equal(t, expected, exists)
}

func doWriteTest(t *testing.T, client obj.Client, object string, writes []string) {
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

	data := ""
	for _, s := range writes {
		data += s
	}

	doRead := func(offset uint64, length uint64) error {
		expected := ""
		if length == 0 || int(offset+length) > len(data) {
			expected = data[offset:]
		} else {
			expected = data[offset : offset+length]
		}

		r, err := client.Reader(context.Background(), object, offset, length)
		require.NoError(t, err)

		buf := make([]byte, len(data)+1)
		size, err := r.Read(buf)
		require.Equal(t, len(expected), size)
		if err != nil && !errors.Is(err, io.EOF) {
			require.NoError(t, r.Close())
			return err
		}

		size, err = r.Read(buf)
		require.Equal(t, 0, size)
		if err != nil && !errors.Is(err, io.EOF) {
			require.NoError(t, r.Close())
			return err
		}

		require.NoError(t, r.Close())
		return nil
	}

	doRead(0, 0)

	if len(data) > 0 {
		// Read the first character
		// TODO: this test is broken on the MicrosoftClient due to how the BlobRange is implemented
		// err := doRead(0, 1)
		// require.NoError(t, err)

		// Read the last character
		err = doRead(uint64(len(data)-1), 1)
		require.NoError(t, err)

		// Read through the end of the object
		err = doRead(uint64(len(data)-1), 10)
		require.YesError(t, err)
		require.Matches(t, "read stream ended after the wrong length", err.Error())

		// Read the middle of the object
		err = doRead(1, uint64(len(data)-2))
		require.NoError(t, err)

		// Read past the end of the object
		_, err = client.Reader(context.Background(), object, uint64(len(data)+1), 1)
		require.YesError(t, err) // The precise error here varies across clients and providers
	}
}

func runTests(t *testing.T, client obj.Client) {
	t.Run("TestMissingObject", func(t *testing.T) {
		t.Parallel()
		object := tu.UniqueString("test-missing-object-")
		requireExists(t, client, object, false)

		r, err := client.Reader(context.Background(), object, 0, 0)
		require.Nil(t, r)
		require.YesError(t, err)
		require.True(t, client.IsNotExist(err))
	})

	t.Run("TestEmptyWrite", func(t *testing.T) {
		t.Parallel()
		doWriteTest(t, client, tu.UniqueString("test-empty-write-"), []string{})
	})

	t.Run("TestSingleWrite", func(t *testing.T) {
		t.Parallel()
		doWriteTest(t, client, tu.UniqueString("test-single-write-"), []string{"foo"})
	})

	t.Run("TestMultiWrite", func(t *testing.T) {
		t.Parallel()
		doWriteTest(t, client, tu.UniqueString("test-multi-write-"), []string{"foo", "bar"})
	})

	t.Run("TestSubdirectory", func(t *testing.T) {
		t.Parallel()
		object := path.Join(tu.UniqueString("test-subdirectory-"), "object")
		doWriteTest(t, client, object, []string{"foo", "bar"})
	})

	// TODO: implement walk test

	t.Run("TestInterruption", func(t *testing.T) {
		t.Skip("Object client interruption is not currently supported across all clients")
		// Interruption is currently not implemented on the Amazon, Microsoft, and Minio clients
		//  Amazon client - use *WithContext methods
		//  Microsoft client - move to github.com/Azure/azure-storage-blob-go which supports contexts
		//  Minio client - upgrade to v7 which supports contexts in all APIs
		t.Parallel()

		// Make a canceled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		object := tu.UniqueString("test-interruption-")
		defer requireExists(t, client, object, false)

		// Some clients return an error immediately and some when the stream is closed
		w, err := client.Writer(ctx, object)
		require.NoError(t, err)
		if w != nil {
			err = w.Close()
		}
		require.YesError(t, err)
		require.True(t, errors.Is(err, context.Canceled))

		// Some clients return an error immediately and some when the stream is closed
		r, err := client.Reader(ctx, object, 0, 0)
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

func TestAmazonClient(t *testing.T) {
	t.Parallel()

	amazonTests := func(t *testing.T, id string, secret string, bucket string, region string, endpoint string) {
		for _, reverse := range []bool{true, false} {
			t.Run(fmt.Sprintf("reverse=%v", reverse), func(t *testing.T) {
				t.Parallel()
				creds := &obj.AmazonCreds{ID: id, Secret: secret}
				client, err := obj.NewAmazonClient(region, bucket, creds, "", endpoint, reverse)
				require.NoError(t, err)
				runTests(t, client)
			})
		}
	}

	// Test the Amazon client against S3
	t.Run("AmazonObjectStorage", func(t *testing.T) {
		t.Parallel()
		id, secret, bucket, region := LoadAmazonParameters(t)
		amazonTests(t, id, secret, bucket, region, "")
	})

	// Test the Amazon client against ECS
	t.Run("ECSObjectStorage", func(t *testing.T) {
		t.Parallel()
		id, secret, bucket, region, endpoint := LoadECSParameters(t)
		amazonTests(t, id, secret, bucket, region, endpoint)
	})

	// Test the Amazon client against GCS
	t.Run("GoogleObjectStorage", func(t *testing.T) {
		t.Skip("Amazon client gets 'InvalidArgument' errors when running against GCS")
		t.Parallel()
		id, secret, bucket, region, endpoint := LoadGoogleHMACParameters(t)
		amazonTests(t, id, secret, bucket, region, endpoint)
	})
}

func TestMinioClient(t *testing.T) {
	t.Parallel()
	minioTests := func(t *testing.T, endpoint string, bucket string, id string, secret string) {
		t.Run("S3v2", func(t *testing.T) {
			t.Skip("Minio client running S3v2 does not handle empty writes properly on S3 and ECS") // (this works for GCS), try upgrading to v7?
			t.Parallel()
			client, err := obj.NewMinioClient(endpoint, bucket, id, secret, true, true)
			require.NoError(t, err)
			runTests(t, client)
		})

		t.Run("S3v4", func(t *testing.T) {
			t.Parallel()
			client, err := obj.NewMinioClient(endpoint, bucket, id, secret, true, false)
			require.NoError(t, err)
			runTests(t, client)
		})
	}

	// Test the Minio client against S3 using the S3v2 and S3v4 APIs
	t.Run("AmazonObjectStorage", func(t *testing.T) {
		t.Parallel()
		id, secret, bucket, region := LoadAmazonParameters(t)
		endpoint := fmt.Sprintf("s3.%s.amazonaws.com", region) // Note that not all AWS regions support both http/https or both S3v2/S3v4
		minioTests(t, endpoint, bucket, id, secret)
	})

	// Test the Minio client against ECS using the S3v2 and S3v4 APIs
	t.Run("ECSObjectStorage", func(t *testing.T) {
		t.Parallel()
		id, secret, bucket, _, endpoint := LoadECSParameters(t)
		minioTests(t, endpoint, bucket, id, secret)
	})

	// Test the Minio client against GCS using the S3v2 and S3v4 APIs
	t.Run("GoogleObjectStorage", func(t *testing.T) {
		t.Parallel()
		id, secret, bucket, _, endpoint := LoadGoogleHMACParameters(t)
		minioTests(t, endpoint, bucket, id, secret)
	})
}

func TestGoogleClient(t *testing.T) {
	t.Parallel()
	bucket, credData := LoadGoogleParameters(t)
	opts := []option.ClientOption{option.WithCredentialsJSON([]byte(credData))}
	client, err := obj.NewGoogleClient(bucket, opts)
	require.NoError(t, err)
	runTests(t, client)
}

func TestMicrosoftClient(t *testing.T) {
	t.Parallel()
	id, secret, container := LoadMicrosoftParameters(t)
	client, err := obj.NewMicrosoftClient(container, id, secret)
	require.NoError(t, err)
	runTests(t, client)
}
