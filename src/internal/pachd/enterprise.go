package pachd

import (
	"context"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
)

// An enterpriseBuilder builds an enterprise-mode pachd.
type enterpriseBuilder struct {
	builder
}

// newEnterpriseBuilder returns a new initialized EnterpriseBuilder.
func newEnterpriseBuilder(config any) *enterpriseBuilder {
	return &enterpriseBuilder{newBuilder(config, "pachyderm-pachd-enterprise")}
}

// buildAndRun builds and starts an enterprise-mode pachd.
func (eb *enterpriseBuilder) buildAndRun(ctx context.Context) error {
	return eb.apply(ctx,
		eb.tweakResources,
		eb.setupProfiling,
		eb.printVersion,
		eb.initJaeger,
		eb.initKube,
		eb.setupDB,
		eb.restartOnSignal,
		eb.maybeInitDexDB,
		eb.initInternalServer,
		eb.initExternalServer,
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
func EnterpriseMode(ctx context.Context, config *pachconfig.EnterpriseServerConfiguration) error {
	return newEnterpriseBuilder(config).buildAndRun(ctx)
}
