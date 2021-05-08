package integrationtests

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

// NOTE: these tests require object storage credentials to be loaded in your
// environment (see util.go for where they are loaded).
func TestAmazonClient(t *testing.T) {
	t.Parallel()

	amazonTests := func(t *testing.T, backendType BackendType, id string, secret string, bucket string, region string, endpoint string) {
		t.Parallel()

		obj.TestSuite(t, func(t testing.TB) obj.Client {
			creds := &obj.AmazonCreds{ID: id, Secret: secret}
			client, err := obj.NewAmazonClient(region, bucket, creds, "", endpoint)
			require.NoError(t, err)
			return client
		})
		creds := &obj.AmazonCreds{ID: id, Secret: secret}
		client, err := obj.NewAmazonClient(region, bucket, creds, "", endpoint)
		require.NoError(t, err)
		testInterruption(t, client)
	}

	// Test the Amazon client against S3
	t.Run("AmazonObjectStorage", func(t *testing.T) {
		t.Parallel()
		id, secret, bucket, region := LoadAmazonParameters(t)
		amazonTests(t, AmazonBackend, id, secret, bucket, region, "")
	})

	// Test the Amazon client against ECS
	t.Run("ECSObjectStorage", func(t *testing.T) {
		t.Parallel()
		id, secret, bucket, region, endpoint := LoadECSParameters(t)
		amazonTests(t, ECSBackend, id, secret, bucket, region, endpoint)
	})

	// Test the Amazon client against GCS
	t.Run("GoogleObjectStorage", func(t *testing.T) {
		t.Skip("Amazon client gets 'InvalidArgument' errors when running against GCS")
		t.Parallel()
		id, secret, bucket, region, endpoint := LoadGoogleHMACParameters(t)
		amazonTests(t, GoogleBackend, id, secret, bucket, region, endpoint)
	})
}
