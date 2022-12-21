package pachd

import (
	"context"

	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	authserver "github.com/pachyderm/pachyderm/v2/src/server/auth/server"
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

func (pachwb *pachwBuilder) registerPFSServer(ctx context.Context) error {
	env, err := pfs_server.EnvFromServiceEnv(pachwb.env, pachwb.txnEnv)
	if err != nil {
		return err
	}
	apiServer, err := pfs_server.NewPachwAPIServer(*env)
	if err != nil {
		return err
	}
	pachwb.forGRPCServer(func(s *grpc.Server) { pfs.RegisterAPIServer(s, apiServer) })
	pachwb.env.SetPfsServer(apiServer)
	return nil
}

func (pachwb *pachwBuilder) registerAuthServer(ctx context.Context) error {
	apiServer, err := authserver.NewAuthServer(authserver.EnvFromServiceEnv(pachwb.env, pachwb.txnEnv), false, false, false)
	if err != nil {
		return err
	}
	pachwb.forGRPCServer(func(s *grpc.Server) {
		auth.RegisterAPIServer(s, apiServer)
	})
	pachwb.env.SetAuthServer(apiServer)
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
		pachwb.setupDB,
		pachwb.initInternalServer,
		pachwb.registerAuthServer,
		pachwb.registerPFSServer, //PFS seems to need a non-nil auth server.
		pachwb.registerTransactionServer,
		pachwb.registerHealthServer,
		pachwb.resumeHealth,

		pachwb.initTransaction,
		pachwb.internallyListen,
		pachwb.daemon.serve,
	)
}

// PachwMode runs a pachw-mode pachd.
// When in pachw mode, the pachd instance processes storage and url tasks.
func PachwMode(ctx context.Context, config any) error {
	return newPachwBuilder(config).buildAndRun(ctx)
}
