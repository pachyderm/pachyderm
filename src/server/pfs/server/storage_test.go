package server

import (
	"context"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

// TestCheckStorage checks that the CheckStorage rpc is wired up correctly.
// An more extensive test lives in the `chunk` package.
func TestCheckStorage(t *testing.T) {
	ctx := pctx.TestContext(t)
	t.Parallel()
	client := newTestClient(ctx, t)

	res, err := client.CheckStorage(ctx, &pfs.CheckStorageRequest{
		ReadChunkData: false,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
}

func newTestClient(ctx context.Context, t testing.TB) pfs.APIClient {
	db := dockertestenv.NewTestDB(t)
	// We don't actually need this etcd, it's just so we can run the migrations.
	menv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, menv, clusterstate.DesiredClusterState))

	objC := dockertestenv.NewTestObjClient(ctx, t)
	return NewTestClient(t, db, objC)
}
