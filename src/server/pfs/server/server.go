package server

import (
	"testing"

	"github.com/coreos/etcd/clientv3"
	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	authtesting "github.com/pachyderm/pachyderm/v2/src/server/auth/testing"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	ppsserver "github.com/pachyderm/pachyderm/v2/src/server/pps"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// NewAPIServer creates an APIServer.
func NewAPIServer(env Env) (pfsserver.APIServer, error) {
	a, err := newAPIServer(env)
	if err != nil {
		return nil, err
	}
	go a.driver.master(env.BackgroundContext)
	return newValidatedAPIServer(a, env.AuthServer), nil
}

func NewSidecarAPIServer(env Env) (pfsserver.APIServer, error) {
	a, err := newAPIServer(env)
	if err != nil {
		return nil, err
	}
	return newValidatedAPIServer(a, env.AuthServer), nil
}

// TODO: This isn't working yet, but this is the goal
func newTestServer(t testing.TB, db *sqlx.DB, etcdClient *clientv3.Client) pfs.APIServer {
	ctx := context.Background()
	err := dbutil.WithTx(ctx, db, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`CREATE SCHEMA storage`)
		require.NoError(t, err)
		_, err = tx.Exec(`CREATE SCHEMA pfs`)
		require.NoError(t, err)
		_, err = tx.Exec(`CREATE SCHEMA collections`)
		require.NoError(t, err)
		require.NoError(t, fileset.SetupPostgresStoreV0(ctx, tx))
		require.NoError(t, chunk.SetupPostgresStoreV0(tx))
		require.NoError(t, collection.SetupPostgresCollections(ctx, tx))
		require.NoError(t, collection.SetupPostgresV0(ctx, tx))
		return nil
	})
	require.NoError(t, err)
	oc := dockertestenv.NewTestObjClient(t)
	srv, err := newAPIServer(Env{
		BackgroundContext: ctx,
		ObjectClient:      oc,
		DB:                db,
		Logger:            logrus.StandardLogger(),
		AuthServer:        &authtesting.InactiveAPIServer{},
		GetPPSServer:      func() ppsserver.APIServer { return nil },
		StorageConfig: serviceenv.StorageConfiguration{
			StorageCompactionMaxFanIn: 2,
		},
		EtcdClient: etcdClient,
		EtcdPrefix: "pfs",
		TxnEnv:     nil,
	})
	require.NoError(t, err)
	return srv
}

// NewTestClient constructs a PFS API client suitable for use in tests by clients of PFS.
// PFS communicates through shared memory with other parts of the system.
// The shared memory components (Postgres and Etcd) are required as parameters.
// Additional resources, such as object storage, are not exposed to clients
// They are constructed internally, and will be freed when the test completes.
func NewTestClient(t testing.TB, db *sqlx.DB, etcdClient *clientv3.Client) pfs.APIClient {
	srv := newTestServer(t, db, etcdClient)
	gc := grpcutil.NewTestClient(t, func(gs *grpc.Server) {
		pfs.RegisterAPIServer(gs, srv)
	})
	return pfs.NewAPIClient(gc)
}
