package server

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/testing/pfstest"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func TestPFS(t *testing.T) {
	pfstest.TestAPI(t, func(t testing.TB) pfs.APIClient {
		srv := newTestServer(t)
		gc := grpcutil.NewTestClient(t, func(gs *grpc.Server) {
			pfs.RegisterAPIServer(gs, srv)
		})
		return pfs.NewAPIClient(gc)
	})
}

// TODO: I wasn't able to get this working, but this is the goal.
func newTestServer(t testing.TB) pfs.APIServer {
	db := dockertestenv.NewTestDB(t)
	ctx := context.Background()
	etcdEnv := testetcd.NewEnv(t)
	err := dbutil.WithTx(ctx, db, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`CREATE SCHEMA storage`)
		require.NoError(t, err)
		_, err = tx.Exec(`CREATE SCHEMA pfs`)
		require.NoError(t, err)
		_, err = tx.Exec(`CREATE SCHEMA collections`)
		require.NoError(t, err)
		require.NoError(t, fileset.SetupPostgresStoreV0(tx))
		require.NoError(t, chunk.SetupPostgresStoreV0(tx))
		require.NoError(t, collection.SetupPostgresCollections(ctx, tx))
		require.NoError(t, collection.SetupPostgresV0(ctx, tx))
		return nil
	})
	require.NoError(t, err)
	oc, _ := obj.NewTestClient(t)
	srv, err := newAPIServer(Env{
		BackgroundContext: ctx,
		ObjectClient:      oc,
		DB:                db,
		Logger:            logrus.StandardLogger(),
		AuthServer:        nil,
		PPSServer:         nil,
		StorageConfig: serviceenv.StorageConfiguration{
			StorageCompactionMaxFanIn: 2,
		},
		EtcdClient: etcdEnv.EtcdClient,
		TxnEnv:     nil,
	})
	require.NoError(t, err)
	return srv
}
