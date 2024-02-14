package pachd

import (
	"context"
	"math"
	"path"
	"runtime/debug"

	"github.com/dustin/go-humanize"
	"go.uber.org/zap"
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
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/metrics"
	authmw "github.com/pachyderm/pachyderm/v2/src/internal/middleware/auth"
	errorsmw "github.com/pachyderm/pachyderm/v2/src/internal/middleware/errors"
	loggingmw "github.com/pachyderm/pachyderm/v2/src/internal/middleware/logging"
	"github.com/pachyderm/pachyderm/v2/src/internal/middleware/validation"
	version_middleware "github.com/pachyderm/pachyderm/v2/src/internal/middleware/version"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	licenseclient "github.com/pachyderm/pachyderm/v2/src/license"
	"github.com/pachyderm/pachyderm/v2/src/logs"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/proxy"
	adminserver "github.com/pachyderm/pachyderm/v2/src/server/admin/server"
	authserver "github.com/pachyderm/pachyderm/v2/src/server/auth/server"
	debugserver "github.com/pachyderm/pachyderm/v2/src/server/debug/server"
	eprsserver "github.com/pachyderm/pachyderm/v2/src/server/enterprise/server"
	"github.com/pachyderm/pachyderm/v2/src/server/http"
	identity_server "github.com/pachyderm/pachyderm/v2/src/server/identity/server"
	licenseserver "github.com/pachyderm/pachyderm/v2/src/server/license/server"
	logsserver "github.com/pachyderm/pachyderm/v2/src/server/logs/server"
	pachw "github.com/pachyderm/pachyderm/v2/src/server/pachw/server"
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
	config *pachconfig.Configuration
	name   string

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

func newBuilder(config any, name string) (b builder) {
	b.name = name
	b.config = pachconfig.NewConfiguration(config)
	b.txnEnv = transactionenv.New()
	return
}

func (b *builder) printVersion(ctx context.Context) error {
	return printVersion().Fn(ctx)
}

func (b *builder) tweakResources(ctx context.Context) error {
	return tweakResources(*b.config.GlobalConfiguration).Fn(ctx)
}

func (b *builder) setupProfiling(ctx context.Context) error {
	return setupProfiling(b.name, b.config).Fn(ctx)
}

func (b *builder) initJaeger(ctx context.Context) error {
	return initJaeger().Fn(ctx)
}

func (b *builder) initKube(ctx context.Context) error {
	b.env = serviceenv.InitWithKube(ctx, b.config)
	if b.env.Config().EtcdPrefix == "" {
		b.env.Config().EtcdPrefix = collection.DefaultPrefix
	}
	b.authInterceptor = authmw.NewInterceptor(b.env.AuthServer)
	b.loggingInterceptor = loggingmw.NewLoggingInterceptor(ctx)
	if b.env.Config() != nil && b.env.Config().PachdSpecificConfiguration != nil {
		b.daemon.criticalServersOnly = b.env.Config().RequireCriticalServersOnly
	}
	return nil
}

func (b *builder) setupDB(ctx context.Context) error {
	if err := dbutil.WaitUntilReady(ctx, b.env.GetDBClient()); err != nil {
		return err
	}
	if err := migrations.ApplyMigrations(ctx, b.env.GetDBClient(), migrations.MakeEnv(nil, b.env.GetEtcdClient()), clusterstate.DesiredClusterState); err != nil {
		return err
	}
	if err := migrations.BlockUntil(ctx, b.env.GetDBClient(), clusterstate.DesiredClusterState); err != nil {
		return err
	}
	return nil
}

func (b *builder) waitForDBState(ctx context.Context) error {
	return awaitMigrations(b.env.GetDBClient()).Fn(ctx)
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
			validation.UnaryServerInterceptor,
		),
		grpc.ChainStreamInterceptor(
			errorsmw.StreamServerInterceptor,
			tracing.StreamServerInterceptor(),
			b.authInterceptor.InterceptStream,
			b.loggingInterceptor.StreamServerInterceptor,
			validation.StreamServerInterceptor,
		),
	)
	return err
}

