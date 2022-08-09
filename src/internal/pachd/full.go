package pachd

import (
	"context"
)

type fullBuilder struct {
	builder
}

func (fb *fullBuilder) maybeInitDexDB(ctx context.Context) error {
	if fb.env.Config().EnterpriseMember {
		return nil
	}
	return fb.builder.initDexDB(ctx)
}

func (fb *fullBuilder) registerIdentityServer(ctx context.Context) error {

	if fb.env.Config().EnterpriseMember {
		return nil
	}
	return fb.builder.registerIdentityServer(ctx)
}

func NewFullBuilder(config interface{}) Builder {
	return &fullBuilder{newBuilder(config, "pachyderm-pachd-full")}
}

func (fb *fullBuilder) Build(ctx context.Context) error {
	fb.daemon.requireNonCriticalServers = !fb.env.Config().RequireCriticalServersOnly
	return fb.apply(ctx,
		fb.setupDB,
		fb.maybeInitDexDB,
		fb.maybeInitReporter,
		fb.initInternalServer,
		fb.initExternalServer,
		fb.registerLicenseServer,
		fb.registerFullEnterpriseServer,
		fb.registerIdentityServer,
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
