package pachd

import (
	"context"
	"path"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	authserver "github.com/pachyderm/pachyderm/v2/src/server/auth/server"
	eprsserver "github.com/pachyderm/pachyderm/v2/src/server/enterprise/server"
	pps_server "github.com/pachyderm/pachyderm/v2/src/server/pps/server"
	"google.golang.org/grpc"
)

type sidecarBuilder struct {
	builder
}

func NewSidecarBuilder(config any) Builder {
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
	sb.bootstrappers = append(sb.bootstrappers, apiServer)
	sb.env.SetEnterpriseServer(apiServer)
	return nil
}

func (sb *sidecarBuilder) Build(ctx context.Context) error {
	return sb.apply(ctx,
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
