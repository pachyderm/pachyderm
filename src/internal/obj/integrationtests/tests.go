package integrationtests

import (
	"bytes"
	"context"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

// BackendType is used to tell the tests which backend is being tested so some
// testing can be skipped for backends that do not support certain behavior.
type BackendType string

const (
	AmazonBackend    BackendType = "Amazon"
	ECSBackend       BackendType = "ECS"
	GoogleBackend    BackendType = "Google"
	MicrosoftBackend BackendType = "Microsoft"
	LocalBackend     BackendType = "Local"
)

// ClientType is used to tell the tests which client is being tested so some testing
// can be skipped for clients that do not support certain behavior.
type ClientType string

const (
	AmazonClient    ClientType = "Amazon"
	MinioClient     ClientType = "ECS"
	GoogleClient    ClientType = "Google"
	MicrosoftClient ClientType = "Microsoft"
	LocalClient     ClientType = "Local"
)

// testInterruption
// Interruption is currently not implemented on the Amazon, Microsoft, and Minio clients
//  Amazon client - use *WithContext methods
//  Microsoft client - move to github.com/Azure/azure-storage-blob-go which supports contexts
//  Minio client - upgrade to v7 which supports contexts in all APIs
//  Local client - interruptible file operations are not a thing in the stdlib
func testInterruption(t *testing.T, client obj.Client) {
	t.Parallel()

	// Make a canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	object := tu.UniqueString("test-interruption-")
	defer requireExists(t, client, object, false)

	// Some clients return an error immediately and some when the stream is closed
	err := client.Put(ctx, object, bytes.NewReader(nil))
	require.YesError(t, err)
	require.True(t, errors.Is(err, context.Canceled))

	// Some clients return an error immediately and some when the stream is closed
	w := &bytes.Buffer{}
	err = client.Get(ctx, object, w)
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

	exists, err := client.Exists(ctx, object)
	require.NoError(t, err)
	require.False(t, exists)
}

func requireExists(t *testing.T, client obj.Client, object string, expected bool) {
	exists, err := client.Exists(context.Background(), object)
	require.NoError(t, err)
	require.Equal(t, expected, exists)
}
