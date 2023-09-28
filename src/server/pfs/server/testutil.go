package server

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"google.golang.org/grpc"
)

// NewTestClient creates a full PFS system which will be torn down at the end of the test.
// It returns a gRPC client for interacting with the API server.
// Background processes are intentionally not exposed.
func NewTestClient(t testing.TB, db *sqlx.DB, objC obj.Client) pfs.APIClient {
	x := newTestPFS(t, db, objC)
	gc := grpcutil.NewTestClient(t, func(gs *grpc.Server) {
		pfs.RegisterAPIServer(gs, x.API)
	})
	return pfs.NewAPIClient(gc)
}

type testPFS struct {
	API    pfs.APIServer
	Master *Master
	Worker *Worker
}

func newTestPFS(t testing.TB, db *sqlx.DB, objC obj.Client) testPFS {
	ctx, cf := context.WithCancel(pctx.TestContext(t))
	t.Cleanup(cf)
	sconf := pachconfig.StorageConfiguration{}

	ec := testetcd.NewEnv(ctx, t).EtcdClient
	ts := task.NewEtcdService(ec, "")
	env := Env{
		DB:           db,
		ObjectClient: objC,
		EtcdClient:   ec,
		EtcdPrefix:   "",
		TaskService:  ts,
		TxnEnv:       transactionenv.New(),

		// TODO
		Listener:             nil,
		GetPipelineInspector: nil,
		Auth:                 nil,
	}

	srv, err := NewAPIServer(env)
	require.NoError(t, err)

	w, err := NewWorker(WorkerEnv{
		DB:          db,
		ObjClient:   objC,
		TaskService: env.TaskService,
	}, WorkerConfig{Storage: sconf})
	require.NoError(t, err)

	m, err := NewMaster(env)
	require.NoError(t, err)

	go w.Run(ctx)
	// go m.Run(ctx)

	return testPFS{
		API:    srv,
		Worker: w,
		Master: m,
	}
}
