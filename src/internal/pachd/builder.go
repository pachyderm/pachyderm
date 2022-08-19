package pachd

import (
	"context"
	"os"
	"runtime/debug"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/auth"
	debugclient "github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	logutil "github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/metrics"
	authmw "github.com/pachyderm/pachyderm/v2/src/internal/middleware/auth"
	errorsmw "github.com/pachyderm/pachyderm/v2/src/internal/middleware/errors"
	loggingmw "github.com/pachyderm/pachyderm/v2/src/internal/middleware/logging"
	version_middleware "github.com/pachyderm/pachyderm/v2/src/internal/middleware/version"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/profileutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	licenseclient "github.com/pachyderm/pachyderm/v2/src/license"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/proxy"
	adminserver "github.com/pachyderm/pachyderm/v2/src/server/admin/server"
	authserver "github.com/pachyderm/pachyderm/v2/src/server/auth/server"
	debugserver "github.com/pachyderm/pachyderm/v2/src/server/debug/server"
	eprsserver "github.com/pachyderm/pachyderm/v2/src/server/enterprise/server"
	identity_server "github.com/pachyderm/pachyderm/v2/src/server/identity/server"
	licenseserver "github.com/pachyderm/pachyderm/v2/src/server/license/server"
	pfs_server "github.com/pachyderm/pachyderm/v2/src/server/pfs/server"
	pps_server "github.com/pachyderm/pachyderm/v2/src/server/pps/server"
	proxyserver "github.com/pachyderm/pachyderm/v2/src/server/proxy/server"
	transactionserver "github.com/pachyderm/pachyderm/v2/src/server/transaction/server"
	"github.com/pachyderm/pachyderm/v2/src/transaction"
	"github.com/pachyderm/pachyderm/v2/src/version"
	"github.com/pachyderm/pachyderm/v2/src/version/versionpb"
)

// An envBootstrapper is a type which needs to have some bootstrap code run
// after initialization and before the dæmon starts listening.
//
// TODO: this could probably be formalized as part of splitting build & run for
// the daemon: build the daemon by building its services, then run the daemon by
// starting each service and finally starting the daemon itself.
type envBootstrapper interface {
	EnvBootstrap(context.Context) error
}

// builder provides the base daemon builder structure.
type builder struct {
	env                serviceenv.ServiceEnv
	daemon             daemon
	txnEnv             *transactionenv.TransactionEnv
	licenseEnv         *licenseserver.Env
	enterpriseEnv      *eprsserver.Env
	reporter           *metrics.Reporter
	authInterceptor    *authmw.Interceptor
	loggingInterceptor *loggingmw.LoggingInterceptor

	txn    transactionserver.APIServer
	health *health.Server

	bootstrappers []envBootstrapper
}

func (b *builder) apply(ctx context.Context, ff ...func(ctx context.Context) error) error {
	for _, f := range ff {
		if err := f(ctx); err != nil {
			return err
		}
	}
	return nil
}

func newBuilder(config interface{}, service string) (b builder) {
	switch logLevel := os.Getenv("LOG_LEVEL"); logLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "info", "":
		log.SetLevel(log.InfoLevel)
	default:
		log.Errorf("Unrecognized log level %s, falling back to default of \"info\"", logLevel)
		log.SetLevel(log.InfoLevel)
	}
	// must run InstallJaegerTracer before InitWithKube (otherwise InitWithKube
	// may create a pach client before tracing is active, not install the Jaeger
	// gRPC interceptor in the client, and not propagate traces)
	if endpoint := tracing.InstallJaegerTracerFromEnv(); endpoint != "" {
		log.Printf("connecting to Jaeger at %q", endpoint)
	} else {
		log.Printf("no Jaeger collector found (JAEGER_COLLECTOR_SERVICE_HOST not set)")
	}
	b.env = serviceenv.InitWithKube(serviceenv.NewConfiguration(config))
	if b.env.Config().LogFormat == "text" {
		log.SetFormatter(logutil.FormatterFunc(logutil.Pretty))
	}
	profileutil.StartCloudProfiler(service, b.env.Config())
	debug.SetGCPercent(b.env.Config().GCPercent)
	if b.env.Config().EtcdPrefix == "" {
		b.env.Config().EtcdPrefix = collection.DefaultPrefix
	}

	b.authInterceptor = authmw.NewInterceptor(b.env.AuthServer)
	b.loggingInterceptor = loggingmw.NewLoggingInterceptor(b.env.Logger())
	b.txnEnv = transactionenv.New()

	return
}

