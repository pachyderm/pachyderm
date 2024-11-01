package pachd

import (
	"context"
	"net"
	"os"
	"path"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
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
	"github.com/pachyderm/pachyderm/v2/src/license"
	"github.com/pachyderm/pachyderm/v2/src/logs"
	"github.com/pachyderm/pachyderm/v2/src/metadata"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pjs"
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
	"github.com/pachyderm/pachyderm/v2/src/server/pfs/s3"
	pfs_server "github.com/pachyderm/pachyderm/v2/src/server/pfs/server"
	ppsiface "github.com/pachyderm/pachyderm/v2/src/server/pps"
	pps_server "github.com/pachyderm/pachyderm/v2/src/server/pps/server"
	proxy_server "github.com/pachyderm/pachyderm/v2/src/server/proxy/server"
	txn_server "github.com/pachyderm/pachyderm/v2/src/server/transaction/server"
	"github.com/pachyderm/pachyderm/v2/src/snapshot"
	"github.com/pachyderm/pachyderm/v2/src/storage"
	"github.com/pachyderm/pachyderm/v2/src/transaction"
	version_server "github.com/pachyderm/pachyderm/v2/src/version"
	version "github.com/pachyderm/pachyderm/v2/src/version/versionpb"

	"github.com/pachyderm/pachyderm/v2/src/internal/authdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	lokiclient "github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
	auth_interceptor "github.com/pachyderm/pachyderm/v2/src/internal/middleware/auth"
	clientlog_interceptor "github.com/pachyderm/pachyderm/v2/src/internal/middleware/logging/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	pjs_server "github.com/pachyderm/pachyderm/v2/src/internal/pjs"
	"github.com/pachyderm/pachyderm/v2/src/internal/restart"
	snapshot_server "github.com/pachyderm/pachyderm/v2/src/internal/snapshot"
	storage_server "github.com/pachyderm/pachyderm/v2/src/internal/storage"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
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
		fb.restartOnSignal,
		fb.maybeInitDexDB,
		fb.maybeInitReporter,
		fb.initInternalServer,
		fb.initExternalServer,
		fb.registerLicenseServer,
		fb.registerEnterpriseServer,
		fb.maybeRegisterIdentityServer,
		fb.registerAuthServer,
		fb.registerPFSServer,
		fb.registerPJSServer,
		fb.registerSnapshotServer,
		fb.registerStorageServer,
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
		fb.ensurePJSWorkerSecret,
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
	Bucket           *obj.Bucket
	EtcdClient       *clientv3.Client
	Listener         net.Listener
	K8sObjects       []runtime.Object
	GetLokiClient    func() (*lokiclient.Client, error)
}
type FullOption struct {
	DesiredState *migrations.State
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

	healthServer     grpc_health_v1.HealthServer
	version          version.APIServer
	txnServer        transaction.APIServer
	authServer       auth.APIServer
	pfsServer        pfs.APIServer
	pjsServer        pjs.APIServer
	snapshotServer   snapshot.APIServer
	storageServer    *storage_server.Server
	ppsServer        pps.APIServer
	metadataServer   metadata.APIServer
	adminServer      admin.APIServer
	enterpriseServer enterprise.APIServer
	licenseServer    license.APIServer
	debugServer      debug.DebugServer
	proxyServer      proxy.APIServer
	logsServer       logs.APIServer
	s3Server         *s3.S3Server

	pfsWorker   *pfs_server.Worker
	ppsWorker   *pps_server.Worker
	debugWorker *debug_server.Worker

	pfsMaster *pfs_server.Master

	pachClient      *client.APIClient
	pachClientReady chan struct{}

	kubeClient kubernetes.Interface
}

