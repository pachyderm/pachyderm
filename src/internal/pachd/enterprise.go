package pachd

import (
	"context"
	"path"

	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	eprsserver "github.com/pachyderm/pachyderm/v2/src/server/enterprise/server"
	"google.golang.org/grpc"
)

// An EnterpriseBuilder builds an enterprise-mode pachd.  It should only be
// created with NewEnterpriseBuilder.
//
// Enterprise mode is the enterprise server which is used to manage multiple
// Pachyderm installations.
type EnterpriseBuilder struct {
	builder
}

func (eb *EnterpriseBuilder) registerEnterpriseServer(ctx context.Context) error {
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
	eb.licenseEnv.EnterpriseServer = apiServer
	return nil
}

// NewEnterpriseBuilder returns a new initialized EnterpriseBuilder.
func NewEnterpriseBuilder(config any) EnterpriseBuilder {
	return EnterpriseBuilder{newBuilder(config, "pachyderm-pachd-enterprise")}
}

// Build builds and starts an enterprise-mode pachd.
func (eb *EnterpriseBuilder) Build(ctx context.Context) error {
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