func (b *builder) setupDB(ctx context.Context) error {
	// TODO: currently all pachds attempt to apply migrations, we should coordinate this
	if err := dbutil.WaitUntilReady(context.Background(), log.StandardLogger(), b.env.GetDBClient()); err != nil {
		return err
	}
	if err := migrations.ApplyMigrations(context.Background(), b.env.GetDBClient(), migrations.MakeEnv(nil, b.env.GetEtcdClient()), clusterstate.DesiredClusterState); err != nil {
		return err
	}
	if err := migrations.BlockUntil(context.Background(), b.env.GetDBClient(), clusterstate.DesiredClusterState); err != nil {
		return err
	}
	return nil
}

func (b *builder) initDexDB(ctx context.Context) error {
	b.env.InitDexDB()
	return nil
}

func (b *builder) maybeInitReporter(ctx context.Context) error {
	if b.env.Config().Metrics {
		b.reporter = metrics.NewReporter(b.env)
	}
	return nil
}

func (b *builder) initInternalServer(ctx context.Context) error {
	var err error
	b.daemon.internal, err = grpcutil.NewServer(
		ctx,
		false,
		grpc.ChainUnaryInterceptor(
			errorsmw.UnaryServerInterceptor,
			tracing.UnaryServerInterceptor(),
			b.authInterceptor.InterceptUnary,
			b.loggingInterceptor.UnaryServerInterceptor,
		),
		grpc.ChainStreamInterceptor(
			errorsmw.StreamServerInterceptor,
			tracing.StreamServerInterceptor(),
			b.authInterceptor.InterceptStream,
			b.loggingInterceptor.StreamServerInterceptor,
		),
	)
	return err
}

func (b *builder) initExternalServer(ctx context.Context) error {
	var err error
	b.daemon.external, err = grpcutil.NewServer(
		ctx,
		true,
		// Add an UnknownServiceHandler to catch the case where the user has a client with the wrong major version.
		// Weirdly, GRPC seems to run the interceptor stack before the UnknownServiceHandler, so this is never called
		// (because the version_middleware interceptor throws an error, or the auth interceptor does).
		grpc.UnknownServiceHandler(func(srv interface{}, stream grpc.ServerStream) error {
			method, _ := grpc.MethodFromServerStream(stream)
			return errors.Errorf("unknown service %v", method)
		}),
		grpc.ChainUnaryInterceptor(
			errorsmw.UnaryServerInterceptor,
			version_middleware.UnaryServerInterceptor,
			tracing.UnaryServerInterceptor(),
			b.authInterceptor.InterceptUnary,
			b.loggingInterceptor.UnaryServerInterceptor,
		),
		grpc.ChainStreamInterceptor(
			errorsmw.StreamServerInterceptor,
			version_middleware.StreamServerInterceptor,
			tracing.StreamServerInterceptor(),
			b.authInterceptor.InterceptStream,
			b.loggingInterceptor.StreamServerInterceptor,
		),
	)
	return err
}

func (b builder) forGRPCServer(f func(*grpc.Server)) {
	b.daemon.forGRPCServer(f)
}

func (b *builder) registerLicenseServer(ctx context.Context) error {
	b.licenseEnv = licenseserver.EnvFromServiceEnv(b.env)
	apiServer, err := licenseserver.New(b.licenseEnv)
	if err != nil {
		return err
	}
	b.forGRPCServer(func(s *grpc.Server) {
		licenseclient.RegisterAPIServer(s, apiServer)
	})
	b.bootstrappers = append(b.bootstrappers, apiServer)
	return nil
}
func (b *builder) registerIdentityServer(ctx context.Context) error {
	apiServer := identity_server.NewIdentityServer(
		identity_server.EnvFromServiceEnv(b.env),
		true,
	)
	b.forGRPCServer(func(s *grpc.Server) { identity.RegisterAPIServer(s, apiServer) })
	b.env.SetIdentityServer(apiServer)
	b.bootstrappers = append(b.bootstrappers, apiServer)
	return nil
}

func (b *builder) registerAuthServer(ctx context.Context) error {
	apiServer, err := authserver.NewAuthServer(
		authserver.EnvFromServiceEnv(b.env, b.txnEnv),
		true, !b.daemon.criticalServersOnly, true,
	)
	if err != nil {
		return err
	}
	b.forGRPCServer(func(s *grpc.Server) {
		auth.RegisterAPIServer(s, apiServer)
	})
	b.env.SetAuthServer(apiServer)
	b.enterpriseEnv.AuthServer = apiServer
	b.bootstrappers = append(b.bootstrappers, apiServer)
	return nil
}

