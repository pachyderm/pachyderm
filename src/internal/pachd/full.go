package pachd

import (
	"context"
	"os"
	"path"

	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	eprsserver "github.com/pachyderm/pachyderm/v2/src/server/enterprise/server"
	"google.golang.org/grpc"
)

type fullBuilder struct {
	builder
}

func (fb *fullBuilder) maybeRegisterIdentityServer(ctx context.Context) error {

	if fb.env.Config().EnterpriseMember {
		return nil
	}
	return fb.builder.registerIdentityServer(ctx)
}

func (fb *fullBuilder) registerEnterpriseServer(ctx context.Context) error {
	fb.enterpriseEnv = eprsserver.EnvFromServiceEnv(
		fb.env,
		path.Join(fb.env.Config().EtcdPrefix, fb.env.Config().EnterpriseEtcdPrefix),
		fb.txnEnv,
		eprsserver.WithMode(eprsserver.FullMode),
		eprsserver.WithUnpausedMode(os.Getenv("UNPAUSED_MODE")),
	)
	apiServer, err := eprsserver.NewEnterpriseServer(
		fb.enterpriseEnv,
		true,
	)
	if err != nil {
		return err
	}
	fb.forGRPCServer(func(s *grpc.Server) {
		enterprise.RegisterAPIServer(s, apiServer)
	})
	fb.bootstrappers = append(fb.bootstrappers, apiServer)
	fb.env.SetEnterpriseServer(apiServer)
	fb.licenseEnv.EnterpriseServer = apiServer
	return nil
}

func NewFullBuilder(config any) Builder {
	return &fullBuilder{newBuilder(config, "pachyderm-pachd-full")}
}

func (fb *fullBuilder) Build(ctx context.Context) error {
	fb.daemon.criticalServersOnly = fb.env.Config().RequireCriticalServersOnly
	return fb.apply(ctx,
		fb.setupDB,
		fb.maybeInitDexDB,
		fb.maybeInitReporter,
		fb.initInternalServer,
		fb.initExternalServer,
		fb.registerLicenseServer,
		fb.registerEnterpriseServer,
		fb.maybeRegisterIdentityServer,
		fb.registerAuthServer,
		fb.registerPFSServer,
		fb.registerPPSServer,
		fb.registerTransactionServer,
		fb.registerAdminServer,
		fb.registerHealthServer,
		fb.registerVersionServer,
		fb.registerDebugServer,
		fb.registerProxyServer,
		fb.initS3Server,
		fb.initPrometheusServer,

		fb.initTransaction,
		fb.internallyListen,
		fb.bootstrap,
		fb.externallyListen,
		fb.resumeHealth,
		fb.daemon.serve,
	)
}
