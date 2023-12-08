package integrationtests

import (
	"testing"

	"google.golang.org/api/option"

	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

// NOTE: these tests require object storage credentials to be loaded in your
// environment (see util.go for where they are loaded).
func TestGoogleClient(t *testing.T) {
	t.Parallel()
	bucket, credData := LoadGoogleParameters(t)
	opts := []option.ClientOption{option.WithCredentialsJSON([]byte(credData))}
	obj.TestSuite(t, func(t testing.TB) obj.Client {
		client, err := obj.NewGoogleClient(bucket, opts)
		require.NoError(t, err)
		return client
	})

	t.Run("Interruption", func(t *testing.T) {
		client, err := obj.NewGoogleClient(bucket, opts)
		require.NoError(t, err)
		obj.TestInterruption(t, client)
	})
	t.Run("EmptyWrite", func(t *testing.T) {
		client, err := obj.NewGoogleClient(bucket, opts)
		require.NoError(t, err)
		obj.TestEmptyWrite(t, client)
	})
}

func TestGoogleBucket(t *testing.T) {
	t.Parallel()
	obj.TestBucket(t, func(t testing.TB) *obj.Bucket {
		ctx := pctx.TestContext(t)
		b, err := obj.NewGoogleBucketFromEnv(ctx)
		require.NoError(t, err)
		t.Cleanup(func() { b.Close() })
		return b
	})
}