func (b *builder) registerPFSServer(ctx context.Context) error {
	env, err := pfs_server.EnvFromServiceEnv(b.env, b.txnEnv)
	if err != nil {
		return err
	}
	apiServer, err := pfs_server.NewAPIServer(*env)
	if err != nil {
		return err
	}
	b.forGRPCServer(func(s *grpc.Server) { pfs.RegisterAPIServer(s, apiServer) })
	b.env.SetPfsServer(apiServer)
	return nil
}

func (b *builder) registerPPSServer(ctx context.Context) error {
	apiServer, err := pps_server.NewAPIServer(pps_server.EnvFromServiceEnv(b.env, b.txnEnv, b.reporter))
	if err != nil {
		return err
	}
	b.forGRPCServer(func(s *grpc.Server) { pps.RegisterAPIServer(s, apiServer) })
	b.env.SetPpsServer(apiServer)
	return nil
}

func (b *builder) registerTransactionServer(ctx context.Context) error {
	var err error
	b.txn, err = transactionserver.NewAPIServer(b.env, b.txnEnv)
	if err != nil {
		return err
	}
	b.forGRPCServer(func(s *grpc.Server) { transaction.RegisterAPIServer(s, b.txn) })
	return nil
}

func (b *builder) registerAdminServer(ctx context.Context) error {
	apiServer := adminserver.NewAPIServer(adminserver.EnvFromServiceEnv(b.env))
	b.forGRPCServer(func(s *grpc.Server) { admin.RegisterAPIServer(s, apiServer) })
	return nil
}

func (b *builder) registerHealthServer(ctx context.Context) error {
	b.health = health.NewServer()
	b.health.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	b.forGRPCServer(func(s *grpc.Server) { grpc_health_v1.RegisterHealthServer(s, b.health) })
	return nil
}

func (b *builder) registerVersionServer(ctx context.Context) error {
	b.forGRPCServer(func(s *grpc.Server) {
		versionpb.RegisterAPIServer(s, version.NewAPIServer(version.Version, version.APIServerOptions{}))
	})
	return nil
}

func (b *builder) registerDebugServer(ctx context.Context) error {
	apiServer := debugserver.NewDebugServer(
		b.env,
		b.env.Config().PachdPodName,
		nil,
		b.env.GetDBClient(),
	)
	b.forGRPCServer(func(s *grpc.Server) { debugclient.RegisterDebugServer(s, apiServer) })
	return nil
}

func (b *builder) registerProxyServer(ctx context.Context) error {
	apiServer := proxyserver.NewAPIServer(proxyserver.Env{
		Listener: b.env.GetPostgresListener(),
	})
	b.forGRPCServer(func(s *grpc.Server) { proxy.RegisterAPIServer(s, apiServer) })
	return nil
}

func (b *builder) initTransaction(ctx context.Context) error {
	b.txnEnv.Initialize(b.env, b.txn)
	return nil
}

func (b *builder) internallyListen(ctx context.Context) error {
	if _, err := b.daemon.internal.ListenTCP("", b.env.Config().PeerPort); err != nil {
		return err
	}
	return nil
}

func (b *builder) externallyListen(ctx context.Context) error {
	if _, err := b.daemon.external.ListenTCP("", b.env.Config().Port); err != nil {
		return err
	}
	return nil
}

func (b *builder) bootstrap(ctx context.Context) error {
	for _, b := range b.bootstrappers {
		if err := b.EnvBootstrap(ctx); err != nil {
			return errors.EnsureStack(err)
		}
	}
	return nil
}

func (b *builder) resumeHealth(ctx context.Context) error {
	b.health.Resume()
	return nil
}

func (b *builder) initS3Server(ctx context.Context) error {
	b.daemon.s3 = &s3Server{
		clientFactory: b.env.GetPachClient,
		port:          b.env.Config().S3GatewayPort,
	}
	return nil
}

func (b *builder) initPrometheusServer(ctx context.Context) error {
	b.daemon.prometheus = &prometheusServer{port: b.env.Config().PrometheusPort}
	return nil
}

func (b *builder) maybeInitDexDB(ctx context.Context) error {
	if b.env.Config().EnterpriseMember {
		return nil
	}
	return b.initDexDB(ctx)
}
