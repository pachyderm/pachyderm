package pachd

import (
	"context"
	"os"
	"path"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	eprsserver "github.com/pachyderm/pachyderm/v2/src/server/enterprise/server"
	"google.golang.org/grpc"
)

// A pausedBuilder builds a paused-mode pachd.
type pausedBuilder struct {
	builder
}

// newPausedBuilder returns an initialized PausedBuilder.
func newPausedBuilder(config any) *pausedBuilder {
	return &pausedBuilder{newBuilder(config, "pachyderm-pachd-paused")}
}

// registerEnterpriseServer registers a PAUSED-mode enterprise server.  This
// differs from full mode in the mode option is set to paused; from enterprise
// mode in that the mode & unpaused-mode options are passed; and from sidecar
// mode in that the mode & unpaused-mode options are passed and the heartbeat is
// true.
//
// TODO: refactor the four modes to have a cleaner license/enterprise server
// abstraction.
func (pb *pausedBuilder) registerEnterpriseServer(ctx context.Context) error {
	pb.enterpriseEnv = eprsserver.EnvFromServiceEnv(
		pb.env,
		path.Join(pb.env.Config().EtcdPrefix, pb.env.Config().EnterpriseEtcdPrefix),
		pb.txnEnv,
		eprsserver.WithMode(eprsserver.PausedMode),
		eprsserver.WithUnpausedMode(os.Getenv("UNPAUSED_MODE")),
	)
	apiServer, err := eprsserver.NewEnterpriseServer(
		pb.enterpriseEnv,
		true,
	)
	if err != nil {
		return err
	}
	pb.forGRPCServer(func(s *grpc.Server) {
		enterprise.RegisterAPIServer(s, apiServer)
	})
	pb.bootstrappers = append(pb.bootstrappers, apiServer)
	pb.env.SetEnterpriseServer(apiServer)
	pb.licenseEnv.EnterpriseServer = apiServer

	// Stop workers because unpaused pachds in the process
	// of rolling may have started them back up.
	go func() error {
		backoff.Retry(func() error {
			ps, err := pb.enterpriseEnv.PauseStatus(ctx)
			if err != nil {
				return err
			}
			if ps.GetStatus() != enterprise.PauseStatusResponse_PAUSED {
				return errors.Errorf("pausing not finished, waiting to stop workers")
			}
			if err := pb.enterpriseEnv.StopWorkers(ctx); err != nil {
				return err
			}
			return nil
		}, backoff.RetryEvery(time.Second))

		return nil
	}()
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
		pb.setupDB,
		pb.maybeInitDexDB,
		pb.initInternalServer,
		pb.initExternalServer,
		pb.registerLicenseServer,
		pb.registerEnterpriseServer,
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
func PausedMode(ctx context.Context, config any) error {
	return newPausedBuilder(config).buildAndRun(ctx)
}
