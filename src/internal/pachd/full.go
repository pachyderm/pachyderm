package pachd

import (
	"context"
	"os"
	"path"

	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	eprsserver "github.com/pachyderm/pachyderm/v2/src/server/enterprise/server"
)

// A fullBuilder builds a full-mode pachd.
type fullBuilder struct {
	builder
}

func (fb *fullBuilder) maybeRegisterIdentityServer(ctx context.Context) error {
	if fb.env.Config().EnterpriseMember {
		return nil
	}
	return fb.builder.registerIdentityServer(ctx)
}

// registerEnterpriseServer registers a FULL-mode enterprise server.  This
// differs from enterprise mode in that the mode & unpaused-mode options are
// passed; it differs from sidecar mode in that the mode & unpaused-mode options
// are passed, the heartbeat is enabled and the license environment’s enterprise
// server is set; it differs from paused mode in that the mode option is in full
// mode.
//
// TODO: refactor the four modes to have a cleaner license/enterprise server
// abstraction.
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

// newFullBuilder returns a new initialized FullBuilder.
func newFullBuilder(config any) *fullBuilder {
	return &fullBuilder{newBuilder(config, "pachyderm-pachd-full")}
}

// buildAndRun builds and starts a full-mode pachd.
func (fb *fullBuilder) buildAndRun(ctx context.Context) error {
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

// FullMode runs a full-mode pachd.
//
// Full mode is that standard pachd which users interact with using pachctl and
// which manages pipelines, files and so forth.
func FullMode(ctx context.Context, config any) error {
	return newFullBuilder(config).buildAndRun(ctx)
}
