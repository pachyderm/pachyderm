package pachd

import (
	"context"
	"fmt"
	"net"
	godebug "runtime/debug"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	auth_interceptor "github.com/pachyderm/pachyderm/v2/src/internal/middleware/auth"
	log_interceptor "github.com/pachyderm/pachyderm/v2/src/internal/middleware/logging"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/profileutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	authserver "github.com/pachyderm/pachyderm/v2/src/server/auth/server"
	ent_server "github.com/pachyderm/pachyderm/v2/src/server/enterprise/server"
	pfs_server "github.com/pachyderm/pachyderm/v2/src/server/pfs/server"
	pps_server "github.com/pachyderm/pachyderm/v2/src/server/pps/server"
	txn_server "github.com/pachyderm/pachyderm/v2/src/server/transaction/server"
	"github.com/pachyderm/pachyderm/v2/src/transaction"
	"github.com/pachyderm/pachyderm/v2/src/version"
	etcd "go.etcd.io/etcd/client/v3"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func printVersion() setupStep {
	return setupStep{
		Name: "printVersion",
		Fn: func(ctx context.Context) error {
			log.Info(ctx, "version info", log.Proto("versionInfo", version.Version))
			return nil
		},
	}
}

func tweakResources(config pachconfig.GlobalConfiguration) setupStep {
	return setupStep{
		Name: "tweakResources",
		Fn: func(ctx context.Context) error {
			// set GOMAXPROCS to the container limit & log outcome to stdout
			maxprocs.Set(maxprocs.Logger(zap.S().Named("maxprocs").Infof)) //nolint:errcheck
			godebug.SetGCPercent(config.GCPercent)
			log.Info(ctx, "gc: set gc percent", zap.Int("value", config.GCPercent))
			setupMemoryLimit(ctx, config)
			return nil
		},
	}
}

func setupProfiling(name string, config *pachconfig.Configuration) setupStep {
	return setupStep{
		Name: "setupProfiling",
		Fn: func(ctx context.Context) error {
			profileutil.StartCloudProfiler(ctx, name, config)
			return nil
		},
	}
}

func initJaeger() setupStep {
	return setupStep{
		Name: "initJaegar",
		Fn: func(ctx context.Context) error {
			// must run InstallJaegerTracer before InitWithKube (otherwise InitWithKube
			// may create a pach client before tracing is active, not install the Jaeger
			// gRPC interceptor in the client, and not propagate traces)
			if endpoint := tracing.InstallJaegerTracerFromEnv(); endpoint != "" {
				log.Info(ctx, "connecting to Jaeger", zap.String("endpoint", endpoint))
			} else {
				log.Info(ctx, "no Jaeger collector found (JAEGER_COLLECTOR_SERVICE_HOST not set)")
			}
			return nil
		},
	}
}

func awaitDB(db *pachsql.DB) setupStep {
	return setupStep{
		Name: "awaitDB",
		Fn: func(ctx context.Context) error {
			return dbutil.WaitUntilReady(ctx, db)
		},
	}
}

func awaitMigrations(db *pachsql.DB) setupStep {
	return setupStep{
		Name: "awaitMigrations",
		Fn: func(ctx context.Context) error {
			return migrations.BlockUntil(ctx, db, clusterstate.DesiredClusterState)
		},
	}
}

func runMigrations(db *pachsql.DB, etcdClient *etcd.Client) setupStep {
	return setupStep{
		Name: "runMigrations",
		Fn: func(ctx context.Context) error {
			env := migrations.MakeEnv(nil, etcdClient)
			return migrations.ApplyMigrations(ctx, db, env, clusterstate.DesiredClusterState)
		},
	}
}

func newSelfGRPC(l net.Listener, opts []grpc.DialOption) *grpc.ClientConn {
	opts = append(opts, grpc.WithInsecure())
	gc, err := grpc.Dial(l.Addr().String(), opts...)
	if err != nil {
		// This is always a configuration issue, Dial does not initiate a network connection before returning.
		panic(err)
	}
	return gc
}

func initTransactionServer(out *transaction.APIServer, env func() txn_server.Env) setupStep {
	return setupStep{
		Name: "initTransactionServer",
		Fn: func(ctx context.Context) error {
			s, err := txn_server.NewAPIServer(env())
			if err != nil {
				return err
			}
			*out = s
			return nil
		},
	}
}

func initPFSAPIServer(out *pfs.APIServer, env func() pfs_server.Env) setupStep {
	return setupStep{
		Name: "initPFSAPIServer",
		Fn: func(ctx context.Context) error {
			apiServer, err := pfs_server.NewAPIServer(env())
			if err != nil {
				return err
			}
			*out = apiServer
			return nil
		},
	}
}

func initPPSAPIServer(out *pps.APIServer, env func() pps_server.Env) setupStep {
	return setupStep{
		Name: "initPPSServer",
		Fn: func(ctx context.Context) error {
			s, err := pps_server.NewAPIServerNoMaster(env())
			if err != nil {
				return err
			}
			*out = s
			return nil
		},
	}
}

func initPFSWorker(out **pfs_server.Worker, config pachconfig.StorageConfiguration, env func() pfs_server.WorkerEnv) setupStep {
	return setupStep{
		Name: "initPFSWorker",
		Fn: func(ctx context.Context) error {
			w, err := pfs_server.NewWorker(env(), pfs_server.WorkerConfig{Storage: config})
			if err != nil {
				return err
			}
			*out = w
			return nil
		},
	}
}

func initAuthServer(out *auth.APIServer, env func() authserver.Env) setupStep {
	return setupStep{
		Name: "initAuthServer",
		Fn: func(ctx context.Context) error {
			apiServer, err := authserver.NewAuthServer(
				env(),
				false, false, false,
			)
			if err != nil {
				return err
			}
			*out = apiServer
			return nil
		},
	}
}

func initEnterpriseServer(out *enterprise.APIServer, env func() *ent_server.Env) setupStep {
	return setupStep{
		Name: "initEnterpriseServer",
		Fn: func(ctx context.Context) error {
			apiServer, err := ent_server.NewEnterpriseServer(
				env(),
				ent_server.Config{
					Heartbeat: false,
				},
			)
			if err != nil {
				return err
			}
			*out = apiServer
			return nil
		},
	}
}

// newServeGRPC returns a background runner which servers gRPC on l.
// reg is called to register functions with the server.
func newServeGRPC(authInterceptor *auth_interceptor.Interceptor, l net.Listener, reg func(gs grpc.ServiceRegistrar)) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		loggingInterceptor := log_interceptor.NewBaseContextInterceptor(ctx)
		fmt.Println(loggingInterceptor)
		gs := grpc.NewServer(
		//grpc.ChainUnaryInterceptor(
		// 	errorsmw.UnaryServerInterceptor,
		// 	version_middleware.UnaryServerInterceptor,
		// 	tracing.UnaryServerInterceptor(),
		// 	authInterceptor.InterceptUnary,
		// loggingInterceptor.UnaryServerInterceptor,
		// 	validation.UnaryServerInterceptor,
		// ),
		// grpc.ChainStreamInterceptor(
		// 	errorsmw.StreamServerInterceptor,
		// 	version_middleware.StreamServerInterceptor,
		// 	tracing.StreamServerInterceptor(),
		// 	authInterceptor.InterceptStream,
		// loggingInterceptor.StreamServerInterceptor,
		// 	validation.StreamServerInterceptor,
		// ),
		)
		reg(gs)
		go func() {
			<-ctx.Done()
			log.Info(ctx, "stopping grpc server")
			gs.Stop()
		}()
		return gs.Serve(l)
	}
}