// NewFull sets up a new Full pachd and returns it.
func NewFull(env Env, config pachconfig.PachdFullConfiguration, opt *FullOption) *Full {
	pd := &Full{
		env:             env,
		config:          config,
		authReady:       make(chan struct{}),
		pachClientReady: make(chan struct{}),
	}

	pd.selfGRPC = newSelfGRPC(env.Listener, nil)
	pd.healthServer = health.NewServer()
	pd.version = version_server.NewAPIServer(version_server.Version, version_server.APIServerOptions{})
	pd.txnEnv = transactionenv.New()
	pd.authInterceptor = auth_interceptor.NewInterceptor(func() auth_iface.APIServer {
		return pd.authServer.(auth_iface.APIServer)
	})
	pd.debugWorker = debug_server.NewWorker(debug_server.WorkerEnv{
		PFS:         pfs.NewAPIClient(pd.selfGRPC),
		TaskService: task.NewEtcdService(env.EtcdClient, "debug"),
	})
	pd.ppsWorker = pps_server.NewWorker(pps_server.WorkerEnv{
		PFS:         pfs.NewAPIClient(pd.selfGRPC),
		TaskService: task.NewEtcdService(env.EtcdClient, config.PPSEtcdPrefix),
	})

	pd.kubeClient = fake.NewSimpleClientset(env.K8sObjects...)

	desiredStateOverride := &clusterstate.DesiredClusterState
	if opt != nil && opt.DesiredState != nil {
		desiredStateOverride = opt.DesiredState
	}

	pd.addSetup(
		printVersion(),
		tweakResources(config.GlobalConfiguration),
		initJaeger(),
		awaitDB(env.DB),
		runMigrations(env.DirectDB, env.EtcdClient, desiredStateOverride),
		awaitMigrations(env.DB, *desiredStateOverride),
		setupStep{
			Name: "setup listener",
			Fn: func(ctx context.Context) error {
				pd.dbListener = collection.NewPostgresListener(env.DBListenerConfig)
				return nil
			},
		},
		setupStep{
			Name: "setup restarter",
			Fn: func(ctx context.Context) error {
				r, err := restart.New(ctx, env.DB, pd.dbListener)
				if err != nil {
					return errors.Wrap(err, "restart.New")
				}
				go func() {
					if err := r.RestartWhenRequired(ctx); err != nil {
						log.Error(ctx, "restart notifier failed", zap.Error(err))
					}
				}()
				return nil
			},
		},

		// API Servers
		initTransactionServer(&pd.txnServer, func() txn_server.Env {
			return txn_server.Env{
				DB:         env.DB,
				PGListener: pd.dbListener,
				TxnEnv:     pd.txnEnv,
			}
		}),
		initAuthServer(&pd.authServer, func() auth_server.Env {
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
					return pd.enterpriseServer.(entiface.APIServer)
				},
				BackgroundContext: pctx.Background("auth"),
				GetPfsServer: func() pfsiface.APIServer {
					return pd.pfsServer.(pfsiface.APIServer)
				},
				GetPpsServer: func() ppsiface.APIServer {
					return pd.ppsServer.(ppsiface.APIServer)
				},
			}
		}),
		initPFSAPIServer(&pd.pfsServer, &pd.pfsMaster, func() pfs_server.Env {
			etcdPrefix := path.Join(config.EtcdPrefix, config.PFSEtcdPrefix)
			return pfs_server.Env{
				DB:            env.DB,
				Bucket:        env.Bucket,
				Listener:      pd.dbListener,
				EtcdClient:    env.EtcdClient,
				EtcdPrefix:    etcdPrefix,
				TaskService:   task.NewEtcdService(env.EtcdClient, etcdPrefix),
				TxnEnv:        pd.txnEnv,
				StorageConfig: config.StorageConfiguration,
				Auth:          pd.authServer.(pfs_server.PFSAuth),
				GetPipelineInspector: func() pfs_server.PipelineInspector {
					return pd.ppsServer.(pfs_server.PipelineInspector)
				},
				GetPPSServer: func() ppsiface.APIServer { return pd.ppsServer.(pps_server.APIServer) },
			}
		}),
		initStorageServer(&pd.storageServer, func() storage_server.Env {
			return storage_server.Env{
				DB:     env.DB,
				Bucket: env.Bucket,
				Config: config.StorageConfiguration,
			}
		}),
		initPJSAPIServer(&pd.pjsServer, func() pjs_server.Env {
			return pjs_server.Env{
				DB:               env.DB,
				GetPermissionser: pd.authServer.(pjs_server.GetPermissionser),
				Storage:          pd.storageServer,
				GetAuthToken:     auth.GetAuthToken,
			}
		}),
		initPPSAPIServer(&pd.ppsServer, func() pps_server.Env {
			return pps_server.Env{
				AuthServer:        pd.authServer.(auth_server.APIServer),
				BackgroundContext: pctx.Background("pps"),
				DB:                env.DB,
				Listener:          pd.dbListener,
				TxnEnv:            pd.txnEnv,
				KubeClient:        pd.kubeClient,
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
				PFSServer: pd.pfsServer.(pfs_server.APIServer),
			}
		}),
		initMetadataServer(&pd.metadataServer, func() (env metadata_server.Env) {
			env.Auth = pd.authServer.(auth_server.APIServer)
			env.TxnEnv = pd.txnEnv
			return
		}),
		setupStep{
			Name: "initTransactionEnv",
			Fn: func(ctx context.Context) error {
				pd.txnEnv.Initialize(env.DB,
					func() transactionenv.AuthBackend { return pd.authServer.(auth_iface.APIServer) },
					func() transactionenv.PFSBackend { return pd.pfsServer.(pfs_server.APIServer) },
					func() transactionenv.PPSBackend { return pd.ppsServer.(pps_server.APIServer) },
					pd.txnServer.(txn_server.APIServer),
				)
				return nil
			},
		},
		setupStep{
			Name: "initAdminServer",
			Fn: func(ctx context.Context) error {
				pd.adminServer = admin_server.NewAPIServer(admin_server.Env{
					ClusterID: "mockPachd",
					Config: &pachconfig.Configuration{
						GlobalConfiguration:             &config.GlobalConfiguration,
						PachdSpecificConfiguration:      &config.PachdSpecificConfiguration,
						EnterpriseSpecificConfiguration: &config.EnterpriseSpecificConfiguration,
					},
					PFSServer: pd.pfsServer,
					Paused:    false,
					DB:        env.DB,
				})
				return nil
			},
		},
		setupStep{
			Name: "initEnterpriseServer",
			Fn: func(ctx context.Context) error {
				var err error
				pd.enterpriseServer, err = ent_server.NewEnterpriseServer(
					&ent_server.Env{
						DB:         env.DB,
						Listener:   pd.dbListener,
						TxnEnv:     pd.txnEnv,
						EtcdClient: env.EtcdClient,
						EtcdPrefix: path.Join(config.EtcdPrefix, config.EnterpriseEtcdPrefix),
						AuthServer: pd.authServer.(auth_server.APIServer),
						GetKubeClient: func() kubernetes.Interface {
							return pd.kubeClient
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
				pd.licenseServer, err = license_server.New(&license_server.Env{
					DB:       env.DB,
					Listener: nil,
					Config: &pachconfig.Configuration{
						GlobalConfiguration:             &config.GlobalConfiguration,
						PachdSpecificConfiguration:      &config.PachdSpecificConfiguration,
						EnterpriseSpecificConfiguration: &config.EnterpriseSpecificConfiguration,
					},
					EnterpriseServer: pd.enterpriseServer,
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
				pd.debugServer = debug_server.NewDebugServer(debug_server.Env{
					DB:            env.DB,
					Name:          "testpachd",
					GetLokiClient: env.GetLokiClient,
					GetKubeClient: func() kubernetes.Interface {
						return pd.kubeClient
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
				pd.proxyServer = proxy_server.NewAPIServer(proxy_server.Env{
					Listener: pd.dbListener,
				})
				return nil
			},
		},
		setupStep{
			Name: "initLogsServer",
			Fn: func(ctx context.Context) error {
				var err error
				pd.logsServer, err = logs_server.NewAPIServer(logs_server.Env{
					GetLokiClient: env.GetLokiClient,
					AuthServer:    pd.authServer.(auth_server.APIServer),
				})
				if err != nil {
					return errors.Wrap(err, "logs_server.NewAPIServer")
				}
				return nil
			},
		},
		setupStep{
			Name: "initS3Server",
			Fn: func(ctx context.Context) error {
				router := s3.Router(ctx, s3.NewMasterDriver(), pd.mustGetPachClient)
				pd.s3Server = s3.Server(ctx, 0, router)
				return nil
			},
		},
		setupStep{
			Name: "initSnapshotServer",
			Fn: func(ctx context.Context) error {
				storageEnv := storage_server.Env{
					DB:     env.DB,
					Bucket: env.Bucket,
					Config: config.StorageConfiguration,
				}
				storageServer, err := storage_server.New(ctx, storageEnv)
				if err != nil {
					return errors.Wrap(err, "new storage")
				}
				pd.snapshotServer = &snapshot_server.APIServer{
					DB:    env.DB,
					Store: storageServer.Filesets,
				}
				return nil
			},
		},
		setupStep{
			Name: "initPJSWorkerAuthToken",
			Fn: func(ctx context.Context) error {
				if pd.config.PJSWorkerAuthToken == "" {
					return nil
				}
				ctx = auth_interceptor.AsInternalUser(ctx, authdb.InternalUser)
				_, err := pd.authServer.RestoreAuthToken(ctx, &auth.RestoreAuthTokenRequest{Token: &auth.TokenInfo{
					HashedToken: auth.HashToken(pd.config.PJSWorkerAuthToken),
					Subject:     auth.RobotPrefix + ":pjs-worker",
				}})
				if err != nil {
					return errors.Wrap(err, "authServer.RestoreAuthToken")
				}
				return nil
			},
		},
		// Workers
		initPFSWorker(&pd.pfsWorker, config.StorageConfiguration, func() pfs_server.WorkerEnv {
			etcdPrefix := path.Join(config.EtcdPrefix, config.PFSEtcdPrefix)
			return pfs_server.WorkerEnv{
				DB:          env.DB,
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
		admin.RegisterAPIServer(gs, pd.adminServer)
		auth.RegisterAPIServer(gs, pd.authServer)
		debug.RegisterDebugServer(gs, pd.debugServer)
		enterprise.RegisterAPIServer(gs, pd.enterpriseServer)
		grpc_health_v1.RegisterHealthServer(gs, pd.healthServer)
		license.RegisterAPIServer(gs, pd.licenseServer)
		logs.RegisterAPIServer(gs, pd.logsServer)
		metadata.RegisterAPIServer(gs, pd.metadataServer)
		pfs.RegisterAPIServer(gs, pd.pfsServer)
		pjs.RegisterAPIServer(gs, pd.pjsServer)
		snapshot.RegisterAPIServer(gs, pd.snapshotServer)
		storage.RegisterFilesetServer(gs, pd.storageServer)
		transaction.RegisterAPIServer(gs, pd.txnServer)
		pps.RegisterAPIServer(gs, pd.ppsServer)
		proxy.RegisterAPIServer(gs, pd.proxyServer)
		version.RegisterAPIServer(gs, pd.version)
	}))
	pd.addBackground("connect", func(ctx context.Context) error {
		addr, err := grpcutil.ParsePachdAddress("http://" + pd.env.Listener.Addr().String())
		if err != nil {
			return errors.Wrap(err, "parse pachd address")
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
			return errors.Wrap(err, "NewPachdFromAddress")
		}
		pd.pachClient = c // shallow copy, to avoid getting auth token set in the next block
		close(pd.pachClientReady)
		return nil
	})
	pd.addBackground("bootstrap", func(ctx context.Context) error {
		defer close(pd.authReady)
		return errors.Join(
			errors.Wrap(bootstrapIfAble(ctx, pd.licenseServer), "license"),
			errors.Wrap(bootstrapIfAble(ctx, pd.enterpriseServer), "enterprise"),
			errors.Wrap(bootstrapIfAble(ctx, pd.authServer), "auth"),
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
	<-pd.pachClientReady
	c := pd.pachClient.WithCtx(ctx)
	if t := pd.config.PJSWorkerAuthToken; t != "" {
		c.SetAuthToken(t)
	}
	if t := pd.config.AuthRootToken; t != "" {
		c.SetAuthToken(t)
	}
	return c, nil
}

// mustGetPachClient returns an unauthenticated client for internal use by API servers.
func (pd *Full) mustGetPachClient(ctx context.Context) *client.APIClient {
	<-pd.pachClientReady
	// Note: this code used to panic when ctx was Done, but sometimes API servers try to get a
	// pach client with a done context and are prepared for that error when they make an RPC,
	// rather than from here.  We should refactor the GetPachClient interface used by API
	// servers to return an error.  For that reason, we wait on pd.pachClientReady without the
	// context, since we can't return an error if ctx is done and there is no pach client.
	return pd.pachClient.WithCtx(ctx)
}

// AwaitAuth returns when auth is ready.  It must be called after Run() starts.
func (pd *Full) AwaitAuth(ctx context.Context) {
	select {
	case <-pd.authReady:
	case <-ctx.Done():
	}
}

// Snapshotter is meant for use in tests which need to get a snapshot.Snapshotter.
func (pd *Full) Snapshotter() *snapshot_server.Snapshotter {
	return &snapshot_server.Snapshotter{
		DB:      pd.env.DB,
		Storage: pd.storageServer.Filesets,
	}
}
