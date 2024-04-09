package pachd

import (
	"context"
	"fmt"
	"net"
	"os"
	"path"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	lokiclient "github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
	auth_interceptor "github.com/pachyderm/pachyderm/v2/src/internal/middleware/auth"
	clientlog_interceptor "github.com/pachyderm/pachyderm/v2/src/internal/middleware/logging/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/license"
	"github.com/pachyderm/pachyderm/v2/src/logs"
	"github.com/pachyderm/pachyderm/v2/src/metadata"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/proxy"
	admin_server "github.com/pachyderm/pachyderm/v2/src/server/admin/server"
	auth_iface "github.com/pachyderm/pachyderm/v2/src/server/auth"
	auth_server "github.com/pachyderm/pachyderm/v2/src/server/auth/server"
	debug_server "github.com/pachyderm/pachyderm/v2/src/server/debug/server"
	entiface "github.com/pachyderm/pachyderm/v2/src/server/enterprise"
	ent_server "github.com/pachyderm/pachyderm/v2/src/server/enterprise/server"
	license_server "github.com/pachyderm/pachyderm/v2/src/server/license/server"
	logs_server "github.com/pachyderm/pachyderm/v2/src/server/logs/server"
	metadata_server "github.com/pachyderm/pachyderm/v2/src/server/metadata/server"
	pfsiface "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	pfs_server "github.com/pachyderm/pachyderm/v2/src/server/pfs/server"
	ppsiface "github.com/pachyderm/pachyderm/v2/src/server/pps"
	pps_server "github.com/pachyderm/pachyderm/v2/src/server/pps/server"
	proxy_server "github.com/pachyderm/pachyderm/v2/src/server/proxy/server"
	txn_server "github.com/pachyderm/pachyderm/v2/src/server/transaction/server"
	"github.com/pachyderm/pachyderm/v2/src/transaction"
	version_server "github.com/pachyderm/pachyderm/v2/src/version"
	version "github.com/pachyderm/pachyderm/v2/src/version/versionpb"
)

// A fullBuilder builds a full-mode pachd.
type fullBuilder struct {
	builder
}

func (fb *fullBuilder) maybeRegisterIdentityServer(ctx context.Context) error {
	if fb.env.Config().EnterpriseMember {
		return nil
	}
	return fb.builder.registerIdentityServer(ctx)
}

// registerEnterpriseServer registers a FULL-mode enterprise server.  This
// differs from enterprise mode in that the mode & unpaused-mode options are
// passed; it differs from sidecar mode in that the mode & unpaused-mode options
// are passed, the heartbeat is enabled and the license environment’s enterprise
// server is set; it differs from paused mode in that the mode option is in full
// mode.
//
// TODO: refactor the four modes to have a cleaner license/enterprise server
// abstraction.
func (fb *fullBuilder) registerEnterpriseServer(ctx context.Context) error {
	fb.enterpriseEnv = EnterpriseEnv(
		fb.env,
		path.Join(fb.env.Config().EtcdPrefix, fb.env.Config().EnterpriseEtcdPrefix),
		fb.txnEnv,
	)
	apiServer, err := ent_server.NewEnterpriseServer(
		fb.enterpriseEnv,
		ent_server.Config{
			Heartbeat:    true,
			Mode:         ent_server.FullMode,
			UnpausedMode: os.Getenv("UNPAUSED_MODE"),
		},
	)
	if err != nil {
		return err
	}
	fb.forGRPCServer(func(s *grpc.Server) {
		enterprise.RegisterAPIServer(s, apiServer)
	})
	fb.bootstrappers = append(fb.bootstrappers, apiServer)
	fb.env.SetEnterpriseServer(apiServer)
	fb.licenseEnv.EnterpriseServer = apiServer
	return nil
}

// newFullBuilder returns a new initialized FullBuilder.
func newFullBuilder(config any) *fullBuilder {
	return &fullBuilder{newBuilder(config, "pachyderm-pachd-full")}
}

