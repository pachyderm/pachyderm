package local

import (
	"context"
	gotls "crypto/tls"
	"fmt"
	"net/http"
	"os"
	"path"
	"runtime/debug"
	"runtime/pprof"

	adminclient "github.com/pachyderm/pachyderm/v2/src/admin"
	authclient "github.com/pachyderm/pachyderm/v2/src/auth"
	debugclient "github.com/pachyderm/pachyderm/v2/src/debug"
	eprsclient "github.com/pachyderm/pachyderm/v2/src/enterprise"
	identityclient "github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/metrics"
	"github.com/pachyderm/pachyderm/v2/src/internal/middleware/auth"
	errorsmw "github.com/pachyderm/pachyderm/v2/src/internal/middleware/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/tls"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	licenseclient "github.com/pachyderm/pachyderm/v2/src/license"
	pfsclient "github.com/pachyderm/pachyderm/v2/src/pfs"
	ppsclient "github.com/pachyderm/pachyderm/v2/src/pps"
	adminserver "github.com/pachyderm/pachyderm/v2/src/server/admin/server"
	authserver "github.com/pachyderm/pachyderm/v2/src/server/auth/server"
	debugserver "github.com/pachyderm/pachyderm/v2/src/server/debug/server"
	eprsserver "github.com/pachyderm/pachyderm/v2/src/server/enterprise/server"
	identity_server "github.com/pachyderm/pachyderm/v2/src/server/identity/server"
	licenseserver "github.com/pachyderm/pachyderm/v2/src/server/license/server"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs/s3"
	pfs_server "github.com/pachyderm/pachyderm/v2/src/server/pfs/server"
	pps_server "github.com/pachyderm/pachyderm/v2/src/server/pps/server"
	txnserver "github.com/pachyderm/pachyderm/v2/src/server/transaction/server"
	transactionclient "github.com/pachyderm/pachyderm/v2/src/transaction"
	"github.com/pachyderm/pachyderm/v2/src/version"
	"github.com/pachyderm/pachyderm/v2/src/version/versionpb"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func RunLocal() (retErr error) {
	log.InitPachctlLogger()
	ctx := pctx.Background("local")

	config := &serviceenv.PachdFullConfiguration{}
	if err := cmdutil.Populate(config); err != nil {
		return err
	}

	config.PostgresSSL = "disable"

	defer func() {
		if retErr != nil {
			log.Error(ctx, "exiting with error", zap.Error(retErr))
			pprof.Lookup("goroutine").WriteTo(os.Stderr, 2) //nolint:errcheck
		}
	}()

	// must run InstallJaegerTracer before InitWithKube/pach client initialization
	if endpoint := tracing.InstallJaegerTracerFromEnv(); endpoint != "" {
		log.Debug(ctx, "connecting to Jaeger", zap.String("endpoint", endpoint))
	} else {
		log.Debug(ctx, "no Jaeger collector found (JAEGER_COLLECTOR_SERVICE_HOST not set)")
	}
	env := serviceenv.InitWithKube(ctx, serviceenv.NewConfiguration(config))
	debug.SetGCPercent(env.Config().GCPercent)
	env.InitDexDB()
	if env.Config().EtcdPrefix == "" {
		env.Config().EtcdPrefix = col.DefaultPrefix
	}

	// TODO: currently all pachds attempt to apply migrations, we should coordinate this
	if err := migrations.ApplyMigrations(context.Background(), env.GetDBClient(), migrations.MakeEnv(nil, env.GetEtcdClient()), clusterstate.DesiredClusterState); err != nil {
		return err
	}
	if err := migrations.BlockUntil(context.Background(), env.GetDBClient(), clusterstate.DesiredClusterState); err != nil {
		return err
	}

	var reporter *metrics.Reporter
	if env.Config().Metrics {
		reporter = metrics.NewReporter(env)
	}
	requireNoncriticalServers := !env.Config().RequireCriticalServersOnly

	// Setup External Pachd GRPC Server.
	authInterceptor := auth.NewInterceptor(env.AuthServer)
	externalServer, err := grpcutil.NewServer(
		pctx.Background("grpc.external"),
		true,
		grpc.ChainUnaryInterceptor(
			errorsmw.UnaryServerInterceptor,
			tracing.UnaryServerInterceptor(),
			authInterceptor.InterceptUnary,
		),
		grpc.ChainStreamInterceptor(
			errorsmw.StreamServerInterceptor,
			tracing.StreamServerInterceptor(),
			authInterceptor.InterceptStream,
		),
	)

	if err != nil {
		return err
	}

	if err := logGRPCServerSetup(ctx, "External Pachd", func() error {
		txnEnv := txnenv.New()
		if err := logGRPCServerSetup(ctx, "PFS API", func() error {
			pfsAPIServer, err := pfs_server.NewAPIServer(pfs_server.Env{})
			if err != nil {
				return err
			}
			pfsclient.RegisterAPIServer(externalServer.Server, pfsAPIServer)
			env.SetPfsServer(pfsAPIServer)
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup(ctx, "PPS API", func() error {
			ppsAPIServer, err := pps_server.NewAPIServer(
				pps_server.EnvFromServiceEnv(env, txnEnv, reporter),
			)
			if err != nil {
				return err
			}
			ppsclient.RegisterAPIServer(externalServer.Server, ppsAPIServer)
			env.SetPpsServer(ppsAPIServer)
			return nil
		}); err != nil {
			return err
		}

		if err := logGRPCServerSetup(ctx, "Identity API", func() error {
			idAPIServer := identity_server.NewIdentityServer(
				identity_server.Env{},
				true,
			)
			if err != nil {
				return err
			}
			identityclient.RegisterAPIServer(externalServer.Server, idAPIServer)
			env.SetIdentityServer(idAPIServer)
			return nil
		}); err != nil {
			return err
		}

		if err := logGRPCServerSetup(ctx, "Auth API", func() error {
			authAPIServer, err := authserver.NewAuthServer(
				authserver.EnvFromServiceEnv(env, txnEnv),
				true, requireNoncriticalServers, true)
			if err != nil {
				return err
			}
			authclient.RegisterAPIServer(externalServer.Server, authAPIServer)
			env.SetAuthServer(authAPIServer)
			return nil
		}); err != nil {
			return err
		}
		var transactionAPIServer txnserver.APIServer
		if err := logGRPCServerSetup(ctx, "Transaction API", func() error {
			transactionAPIServer, err = txnserver.NewAPIServer(
				env,
				txnEnv,
			)
			if err != nil {
				return err
			}
			transactionclient.RegisterAPIServer(externalServer.Server, transactionAPIServer)
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup(ctx, "Enterprise API", func() error {
			enterpriseAPIServer, err := eprsserver.NewEnterpriseServer(
				eprsserver.EnvFromServiceEnv(env, path.Join(env.Config().EtcdPrefix, env.Config().EnterpriseEtcdPrefix), txnEnv), true)
			if err != nil {
				return err
			}
			eprsclient.RegisterAPIServer(externalServer.Server, enterpriseAPIServer)
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup(ctx, "License API", func() error {
			licenseAPIServer, err := licenseserver.New(&licenseserver.Env{})
			if err != nil {
				return err
			}
			licenseclient.RegisterAPIServer(externalServer.Server, licenseAPIServer)
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup(ctx, "Admin API", func() error {
			adminclient.RegisterAPIServer(externalServer.Server, adminserver.NewAPIServer(adminserver.Env{}))
			return nil
		}); err != nil {
			return err
		}
		healthServer := health.NewServer()
		healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		if err := logGRPCServerSetup(ctx, "Health", func() error {
			grpc_health_v1.RegisterHealthServer(externalServer.Server, healthServer)
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup(ctx, "Version API", func() error {
			versionpb.RegisterAPIServer(externalServer.Server, version.NewAPIServer(version.Version, version.APIServerOptions{}))
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup(ctx, "Debug", func() error {
			debugclient.RegisterDebugServer(externalServer.Server, debugserver.NewDebugServer(
				env,
				env.Config().PachdPodName,
				nil,
				env.GetDBClient(),
			))
			return nil
		}); err != nil {
			return err
		}
		txnEnv.Initialize(env, transactionAPIServer)
		log.Info(ctx, "listening on port", zap.Uint16("port", env.Config().Port))
		if _, err := externalServer.ListenTCP("", env.Config().Port); err != nil {
			return err
		}
		healthServer.Resume()
		return nil
	}); err != nil {
		return err
	}
	// Setup Internal Pachd GRPC Server.
	internalServer, err := grpcutil.NewServer(pctx.Background("grpc.internal"), false, grpc.ChainUnaryInterceptor(tracing.UnaryServerInterceptor(), authInterceptor.InterceptUnary), grpc.StreamInterceptor(authInterceptor.InterceptStream))
	if err != nil {
		return err
	}
	if err := logGRPCServerSetup(ctx, "Internal Pachd", func() error {
		txnEnv := txnenv.New()
		if err := logGRPCServerSetup(ctx, "PFS API", func() error {
			pfsAPIServer, err := pfs_server.NewAPIServer(
				pfs_server.Env{},
			)
			if err != nil {
				return err
			}
			pfsclient.RegisterAPIServer(internalServer.Server, pfsAPIServer)
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup(ctx, "PPS API", func() error {
			ppsAPIServer, err := pps_server.NewAPIServer(
				pps_server.EnvFromServiceEnv(env, txnEnv, reporter),
			)
			if err != nil {
				return err
			}
			ppsclient.RegisterAPIServer(internalServer.Server, ppsAPIServer)
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup(ctx, "Identity API", func() error {
			idAPIServer := identity_server.NewIdentityServer(
				identity_server.Env{},
				false,
			)
			identityclient.RegisterAPIServer(internalServer.Server, idAPIServer)
			env.SetIdentityServer(idAPIServer)
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup(ctx, "Auth API", func() error {
			authAPIServer, err := authserver.NewAuthServer(
				authserver.EnvFromServiceEnv(env, txnEnv),
				false,
				requireNoncriticalServers,
				true,
			)
			if err != nil {
				return err
			}
			authclient.RegisterAPIServer(internalServer.Server, authAPIServer)
			return nil
		}); err != nil {
			return err
		}
		var transactionAPIServer txnserver.APIServer
		if err := logGRPCServerSetup(ctx, "Transaction API", func() error {
			transactionAPIServer, err = txnserver.NewAPIServer(
				env,
				txnEnv,
			)
			if err != nil {
				return err
			}
			transactionclient.RegisterAPIServer(internalServer.Server, transactionAPIServer)
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup(ctx, "Enterprise API", func() error {
			enterpriseAPIServer, err := eprsserver.NewEnterpriseServer(
				eprsserver.EnvFromServiceEnv(env, path.Join(env.Config().EtcdPrefix, env.Config().EnterpriseEtcdPrefix), txnEnv), false)
			if err != nil {
				return err
			}
			eprsclient.RegisterAPIServer(internalServer.Server, enterpriseAPIServer)
			return nil
		}); err != nil {
			return err
		}
		healthServer := health.NewServer()
		healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)

		if err := logGRPCServerSetup(ctx, "Health", func() error {
			grpc_health_v1.RegisterHealthServer(internalServer.Server, healthServer)
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup(ctx, "Version API", func() error {
			versionpb.RegisterAPIServer(internalServer.Server, version.NewAPIServer(version.Version, version.APIServerOptions{}))
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup(ctx, "Admin API", func() error {
			adminclient.RegisterAPIServer(internalServer.Server, adminserver.NewAPIServer(adminserver.Env{}))
			return nil
		}); err != nil {
			return err
		}
		txnEnv.Initialize(env, transactionAPIServer)
		if _, err := internalServer.ListenTCP("", env.Config().PeerPort); err != nil {
			return err
		}
		healthServer.Resume()
		return nil
	}); err != nil {
		return err
	}
	// Create the goroutines for the servers.
	// Any server error is considered critical and will cause Pachd to exit.
	// The first server that errors will have its error message logged.
	errChan := make(chan error, 1)
	go waitForError(ctx, "External Pachd GRPC Server", errChan, true, func() error {
		return externalServer.Wait()
	})
	go waitForError(ctx, "Internal Pachd GRPC Server", errChan, true, func() error {
		return internalServer.Wait()
	})
	go waitForError(ctx, "S3 Server", errChan, requireNoncriticalServers, func() error {
		router := s3.Router(ctx, s3.NewMasterDriver(), env.GetPachClient)
		server := s3.Server(ctx, env.Config().S3GatewayPort, router)

		if err != nil {
			return err
		}
		certPath, keyPath, err := tls.GetCertPaths()
		if err != nil {
			log.Info(ctx, "s3gateway TLS disabled", zap.NamedError("reason", err))
			return errors.EnsureStack(server.ListenAndServe())
		}
		cLoader := tls.NewCertLoader(certPath, keyPath, tls.CertCheckFrequency)
		// Read TLS cert and key
		err = cLoader.LoadAndStart()
		if err != nil {
			return errors.Wrapf(err, "couldn't load TLS cert for s3gateway: %v", err)
		}
		server.TLSConfig = &gotls.Config{GetCertificate: cLoader.GetCertificate}
		return errors.EnsureStack(server.ListenAndServeTLS(certPath, keyPath))
	})
	go waitForError(ctx, "Prometheus Server", errChan, requireNoncriticalServers, func() error {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		return errors.EnsureStack(http.ListenAndServe(fmt.Sprintf(":%v", env.Config().PrometheusPort), mux))
	})
	return <-errChan
}

func logGRPCServerSetup(ctx context.Context, name string, f func() error) (retErr error) {
	log.Debug(ctx, "started setting up GRPC Server", zap.String("name", name))
	defer func() {
		if retErr != nil {
			retErr = errors.Wrapf(retErr, "error setting up %v GRPC Server", name)
		} else {
			log.Debug(ctx, "finished setting up GRPC Server", zap.String("name", name))
		}
	}()
	return f()
}

func waitForError(ctx context.Context, name string, errChan chan error, required bool, f func() error) {
	if err := f(); !errors.Is(err, http.ErrServerClosed) {
		if !required {
			log.Error(ctx, "error setting up and/or running server", zap.String("server", name), zap.Error(err))
		} else {
			errChan <- errors.Wrapf(err, "error setting up and/or running %v (use --require-critical-servers-only deploy flag to ignore errors from noncritical servers)", name)
		}
	}
}
