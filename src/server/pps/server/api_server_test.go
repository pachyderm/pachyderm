package server

import (
	"context"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	logrus "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func TestRenderTemplate(t *testing.T) {
	ctx := context.Background()
	client := newClient(t)
	res, err := client.RenderTemplate(ctx, &pps.RenderTemplateRequest{
		Args: map[string]string{
			"arg1": "value1",
		},
		Template: `
			function (arg1) {
				pipeline: {name: arg1},
			}
		`,
	})
	require.NoError(t, err)
	require.Len(t, res.Specs, 1)
}

func newClient(t testing.TB) pps.APIClient {
	srv := newServer(t)
	gc := grpcutil.NewTestClient(t, func(gs *grpc.Server) {
		pps.RegisterAPIServer(gs, srv)
	})
	return pps.NewAPIClient(gc)
}

func newServer(t testing.TB) pps.APIServer {
	txnEnv := transactionenv.New()
	db := dockertestenv.NewTestDB(t)
	etcdEnv := testetcd.NewEnv(t)
	env := Env{
		BackgroundContext: context.Background(),
		Logger:            logrus.StandardLogger(),

		DB:         db,
		EtcdClient: etcdEnv.EtcdClient,
		KubeClient: nil,

		AuthServer: nil,
		PFSServer:  nil,

		TxnEnv:        txnEnv,
		GetPachClient: nil,
		Config:        newConfig(t),
	}
	srv, err := NewAPIServerNoMaster(env)
	require.NoError(t, err)
	return srv
}

func newConfig(testing.TB) serviceenv.Configuration {
	return *serviceenv.ConfigFromOptions()
}