// buildAndRun builds and starts a full-mode pachd.
func (fb *fullBuilder) buildAndRun(ctx context.Context) error {
	return fb.apply(ctx,
		fb.tweakResources,
		fb.setupProfiling,
		fb.printVersion,
		fb.initJaeger,
		fb.initKube,
		fb.setupDB,
		fb.maybeInitDexDB,
		fb.maybeInitReporter,
		fb.initInternalServer,
		fb.initExternalServer,
		fb.registerLicenseServer,
		fb.registerEnterpriseServer,
		fb.maybeRegisterIdentityServer,
		fb.registerAuthServer,
		fb.registerPFSServer,
		fb.registerPPSServer,
		fb.registerTransactionServer,
		fb.registerAdminServer,
		fb.registerHealthServer,
		fb.registerVersionServer,
		fb.registerDebugServer,
		fb.registerProxyServer,
		fb.registerMetadataServer,
		fb.registerLogsServer,
		fb.initS3Server,
		fb.initPrometheusServer,
		fb.initPachwController,

		fb.initTransaction,
		fb.internallyListen,
		fb.initPachHTTPServer,
		fb.bootstrap,
		fb.externallyListen,
		fb.resumeHealth,
		fb.startPFSWorker,
		fb.startPFSMaster,
		fb.startPPSWorker,
		fb.startDebugWorker,
		fb.daemon.serve,
	)
}

// FullMode runs a full-mode pachd.
//
// Full mode is that standard pachd which users interact with using pachctl and
// which manages pipelines, files and so forth.
func FullMode(ctx context.Context, config *pachconfig.PachdFullConfiguration) error {
	return newFullBuilder(config).buildAndRun(ctx)
}

type Env struct {
	DB               *pachsql.DB
	DirectDB         *pachsql.DB
	DBListenerConfig string
	ObjClient        obj.Client
	Bucket           *obj.Bucket
	EtcdClient       *clientv3.Client
	Listener         net.Listener
	K8sObjects       []runtime.Object
	GetLokiClient    func() (*lokiclient.Client, error)
}

type Full struct {
	base
	env        Env
	config     pachconfig.PachdFullConfiguration
	dbListener collection.PostgresListener
	authReady  chan struct{}

	selfGRPC        *grpc.ClientConn
	authInterceptor *auth_interceptor.Interceptor
	txnEnv          *transactionenv.TransactionEnv

	healthSrv     grpc_health_v1.HealthServer
	version       version.APIServer
	txnSrv        transaction.APIServer
	authSrv       auth.APIServer
	pfsSrv        pfs.APIServer
	ppsSrv        pps.APIServer
	metadataSrv   metadata.APIServer
	adminSrv      admin.APIServer
	enterpriseSrv enterprise.APIServer
	licenseSrv    license.APIServer
	debugSrv      debug.DebugServer
	proxySrv      proxy.APIServer
	logsSrv       logs.APIServer

	pfsWorker   *pfs_server.Worker
	ppsWorker   *pps_server.Worker
	debugWorker *debug_server.Worker

	pfsMaster *pfs_server.Master
}

