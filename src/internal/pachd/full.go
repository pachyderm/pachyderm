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
	"k8s.io/client-go/kubernetes"

	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	auth_interceptor "github.com/pachyderm/pachyderm/v2/src/internal/middleware/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/license"
	"github.com/pachyderm/pachyderm/v2/src/metadata"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	admin_server "github.com/pachyderm/pachyderm/v2/src/server/admin/server"
	auth_iface "github.com/pachyderm/pachyderm/v2/src/server/auth"
	auth_server "github.com/pachyderm/pachyderm/v2/src/server/auth/server"
	debug_server "github.com/pachyderm/pachyderm/v2/src/server/debug/server"
	entiface "github.com/pachyderm/pachyderm/v2/src/server/enterprise"
	ent_server "github.com/pachyderm/pachyderm/v2/src/server/enterprise/server"
	license_server "github.com/pachyderm/pachyderm/v2/src/server/license/server"
	metadata_server "github.com/pachyderm/pachyderm/v2/src/server/metadata/server"
	pfsiface "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	pfs_server "github.com/pachyderm/pachyderm/v2/src/server/pfs/server"
	ppsiface "github.com/pachyderm/pachyderm/v2/src/server/pps"
	pps_server "github.com/pachyderm/pachyderm/v2/src/server/pps/server"
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
	DB         *pachsql.DB
	DirectDB   *pachsql.DB
	ObjClient  obj.Client
	Bucket     *obj.Bucket
	EtcdClient *clientv3.Client
	Listener   net.Listener
}

type Full struct {
	base
	env    Env
	config pachconfig.PachdFullConfiguration

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
	// TODO
	// debugSrv debug.DebugServer

	pfsWorker   *pfs_server.Worker
	ppsWorker   *pps_server.Worker
	debugWorker *debug_server.Worker
}

func (pd *Full) PachClient(ctx context.Context) (*client.APIClient, error) {
	addr, err := grpcutil.ParsePachdAddress("http://" + pd.env.Listener.Addr().String())
	if err != nil {
		return nil, errors.Wrap(err, "parse pachd address")
	}
	return client.NewFromPachdAddress(ctx, addr)
}

// NewFull sets up a new Full pachd and returns it.
func NewFull(env Env, config pachconfig.PachdFullConfiguration) *Full {
	pd := &Full{env: env, config: config}

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

	pd.addSetup(
		printVersion(),
		setupProfiling("pachd", pachconfig.NewConfiguration(config)),
		tweakResources(config.GlobalConfiguration),
		initJaeger(),

		awaitDB(env.DB),
		runMigrations(env.DirectDB, env.EtcdClient),
		awaitMigrations(env.DB),

		// API Servers
		initTransactionServer(&pd.txnSrv, func() txn_server.Env {
			return txn_server.Env{
				DB:         env.DB,
				PGListener: nil,
				TxnEnv:     pd.txnEnv,
			}
		}),
		initAuthServer(&pd.authSrv, func() auth_server.Env {
			return auth_server.Env{
				DB:         env.DB,
				EtcdClient: env.EtcdClient,
				Listener:   nil,
				TxnEnv:     pd.txnEnv,
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
		initPFSAPIServer(&pd.pfsSrv, func() pfs_server.Env {
			etcdPrefix := path.Join(config.EtcdPrefix, config.PFSEtcdPrefix)
			return pfs_server.Env{
				DB:           env.DB,
				Bucket:       env.Bucket,
				ObjectClient: env.ObjClient,
				EtcdClient:   env.EtcdClient,
				EtcdPrefix:   etcdPrefix,
				TaskService:  task.NewEtcdService(env.EtcdClient, etcdPrefix),

				TxnEnv:        pd.txnEnv,
				StorageConfig: config.StorageConfiguration,
				Auth:          pd.authSrv.(pfs_server.PFSAuth),
				GetPipelineInspector: func() pfs_server.PipelineInspector {
					panic("GetPipelineInspector")
				},
				GetPPSServer: func() ppsiface.APIServer { return pd.ppsSrv.(pps_server.APIServer) },
			}
		}),
		initPPSAPIServer(&pd.ppsSrv, func() pps_server.Env {
			return pps_server.Env{
				BackgroundContext: pctx.TODO(),
				AuthServer:        pd.authSrv.(auth_server.APIServer),
				DB:                pd.env.DB,
				Config: pachconfig.Configuration{
					GlobalConfiguration:        &config.GlobalConfiguration,
					PachdSpecificConfiguration: &config.PachdSpecificConfiguration,
				},
				PFSServer: pd.pfsSrv.(pfs_server.APIServer),
				TxnEnv:    pd.txnEnv,
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
						GlobalConfiguration:        &config.GlobalConfiguration,
						PachdSpecificConfiguration: &config.PachdSpecificConfiguration,
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
						Listener:   nil,
						TxnEnv:     pd.txnEnv,
						EtcdClient: env.EtcdClient,
						EtcdPrefix: path.Join(config.EtcdPrefix, config.EnterpriseEtcdPrefix),
						AuthServer: pd.authSrv.(auth_server.APIServer),
						GetKubeClient: func() kubernetes.Interface {
							panic("attempt to do k8s things from TestPachd")
						},
						GetPachClient: func(ctx context.Context) *client.APIClient {
							c, err := pd.PachClient(ctx)
							if err != nil {
								panic(fmt.Sprintf("build enterprise pach client: %v", err))
							}
							return c
						},
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
					DB:       pd.env.DB,
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
	pd.addBackground("ppsWorker", func(ctx context.Context) error {
		return pd.ppsWorker.Run(ctx)
	})
	pd.addBackground("debugWorker", func(ctx context.Context) error {
		return pd.debugWorker.Run(ctx)
	})
	pd.addBackground("grpc", newServeGRPC(pd.authInterceptor, env.Listener, func(gs grpc.ServiceRegistrar) {
		grpc_health_v1.RegisterHealthServer(gs, pd.healthSrv)
		version.RegisterAPIServer(gs, pd.version)
		auth.RegisterAPIServer(gs, pd.authSrv)
		pfs.RegisterAPIServer(gs, pd.pfsSrv)
		pps.RegisterAPIServer(gs, pd.ppsSrv)
		metadata.RegisterAPIServer(gs, pd.metadataSrv)
		admin.RegisterAPIServer(gs, pd.adminSrv)
		enterprise.RegisterAPIServer(gs, pd.enterpriseSrv)
		license.RegisterAPIServer(gs, pd.licenseSrv)
	}))
	return pd
}
