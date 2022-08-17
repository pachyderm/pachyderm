package pachd

import (
	"context"
	"path"

	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	eprsserver "github.com/pachyderm/pachyderm/v2/src/server/enterprise/server"
)

// An enterpriseBuilder builds an enterprise-mode pachd.
type enterpriseBuilder struct {
	builder
}

func (eb *enterpriseBuilder) registerEnterpriseServer(ctx context.Context) error {
	eb.enterpriseEnv = eprsserver.EnvFromServiceEnv(
		eb.env,
		path.Join(eb.env.Config().EtcdPrefix, eb.env.Config().EnterpriseEtcdPrefix),
		eb.txnEnv,
	)
	apiServer, err := eprsserver.NewEnterpriseServer(
		eb.enterpriseEnv,
		true,
	)
	if err != nil {
		return err
	}
	eb.forGRPCServer(func(s *grpc.Server) {
		enterprise.RegisterAPIServer(s, apiServer)
	})
	eb.bootstrappers = append(eb.bootstrappers, apiServer)
	eb.env.SetEnterpriseServer(apiServer)
	eb.daemon.license.EnterpriseServer = apiServer
	return nil
}

// newEnterpriseBuilder returns a new initialized EnterpriseBuilder.
func newEnterpriseBuilder(config any) *enterpriseBuilder {
	return &enterpriseBuilder{newBuilder(config, "pachyderm-pachd-enterprise")}
}

// buildAndRun builds and starts an enterprise-mode pachd.
func (eb *enterpriseBuilder) buildAndRun(ctx context.Context) error {
	return eb.apply(ctx,
		eb.setupDB,
		eb.maybeInitDexDB,
		eb.initInternalServer,
		eb.initExternalServer,
		eb.registerLicenseServer,
		eb.registerEnterpriseServer,
		eb.registerIdentityServer,
		eb.registerAuthServer,
		eb.registerHealthServer,
		eb.registerAdminServer,
		eb.registerVersionServer,

		eb.initTransaction,
		eb.internallyListen,
		eb.bootstrap,
		eb.externallyListen,
		eb.resumeHealth,
		eb.daemon.serve,
	)
}

// EnterpriseMode runs an enterprise-mode pachd.
//
// Enterprise mode is the enterprise server which is used to manage multiple
// Pachyderm installations.
func EnterpriseMode(ctx context.Context, config any) error {
	return newEnterpriseBuilder(config).buildAndRun(ctx)
}