// NewFull sets up a new Full pachd and returns it.
func NewFull(env Env, config pachconfig.PachdFullConfiguration) *Full {
	pd := &Full{env: env, config: config, authReady: make(chan struct{})}

	pd.selfGRPC = newSelfGRPC(env.Listener, nil)
	pd.healthSrv = health.NewServer()
	pd.version = version_server.NewAPIServer(version_server.Version, version_server.APIServerOptions{})
	pd.txnEnv = transactionenv.New()
	pd.authInterceptor = auth_interceptor.NewInterceptor(func() auth_iface.APIServer {
		return pd.authSrv.(auth_iface.APIServer)
	})
	pd.debugWorker = debug_server.NewWorker(debug_server.WorkerEnv{
		PFS:         pfs.NewAPIClient(pd.selfGRPC),
		TaskService: task.NewEtcdService(env.EtcdClient, "debug"),
	})
	pd.ppsWorker = pps_server.NewWorker(pps_server.WorkerEnv{
		PFS:         pfs.NewAPIClient(pd.selfGRPC),
		TaskService: task.NewEtcdService(env.EtcdClient, config.PPSEtcdPrefix),
	})

	kubeClient := fake.NewSimpleClientset(env.K8sObjects...)

	pd.addSetup(
		printVersion(),
		tweakResources(config.GlobalConfiguration),
		initJaeger(),
		awaitDB(env.DB),
		runMigrations(env.DirectDB, env.EtcdClient),
		awaitMigrations(env.DB),
		setupStep{
			Name: "setup listener",
			Fn: func(ctx context.Context) error {
				pd.dbListener = collection.NewPostgresListener(env.DBListenerConfig)
				return nil
			},
		},

		// API Servers
		initTransactionServer(&pd.txnSrv, func() txn_server.Env {
			return txn_server.Env{
				DB:         env.DB,
				PGListener: pd.dbListener,
				TxnEnv:     pd.txnEnv,
			}
		}),
		initAuthServer(&pd.authSrv, func() auth_server.Env {
			return auth_server.Env{
				DB:         env.DB,
				EtcdClient: env.EtcdClient,
				Listener:   pd.dbListener,
				TxnEnv:     pd.txnEnv,
				Config: pachconfig.Configuration{
					GlobalConfiguration:             &config.GlobalConfiguration,
					PachdSpecificConfiguration:      &config.PachdSpecificConfiguration,
					EnterpriseSpecificConfiguration: &config.EnterpriseSpecificConfiguration,
				},
				GetEnterpriseServer: func() entiface.APIServer {
					return pd.enterpriseSrv.(entiface.APIServer)
				},
				BackgroundContext: pctx.Background("auth"),
				GetPfsServer: func() pfsiface.APIServer {
					return pd.pfsSrv.(pfsiface.APIServer)
				},
				GetPpsServer: func() ppsiface.APIServer {
					return pd.ppsSrv.(ppsiface.APIServer)
				},
			}
		}),
		initPFSAPIServer(&pd.pfsSrv, &pd.pfsMaster, func() pfs_server.Env {
			etcdPrefix := path.Join(config.EtcdPrefix, config.PFSEtcdPrefix)
			return pfs_server.Env{
				DB:            env.DB,
				Bucket:        env.Bucket,
				Listener:      pd.dbListener,
				ObjectClient:  env.ObjClient,
				EtcdClient:    env.EtcdClient,
				EtcdPrefix:    etcdPrefix,
				TaskService:   task.NewEtcdService(env.EtcdClient, etcdPrefix),
				TxnEnv:        pd.txnEnv,
				StorageConfig: config.StorageConfiguration,
				Auth:          pd.authSrv.(pfs_server.PFSAuth),
				GetPipelineInspector: func() pfs_server.PipelineInspector {
					return pd.ppsSrv.(pfs_server.PipelineInspector)
				},
				GetPPSServer: func() ppsiface.APIServer { return pd.ppsSrv.(pps_server.APIServer) },
			}
		}),
		initPPSAPIServer(&pd.ppsSrv, func() pps_server.Env {
			return pps_server.Env{
				AuthServer:        pd.authSrv.(auth_server.APIServer),
				BackgroundContext: pctx.Background("pps"),
				DB:                env.DB,
				Listener:          pd.dbListener,
				TxnEnv:            pd.txnEnv,
				KubeClient:        kubeClient,
				EtcdClient:        env.EtcdClient,
				EtcdPrefix:        path.Join(config.EtcdPrefix, config.PPSEtcdPrefix),
				TaskService:       task.NewEtcdService(env.EtcdClient, path.Join(config.EtcdPrefix, config.PPSEtcdPrefix)),
				GetLokiClient:     env.GetLokiClient,
				GetPachClient:     pd.mustGetPachClient,
				Config: pachconfig.Configuration{
					GlobalConfiguration:             &config.GlobalConfiguration,
					PachdSpecificConfiguration:      &config.PachdSpecificConfiguration,
					EnterpriseSpecificConfiguration: &config.EnterpriseSpecificConfiguration,
				},
				PFSServer: pd.pfsSrv.(pfs_server.APIServer),
			}
		}),
		initMetadataServer(&pd.metadataSrv, func() (env metadata_server.Env) {
			env.Auth = pd.authSrv.(auth_server.APIServer)
			env.TxnEnv = pd.txnEnv
			return
		}),
		setupStep{
			Name: "initTransactionEnv",
			Fn: func(ctx context.Context) error {
				pd.txnEnv.Initialize(env.DB,
					func() transactionenv.AuthBackend { return pd.authSrv.(auth_iface.APIServer) },
					func() transactionenv.PFSBackend { return pd.pfsSrv.(pfs_server.APIServer) },
					func() transactionenv.PPSBackend { return pd.ppsSrv.(pps_server.APIServer) },
					pd.txnSrv.(txn_server.APIServer),
				)
				return nil
			},
		},
		setupStep{
			Name: "initAdminServer",
			Fn: func(ctx context.Context) error {
				pd.adminSrv = admin_server.NewAPIServer(admin_server.Env{
					ClusterID: "mockPachd",
					Config: &pachconfig.Configuration{
						GlobalConfiguration:             &config.GlobalConfiguration,
						PachdSpecificConfiguration:      &config.PachdSpecificConfiguration,
						EnterpriseSpecificConfiguration: &config.EnterpriseSpecificConfiguration,
					},
					PFSServer: pd.pfsSrv,
					Paused:    false,
				})
				return nil
			},
		},
		setupStep{
			Name: "initEnterpriseServer",
			Fn: func(ctx context.Context) error {
				var err error
				pd.enterpriseSrv, err = ent_server.NewEnterpriseServer(
					&ent_server.Env{
						DB:         env.DB,
						Listener:   pd.dbListener,
						TxnEnv:     pd.txnEnv,
						EtcdClient: env.EtcdClient,
						EtcdPrefix: path.Join(config.EtcdPrefix, config.EnterpriseEtcdPrefix),
						AuthServer: pd.authSrv.(auth_server.APIServer),
						GetKubeClient: func() kubernetes.Interface {
							return kubeClient
						},
						GetPachClient:     pd.mustGetPachClient,
						Namespace:         "default",
						BackgroundContext: pctx.Background("enterprise"),
						Config: pachconfig.Configuration{
							GlobalConfiguration:             &config.GlobalConfiguration,
							PachdSpecificConfiguration:      &config.PachdSpecificConfiguration,
							EnterpriseSpecificConfiguration: &config.EnterpriseSpecificConfiguration,
						},
					},
					ent_server.Config{
						Heartbeat:    false,
						Mode:         ent_server.FullMode,
						UnpausedMode: "full",
					},
				)
				if err != nil {
					return errors.Wrap(err, "NewEnterpriseServer")
				}
				return nil
			},
		},
		setupStep{
			Name: "initLicenseServer",
			Fn: func(ctx context.Context) error {
				var err error
				pd.licenseSrv, err = license_server.New(&license_server.Env{
					DB:       env.DB,
					Listener: nil,
					Config: &pachconfig.Configuration{
						GlobalConfiguration:             &config.GlobalConfiguration,
						PachdSpecificConfiguration:      &config.PachdSpecificConfiguration,
						EnterpriseSpecificConfiguration: &config.EnterpriseSpecificConfiguration,
					},
					EnterpriseServer: pd.enterpriseSrv,
				})
				if err != nil {
					return errors.Wrap(err, "license_server.New")
				}
				return nil
			},
		},
		setupStep{
			Name: "initDebugServer",
			Fn: func(ctx context.Context) error {
				pd.debugSrv = debug_server.NewDebugServer(debug_server.Env{
					DB:            env.DB,
					Name:          "testpachd",
					GetLokiClient: env.GetLokiClient,
					GetKubeClient: func() kubernetes.Interface {
						return kubeClient
					},
					GetDynamicKubeClient: func() dynamic.Interface {
						return dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(), env.K8sObjects...)
					},
					Config: pachconfig.Configuration{
						GlobalConfiguration:             &config.GlobalConfiguration,
						PachdSpecificConfiguration:      &config.PachdSpecificConfiguration,
						EnterpriseSpecificConfiguration: &config.EnterpriseSpecificConfiguration,
					},
					TaskService:   task.NewEtcdService(env.EtcdClient, path.Join(config.EtcdPrefix, "debug")),
					GetPachClient: pd.mustGetPachClient,
				})
				return nil
			},
		},
		setupStep{
			Name: "initProxyServer",
			Fn: func(ctx context.Context) error {
				pd.proxySrv = proxy_server.NewAPIServer(proxy_server.Env{
					Listener: pd.dbListener,
				})
				return nil
			},
		},
		setupStep{
			Name: "initLogsServer",
			Fn: func(ctx context.Context) error {
				var err error
				pd.logsSrv, err = logs_server.NewAPIServer(logs_server.Env{
					GetLokiClient: env.GetLokiClient,
				})
				if err != nil {
					return errors.Wrap(err, "logs_server.NewAPIServer")
				}
				return nil
			},
		},

		// Workers
		initPFSWorker(&pd.pfsWorker, config.StorageConfiguration, func() pfs_server.WorkerEnv {
			etcdPrefix := path.Join(config.EtcdPrefix, config.PFSEtcdPrefix)
			return pfs_server.WorkerEnv{
				DB:          env.DB,
				ObjClient:   env.ObjClient,
				Bucket:      env.Bucket,
				TaskService: task.NewEtcdService(env.EtcdClient, etcdPrefix),
			}
		}),
	)
	pd.addBackground("pfsWorker", func(ctx context.Context) error {
		return pd.pfsWorker.Run(ctx)
	})
	pd.addBackground("pfsMaster", func(ctx context.Context) error {
		return pd.pfsMaster.Run(ctx)
	})
	pd.addBackground("ppsWorker", func(ctx context.Context) error {
		return pd.ppsWorker.Run(ctx)
	})
	pd.addBackground("debugWorker", func(ctx context.Context) error {
		return pd.debugWorker.Run(ctx)
	})
	pd.addBackground("grpc", newServeGRPC(pd.authInterceptor, env.Listener, func(gs grpc.ServiceRegistrar) {
		admin.RegisterAPIServer(gs, pd.adminSrv)
		auth.RegisterAPIServer(gs, pd.authSrv)
		debug.RegisterDebugServer(gs, pd.debugSrv)
		enterprise.RegisterAPIServer(gs, pd.enterpriseSrv)
		grpc_health_v1.RegisterHealthServer(gs, pd.healthSrv)
		license.RegisterAPIServer(gs, pd.licenseSrv)
		logs.RegisterAPIServer(gs, pd.logsSrv)
		metadata.RegisterAPIServer(gs, pd.metadataSrv)
		pfs.RegisterAPIServer(gs, pd.pfsSrv)
		pps.RegisterAPIServer(gs, pd.ppsSrv)
		proxy.RegisterAPIServer(gs, pd.proxySrv)
		version.RegisterAPIServer(gs, pd.version)
	}))
	pd.addBackground("bootstrap", func(ctx context.Context) error {
		defer close(pd.authReady)
		return errors.Join(
			errors.Wrap(bootstrapIfAble(ctx, pd.licenseSrv), "license"),
			errors.Wrap(bootstrapIfAble(ctx, pd.enterpriseSrv), "enterprise"),
			errors.Wrap(bootstrapIfAble(ctx, pd.authSrv), "auth"),
		)
	})
	return pd
}

