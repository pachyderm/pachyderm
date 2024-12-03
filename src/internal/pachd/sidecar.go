package pachd

import (
	"context"
	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	storageserver "github.com/pachyderm/pachyderm/v2/src/internal/storage"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	authserver "github.com/pachyderm/pachyderm/v2/src/server/auth/server"
	pfs_server "github.com/pachyderm/pachyderm/v2/src/server/pfs/server"
	pps_server "github.com/pachyderm/pachyderm/v2/src/server/pps/server"
	"github.com/pachyderm/pachyderm/v2/src/storage"
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
	apiServer, err := authserver.NewAuthServer(AuthEnv(sb.env, sb.txnEnv), false, false, false)
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
	env, err := PFSEnv(sb.env, sb.txnEnv)
	if err != nil {
		return err
	}
	apiServer, err := pfs_server.NewAPIServer(ctx, *env)
	if err != nil {
		return err
	}
	sb.forGRPCServer(func(s *grpc.Server) { pfs.RegisterAPIServer(s, apiServer) })
	sb.env.SetPfsServer(apiServer)
	return nil
}

func (sb *sidecarBuilder) registerStorageServer(ctx context.Context) error {
	env, err := StorageEnv(sb.env)
	if err != nil {
		return err
	}
	server, err := storageserver.New(ctx, *env)
	if err != nil {
		return err
	}
	sb.forGRPCServer(func(s *grpc.Server) { storage.RegisterFilesetServer(s, server) })
	return nil
}

func (sb *sidecarBuilder) registerPPSServer(ctx context.Context) error {
	apiServer, err := pps_server.NewSidecarAPIServer(PPSEnv(sb.env, sb.txnEnv, nil),
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

// buildAndRun builds & starts a sidecar-mode pachd.
func (sb *sidecarBuilder) buildAndRun(ctx context.Context) error {
	return sb.apply(ctx,
		sb.printVersion,
		sb.tweakResources,
		sb.setupProfiling,
		sb.initJaeger,
		sb.initKube,
		sb.restartOnSignal,
		sb.initInternalServer,
		sb.registerAuthServer,
		sb.registerPFSServer,
		sb.registerStorageServer,
		sb.registerPPSServer,
		sb.registerTransactionServer,
		sb.registerHealthServer,
		sb.resumeHealth,
		sb.registerDebugServer,
		sb.initPrometheusServer,

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
