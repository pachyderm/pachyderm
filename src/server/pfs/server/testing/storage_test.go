package testing

import (
	"context"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/pfs"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
)

// TestCheckStorage checks that the CheckStorage rpc is wired up correctly.
// An more extensive test lives in the `chunk` package.
func TestCheckStorage(t *testing.T) {
	ctx := context.Background()
	t.Parallel()
	client := newClient(t)
	res, err := client.CheckStorage(ctx, &pfs.CheckStorageRequest{
		ReadChunkData: false,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
}

func newClient(t testing.TB) pfs.APIClient {
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	return env.PachClient.PfsAPIClient
}
