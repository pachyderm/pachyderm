package pachd

import (
	"context"
	"path"

	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	authserver "github.com/pachyderm/pachyderm/v2/src/server/auth/server"
	eprsserver "github.com/pachyderm/pachyderm/v2/src/server/enterprise/server"
	pfs_server "github.com/pachyderm/pachyderm/v2/src/server/pfs/server"
	pps_server "github.com/pachyderm/pachyderm/v2/src/server/pps/server"
)

// sidecarBuilder builds a sidecar-mode pachd instance.
type sidecarBuilder struct {
	builder
}

// newSidecarBuilder returns an initialized SidecarBuilder.
func newSidecarBuilder(config any) *sidecarBuilder {
	return &sidecarBuilder{newBuilder(config, "pachyderm-pachd-sidecar")}
}

func (sb *sidecarBuilder) registerAuthServer(ctx context.Context) error {
	apiServer, err := authserver.NewAuthServer(authserver.EnvFromServiceEnv(sb.env, sb.txnEnv), false, false, false)
	if err != nil {
		return err
	}
	sb.forGRPCServer(func(s *grpc.Server) {
		auth.RegisterAPIServer(s, apiServer)
	})
	sb.env.SetAuthServer(apiServer)
	return nil
}

func (sb *sidecarBuilder) registerPFSServer(ctx context.Context) error {
	env, err := pfs_server.EnvFromServiceEnv(sb.env, sb.txnEnv)
	if err != nil {
		return err
	}
	apiServer, err := pfs_server.NewSidecarAPIServer(*env)
	if err != nil {
		return err
	}
	sb.forGRPCServer(func(s *grpc.Server) { pfs.RegisterAPIServer(s, apiServer) })
	sb.env.SetPfsServer(apiServer)
	return nil
}

func (sb *sidecarBuilder) registerPPSServer(ctx context.Context) error {
	apiServer, err := pps_server.NewSidecarAPIServer(pps_server.EnvFromServiceEnv(sb.env, sb.txnEnv, nil),
		sb.env.Config().Namespace,
		sb.env.Config().PPSWorkerPort,
		sb.env.Config().PeerPort)
	if err != nil {
		return err
	}
	sb.forGRPCServer(func(s *grpc.Server) { pps.RegisterAPIServer(s, apiServer) })
	sb.env.SetPpsServer(apiServer)
	return nil
}

// registerEnterpriseServer registers a SIDECAR-mode enterprise server.  This
// differs from full & paused modes in that the mode & unpaused-mode options are
// not passed and there is no license server; and from enterprise mode in that
// heartbeat is false and there is no license server.
//
// TODO: refactor the four modes to have a cleaner license/enterprise server
// abstraction.
func (sb *sidecarBuilder) registerEnterpriseServer(ctx context.Context) error {
	sb.enterpriseEnv = eprsserver.EnvFromServiceEnv(
		sb.env,
		path.Join(sb.env.Config().EtcdPrefix, sb.env.Config().EnterpriseEtcdPrefix),
		sb.txnEnv)
	apiServer, err := eprsserver.NewEnterpriseServer(
		sb.enterpriseEnv,
		false,
	)
	if err != nil {
		return err
	}
	sb.forGRPCServer(func(s *grpc.Server) {
		enterprise.RegisterAPIServer(s, apiServer)
	})
	sb.env.SetEnterpriseServer(apiServer)
	return nil
}

// buildAndRun builds & starts a sidecar-mode pachd.
func (sb *sidecarBuilder) buildAndRun(ctx context.Context) error {
	return sb.apply(ctx,
		sb.printVersion,
		sb.tweakResources,
		sb.setupProfiling,
		sb.initJaeger,
		sb.initKube,
		sb.initInternalServer,
		sb.registerAuthServer,
		sb.registerPFSServer,
		sb.registerPPSServer,
		sb.registerEnterpriseServer,
		sb.registerTransactionServer,
		sb.registerHealthServer,
		sb.resumeHealth,
		sb.registerDebugServer,

		sb.initTransaction,
		sb.internallyListen,
		sb.daemon.serve,
	)
}

// SidecarMode runs a sidecar-mode pachd.
//
// Sidecar mode is run as a sidecar in a pipeline pod; it provides services to
// the pipeline worker code running in that pod.
func SidecarMode(ctx context.Context, config *pachconfig.PachdFullConfiguration) error {
	return newSidecarBuilder(config).buildAndRun(ctx)
}
