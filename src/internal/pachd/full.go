package pachd

import (
	"context"
	"os"
	"path"

	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	eprsserver "github.com/pachyderm/pachyderm/v2/src/server/enterprise/server"
)

// A FullBuilder builds a full-mode pachd.  It should always be created with
// NewFullBuilder.
//
// Full mode is that standard pachd which users interact with using pachctl and
// which manages pipelines, files and so forth.
type FullBuilder struct {
	builder
}

func (fb *FullBuilder) maybeRegisterIdentityServer(ctx context.Context) error {

	if fb.env.Config().EnterpriseMember {
		return nil
	}
	return fb.builder.registerIdentityServer(ctx)
}

func (fb *FullBuilder) registerEnterpriseServer(ctx context.Context) error {
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

// NewFullBuilder returns a new initialized FullBuilder.
func NewFullBuilder(config any) *FullBuilder {
	return &FullBuilder{newBuilder(config, "pachyderm-pachd-full")}
}

// BuildAndRun builds and starts a full-mode pachd.
func (fb *FullBuilder) BuildAndRun(ctx context.Context) error {
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
