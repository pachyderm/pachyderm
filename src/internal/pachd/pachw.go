package pachd

import (
	"context"
	"path"

	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	authserver "github.com/pachyderm/pachyderm/v2/src/server/auth/server"
	eprsserver "github.com/pachyderm/pachyderm/v2/src/server/enterprise/server"
	pfs_server "github.com/pachyderm/pachyderm/v2/src/server/pfs/server"
)

// pachwBuilder builds a pachw-mode pachd instance.
type pachwBuilder struct {
	builder
}

// newPachwBuilder returns an initialized PachwBuilder.
func newPachwBuilder(config any) *pachwBuilder {
	return &pachwBuilder{newBuilder(config, "pachyderm-pachd-pachw")}
}

func (b *pachwBuilder) registerPFSServer(ctx context.Context) error {
	etcdPrefix := path.Join(b.env.Config().EtcdPrefix, b.env.Config().PFSEtcdPrefix)
	env := pfs_server.WorkerEnv{
		DB:         b.env.GetDBClient(),
		EtcdClient: b.env.GetEtcdClient(),
		EtcdPrefix: etcdPrefix,
	}
	config := pfs_server.WorkerConfig{
		Storage: b.env.Config().StorageConfiguration,
	}
	w, err := pfs_server.NewWorker(env, config)
	if err != nil {
		return err
	}
	go w.Run(ctx)
	return nil
}

func (pachwb *pachwBuilder) registerAuthServer(ctx context.Context) error {
	apiServer, err := authserver.NewAuthServer(
		AuthEnv(pachwb.env, pachwb.txnEnv),
		false, !pachwb.daemon.criticalServersOnly, false,
	)
	if err != nil {
		return err
	}
	pachwb.forGRPCServer(func(s *grpc.Server) {
		auth.RegisterAPIServer(s, apiServer)
	})
	pachwb.env.SetAuthServer(apiServer)
	pachwb.enterpriseEnv.AuthServer = apiServer
	return nil
}

func (pachwb *pachwBuilder) registerEnterpriseServer(ctx context.Context) error {
	pachwb.enterpriseEnv = EnterpriseEnv(
		pachwb.env,
		path.Join(pachwb.env.Config().EtcdPrefix, pachwb.env.Config().EnterpriseEtcdPrefix),
		pachwb.txnEnv,
	)
	apiServer, err := eprsserver.NewEnterpriseServer(
		pachwb.enterpriseEnv,
		eprsserver.Config{
			Heartbeat: false,
		},
	)
	if err != nil {
		return err
	}
	pachwb.forGRPCServer(func(s *grpc.Server) {
		enterprise.RegisterAPIServer(s, apiServer)
	})
	pachwb.env.SetEnterpriseServer(apiServer)
	return nil
}

func (pachwb *pachwBuilder) registerDebugServer(ctx context.Context) error {
	apiServer := pachwb.newDebugServer()
	pachwb.forGRPCServer(func(s *grpc.Server) { debug.RegisterDebugServer(s, apiServer) })
	return nil
}

// buildAndRun builds & starts a pachw-mode pachd.
func (pachwb *pachwBuilder) buildAndRun(ctx context.Context) error {
	return pachwb.apply(ctx,
		pachwb.tweakResources,
		pachwb.setupProfiling,
		pachwb.printVersion,
		pachwb.initJaeger,
		pachwb.initKube,
		pachwb.waitForDBState,
		pachwb.initInternalServer,
		pachwb.registerEnterpriseServer,
		pachwb.registerAuthServer,
		pachwb.registerPFSServer, // PFS needs a non-nil auth server.
		pachwb.registerTransactionServer,
		pachwb.registerDebugServer,
		pachwb.registerHealthServer,

		pachwb.initTransaction,
		pachwb.internallyListen,
		pachwb.resumeHealth,
		pachwb.daemon.serve,
	)
}

// PachwMode runs a pachw-mode pachd.
// When in pachw mode, the pachd instance processes storage and url tasks.
func PachwMode(ctx context.Context, config *pachconfig.PachdFullConfiguration) error {
	return newPachwBuilder(config).buildAndRun(ctx)
}