func (b *builder) initExternalServer(ctx context.Context) error {
	var err error
	b.daemon.external, err = grpcutil.NewServer(
		ctx,
		true,
		grpc.ChainUnaryInterceptor(
			errorsmw.UnaryServerInterceptor,
			version_middleware.UnaryServerInterceptor,
			tracing.UnaryServerInterceptor(),
			b.authInterceptor.InterceptUnary,
			b.loggingInterceptor.UnaryServerInterceptor,
			validation.UnaryServerInterceptor,
		),
		grpc.ChainStreamInterceptor(
			errorsmw.StreamServerInterceptor,
			version_middleware.StreamServerInterceptor,
			tracing.StreamServerInterceptor(),
			b.authInterceptor.InterceptStream,
			b.loggingInterceptor.StreamServerInterceptor,
			validation.StreamServerInterceptor,
		),
	)
	return err
}

func (b builder) forGRPCServer(f func(*grpc.Server)) {
	b.daemon.forGRPCServer(f)
}

func (b *builder) registerLicenseServer(ctx context.Context) error {
	b.licenseEnv = LicenseEnv(b.env)
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
		AuthEnv(b.env, b.txnEnv),
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
	env, err := PFSEnv(b.env, b.txnEnv)
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
	apiServer, err := pps_server.NewAPIServer(PPSEnv(b.env, b.txnEnv, b.reporter))
	if err != nil {
		return err
	}
	b.forGRPCServer(func(s *grpc.Server) { pps.RegisterAPIServer(s, apiServer) })
	b.env.SetPpsServer(apiServer)
	return nil
}

func (b *builder) registerTransactionServer(ctx context.Context) error {
	var err error
	b.txn, err = transactionserver.NewAPIServer(transactionserver.Env{
		DB:         b.env.GetDBClient(),
		PGListener: b.env.GetPostgresListener(),
		TxnEnv:     b.txnEnv,
	})
	if err != nil {
		return err
	}
	b.forGRPCServer(func(s *grpc.Server) { transaction.RegisterAPIServer(s, b.txn) })
	return nil
}