func bootstrapIfAble(ctx context.Context, x any) error {
	if b, ok := x.(interface{ EnvBootstrap(context.Context) error }); ok {
		return b.EnvBootstrap(ctx)
	}
	return nil
}

// PachClient returns a pach client that can talk to the server with root privileges.
func (pd *Full) PachClient(ctx context.Context) (*client.APIClient, error) {
	addr, err := grpcutil.ParsePachdAddress("http://" + pd.env.Listener.Addr().String())
	if err != nil {
		return nil, errors.Wrap(err, "parse pachd address")
	}
	c, err := client.NewFromPachdAddress(ctx, addr,
		client.WithAdditionalUnaryClientInterceptors(
			clientlog_interceptor.LogUnary,
		),
		client.WithAdditionalStreamClientInterceptors(
			clientlog_interceptor.LogStream,
		),
	)
	if err != nil {
		return nil, errors.Wrap(err, "NewPachdFromAddress")
	}
	if t := pd.config.AuthRootToken; t != "" {
		c.SetAuthToken(t)
	}
	return c, nil
}

// mustGetPachClient returns an unauthenticated client for internal use by API servers.
func (pd *Full) mustGetPachClient(ctx context.Context) *client.APIClient {
	addr, err := grpcutil.ParsePachdAddress("http://" + pd.env.Listener.Addr().String())
	if err != nil {
		panic(fmt.Sprintf("parse pachd address: %v", err))
	}
	c, err := client.NewFromPachdAddress(ctx, addr,
		client.WithAdditionalUnaryClientInterceptors(
			clientlog_interceptor.LogUnary,
		),
		client.WithAdditionalStreamClientInterceptors(
			clientlog_interceptor.LogStream,
		),
	)
	if err != nil {
		panic(fmt.Sprintf("NewFromPachdAddress: %v", err))
	}
	return c
}

// AwaitAuth returns when auth is ready.  It must be called after Run() starts.
func (pd *Full) AwaitAuth(ctx context.Context) {
	select {
	case <-pd.authReady:
	case <-ctx.Done():
	}
}
