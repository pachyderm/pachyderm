package pachd

import (
	"context"
	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	adminserver "github.com/pachyderm/pachyderm/v2/src/server/admin/server"
)

// A pausedBuilder builds a paused-mode pachd.
type pausedBuilder struct {
	builder
}

// newPausedBuilder returns an initialized PausedBuilder.
func newPausedBuilder(config any) *pausedBuilder {
	return &pausedBuilder{newBuilder(config, "pachyderm-pachd-paused")}
}

func (pb *pausedBuilder) registerAdminServer(ctx context.Context) error {
	apiServer := adminserver.NewAPIServer(AdminEnv(pb.env, true))
	pb.forGRPCServer(func(s *grpc.Server) { admin.RegisterAPIServer(s, apiServer) })
	return nil
}

func (pb *pausedBuilder) maybeRegisterIdentityServer(ctx context.Context) error {
	if pb.env.Config().EnterpriseMember {
		return nil
	}
	return pb.builder.registerIdentityServer(ctx)
}

// buildAndRun builds and starts a paused-mode pachd.
func (pb *pausedBuilder) buildAndRun(ctx context.Context) error {
	return pb.apply(ctx,
		pb.printVersion,
		pb.tweakResources,
		pb.setupProfiling,
		pb.initJaeger,
		pb.initKube,
		pb.waitForDBState,
		pb.restartOnSignal,
		pb.maybeInitDexDB,
		pb.initInternalServer,
		pb.initExternalServer,
		pb.registerAdminServer,
		pb.maybeRegisterIdentityServer,
		pb.registerAuthServer,
		pb.registerHealthServer,
		pb.registerTransactionServer,
		pb.initPrometheusServer,

		pb.initTransaction,
		pb.internallyListen,
		pb.externallyListen,
		pb.resumeHealth,
		pb.daemon.serve,
	)
}

// PausedMode runs a paused-mode pachd.
//
// Paused mode is a restricted mode which runs Pachyderm read-only in order to
// take offline backups.
func PausedMode(ctx context.Context, config *pachconfig.PachdFullConfiguration) error {
	return newPausedBuilder(config).buildAndRun(ctx)
}