func (b *builder) registerAdminServer(ctx context.Context) error {
	apiServer := adminserver.NewAPIServer(AdminEnv(b.env, false))
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
	apiServer := b.newDebugServer()
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

func (b *builder) registerLogsServer(ctx context.Context) error {
	apiServer, err := logsserver.NewAPIServer()
	if err != nil {
		return err
	}
	b.forGRPCServer(func(s *grpc.Server) { logs.RegisterAPIServer(s, apiServer) })
	return nil
}

func (b *builder) initTransaction(ctx context.Context) error {
	b.txnEnv.Initialize(
		b.env.GetDBClient(),
		func() transactionenv.AuthBackend { return b.env.AuthServer() },
		func() transactionenv.PFSBackend { return b.env.PfsServer() },
		func() transactionenv.PPSBackend { return b.env.PpsServer() },
		b.txn,
	)
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
		if err := b.EnvBootstrap(pctx.Child(ctx, "EnvBootstrap")); err != nil {
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

func (b *builder) initPachHTTPServer(ctx context.Context) error {
	var err error
	b.daemon.pachhttp, err = http.New(ctx, b.env.Config().DownloadPort, b.env.GetPachClient)
	if err != nil {
		return errors.Wrap(err, "new http server")
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

func (b *builder) initPachwController(ctx context.Context) error {
	env, err := PachwEnv(b.env)
	if err != nil {
		return err
	}
	pachw.NewController(ctx, env)
	return nil
}

func (b *builder) startPFSWorker(ctx context.Context) error {
	env, err := PFSWorkerEnv(b.env)
	if err != nil {
		return err
	}
	config := pfs_server.WorkerConfig{
		Storage: b.env.Config().StorageConfiguration,
	}
	w, err := pfs_server.NewWorker(*env, config)
	if err != nil {
		return err
	}
	go func() {
		ctx := pctx.Child(ctx, "pfs-worker")
		if err := w.Run(ctx); err != nil {
			log.Error(ctx, "from pfs-worker", zap.Error(err))
		}
	}()
	return nil
}

func (b *builder) startPFSMaster(ctx context.Context) error {
	env, err := PFSEnv(b.env, b.txnEnv)
	if err != nil {
		return err
	}
	m, err := pfs_server.NewMaster(*env)
	if err != nil {
		return err
	}
	go func() {
		ctx := pctx.Child(ctx, "pfs-master")
		if err := m.Run(ctx); err != nil {
			log.Error(ctx, "from pfs-master", zap.Error(err))
		}
	}()
	return nil
}

func (b *builder) startPPSWorker(ctx context.Context) error {
	etcdPrefix := path.Join(b.env.Config().EtcdPrefix, b.env.Config().PPSEtcdPrefix)
	w := pps_server.NewWorker(pps_server.WorkerEnv{
		PFS:         b.env.GetPachClient(ctx).PfsAPIClient,
		TaskService: b.env.GetTaskService(etcdPrefix),
	})
	go func() {
		ctx := pctx.Child(ctx, "pps-worker")
		if err := w.Run(ctx); err != nil {
			log.Error(ctx, "from pps-worker", zap.Error(err))
		}
	}()
	return nil
}

func (b *builder) startDebugWorker(ctx context.Context) error {
	w := debugserver.NewWorker(debugserver.WorkerEnv{
		PFS:         b.env.GetPachClient(ctx).PfsAPIClient,
		TaskService: b.env.GetTaskService(b.env.Config().EtcdPrefix),
	})
	go func() {
		if err := w.Run(ctx); err != nil {
			log.Error(ctx, "from debug worker", zap.Error(err))
		}
	}()
	return nil
}

// setupMemoryLimit sets GOMEMLIMIT.  If not already set through the environment, set GOMEMLIMIT to
// the container memory request, or if not set, the container memory limit minus some accounting for
// the runtime (100MiB).
func setupMemoryLimit(ctx context.Context, config pachconfig.GlobalConfiguration) {
	if memLimit := debug.SetMemoryLimit(-1); memLimit != math.MaxInt64 {
		log.Info(ctx, "memlimit: using configured GOMEMLIMIT", zap.String("limit", humanize.IBytes(uint64(memLimit))))
		return
	}

	// From https://go.dev/doc/gc-guide:
	// > Do take advantage of the memory limit when the execution environment of your Go program
	// > is entirely within your control, and the Go program is the only program with access to
	// > some set of resources (i.e. some kind of memory reservation, like a container memory
	// > limit).
	// >
	// > In this case, a good rule of thumb is to leave an additional 5-10% of headroom to
	// > account for memory sources the Go runtime is unaware of.
	//
	// We pick 5%, since CGO_ENABLED=0 which reduces "unknown" sources of memory.
	var target float64
	var source string
	if v := config.K8sMemoryRequest; v > 0 {
		target = v - 0.05*v
		source = "kubernetes request"
	} else if v := config.K8sMemoryLimit; v > 0 {
		target = v - 0.05*v
		source = "kubernetes limit"
	}
	if target <= 0 {
		log.Info(ctx, "memlimit: not setting GOMEMLIMIT; not configured explicitly, or as a kubernetes request, or as a kubernetes limit")
		return
	}

	actual := int64(math.Ceil(target))
	log.Info(ctx, "memlimit: setting GOMEMLIMIT (95% of the k8s value)", zap.String("limit", humanize.IBytes(uint64(actual))), zap.String("setFrom", source))
	debug.SetMemoryLimit(actual)
}

func (b *builder) newDebugServer() debugclient.DebugServer {
	return debugserver.NewDebugServer(DebugEnv(b.env))
}
