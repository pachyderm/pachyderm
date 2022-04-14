package server

import (
	"context"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/proxy"
)

func TestPPSMaster(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	txnEnv := transactionenv.New()
	realEnv := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	listener := client.NewProxyPostgresListener(func() (proxy.APIClient, error) { return env.PachClient.ProxyClient, nil })

	infraDriver := newMockInfraDriver()
	master := newMaster(ctx,
		EnvFromServiceEnv(realEnv.ServiceEnv, txnEnv, nil, WithoutKubeClient),
		txnEnv,
		ppsdb.Pipelines(env.ServiceEnv.GetDBClient(), listener),
		realEnv.ServiceEnv.Config().EtcdPrefix,
		infraDriver)
	go master.run()
	pipelines := ppsdb.Pipelines(env.ServiceEnv.GetDBClient(), listener)
	require.NoError(t, txnEnv.WithWriteContext(ctx, func(txCtx *txncontext.TransactionContext) error {
		if err := pipelines.ReadWrite(txCtx.SqlTx).Create("my-pipeline", &pps.PipelineInfo{}); err != nil {
			return errors.EnsureStack(err)
		}
		return nil
	}))
}
