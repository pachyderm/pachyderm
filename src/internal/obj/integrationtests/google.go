package integrationtests

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"google.golang.org/api/option"
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
	client, err := obj.NewGoogleClient(bucket, opts)
	require.NoError(t, err)
	testInterruption(t, client)
}
