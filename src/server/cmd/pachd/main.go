package main

import (
	gotls "crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"path"
	"runtime/debug"
	"runtime/pprof"

	adminclient "github.com/pachyderm/pachyderm/v2/src/admin"
	authclient "github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	debugclient "github.com/pachyderm/pachyderm/v2/src/debug"
	eprsclient "github.com/pachyderm/pachyderm/v2/src/enterprise"
	healthclient "github.com/pachyderm/pachyderm/v2/src/health"
	identityclient "github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/deploy/assets"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	logutil "github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/metrics"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/netutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/tls"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	licenseclient "github.com/pachyderm/pachyderm/v2/src/license"
	pfsclient "github.com/pachyderm/pachyderm/v2/src/pfs"
	ppsclient "github.com/pachyderm/pachyderm/v2/src/pps"
	adminserver "github.com/pachyderm/pachyderm/v2/src/server/admin/server"
	authserver "github.com/pachyderm/pachyderm/v2/src/server/auth/server"
	debugserver "github.com/pachyderm/pachyderm/v2/src/server/debug/server"
	eprsserver "github.com/pachyderm/pachyderm/v2/src/server/enterprise/server"
	"github.com/pachyderm/pachyderm/v2/src/server/health"
	pach_http "github.com/pachyderm/pachyderm/v2/src/server/http"
	identity_server "github.com/pachyderm/pachyderm/v2/src/server/identity/server"
	licenseserver "github.com/pachyderm/pachyderm/v2/src/server/license/server"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs/s3"
	pfs_server "github.com/pachyderm/pachyderm/v2/src/server/pfs/server"
	pps_server "github.com/pachyderm/pachyderm/v2/src/server/pps/server"
	"github.com/pachyderm/pachyderm/v2/src/server/pps/server/githook"
	txnserver "github.com/pachyderm/pachyderm/v2/src/server/transaction/server"
	transactionclient "github.com/pachyderm/pachyderm/v2/src/transaction"
	"github.com/pachyderm/pachyderm/v2/src/version"
	"github.com/pachyderm/pachyderm/v2/src/version/versionpb"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var mode string
var readiness bool

func init() {
	flag.StringVar(&mode, "mode", "full", "Pachd currently supports two modes: full and sidecar.  The former includes everything you need in a full pachd node.  The latter runs only PFS, the Auth service, and a stripped-down version of PPS.")
	flag.BoolVar(&readiness, "readiness", false, "Run readiness check.")
	flag.Parse()
}

func main() {
	log.SetFormatter(logutil.FormatterFunc(logutil.Pretty))

	switch {
	case readiness:
		cmdutil.Main(doReadinessCheck, &serviceenv.PachdFullConfiguration{})
	case mode == "full":
		cmdutil.Main(doFullMode, &serviceenv.PachdFullConfiguration{})
	case mode == "sidecar":
		cmdutil.Main(doSidecarMode, &serviceenv.PachdFullConfiguration{})
	default:
		fmt.Printf("unrecognized mode: %s\n", mode)
	}
}

func doReadinessCheck(config interface{}) error {
	env := serviceenv.InitPachOnlyEnv(serviceenv.NewConfiguration(config))
	return env.GetPachClient(context.Background()).Health()
}

func doSidecarMode(config interface{}) (retErr error) {
	defer func() {
		if retErr != nil {
			pprof.Lookup("goroutine").WriteTo(os.Stderr, 2)
		}
	}()
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
	env := serviceenv.InitWithKube(serviceenv.NewConfiguration(config))
	debug.SetGCPercent(env.GCPercent)
	if env.EtcdPrefix == "" {
		env.EtcdPrefix = col.DefaultPrefix
	}
	clusterID, err := getClusterID(env.GetEtcdClient())
	if err != nil {
		return errors.Wrapf(err, "getClusterID")
	}
	var reporter *metrics.Reporter
	if env.Metrics {
		reporter = metrics.NewReporter(clusterID, env)
	}
	authInterceptor := auth.NewInterceptor(env)
	server, err := grpcutil.NewServer(
		context.Background(),
		false,
		grpc.ChainUnaryInterceptor(
			tracing.UnaryServerInterceptor(),
			authInterceptor.InterceptUnary,
		),
		grpc.ChainStreamInterceptor(
			tracing.StreamServerInterceptor(),
			authInterceptor.InterceptStream,
		),
	)
	if err != nil {
		return err
	}
	txnEnv := &txnenv.TransactionEnv{}
	var pfsAPIServer pfs_server.APIServer
	if err := logGRPCServerSetup("PFS API", func() error {
		pfsAPIServer, err = pfs_server.NewAPIServer(
			env,
			txnEnv,
			path.Join(env.EtcdPrefix, env.PFSEtcdPrefix),
			env.GetDBClient(),
		)
		if err != nil {
			return err
		}
		pfsclient.RegisterAPIServer(server.Server, pfsAPIServer)
		return nil
	}); err != nil {
		return err
	}
	var ppsAPIServer pps_server.APIServer
	if err := logGRPCServerSetup("PPS API", func() error {
		ppsAPIServer, err = pps_server.NewSidecarAPIServer(
			env,
			txnEnv,
			path.Join(env.EtcdPrefix, env.PPSEtcdPrefix),
			env.Namespace,
			env.IAMRole,
			reporter,
			env.PPSWorkerPort,
			env.HTTPPort,
			env.PeerPort,
		)
		if err != nil {
			return err
		}
		ppsclient.RegisterAPIServer(server.Server, ppsAPIServer)
		return nil
	}); err != nil {
		return err
	}
	var authAPIServer authserver.APIServer
	if err := logGRPCServerSetup("Auth API", func() error {
		authAPIServer, err = authserver.NewAuthServer(
			env,
			txnEnv,
			path.Join(env.EtcdPrefix, env.AuthEtcdPrefix),
			false,
			false,
		)
		if err != nil {
			return err
		}
		authclient.RegisterAPIServer(server.Server, authAPIServer)
		return nil
	}); err != nil {
		return err
	}
	var transactionAPIServer txnserver.APIServer
	if err := logGRPCServerSetup("Transaction API", func() error {
		transactionAPIServer, err = txnserver.NewAPIServer(
			env,
			txnEnv,
			path.Join(env.EtcdPrefix, env.PFSEtcdPrefix),
		)
		if err != nil {
			return err
		}
		transactionclient.RegisterAPIServer(server.Server, transactionAPIServer)
		return nil
	}); err != nil {
		return err
	}
	if err := logGRPCServerSetup("Enterprise API", func() error {
		enterpriseAPIServer, err := eprsserver.NewEnterpriseServer(
			env, path.Join(env.EtcdPrefix, env.EnterpriseEtcdPrefix))
		if err != nil {
			return err
		}
		eprsclient.RegisterAPIServer(server.Server, enterpriseAPIServer)
		return nil
	}); err != nil {
		return err
	}
	if err := logGRPCServerSetup("Health", func() error {
		healthclient.RegisterHealthServer(server.Server, health.NewHealthServer())
		return nil
	}); err != nil {
		return err
	}
	if err := logGRPCServerSetup("Debug", func() error {
		debugclient.RegisterDebugServer(server.Server, debugserver.NewDebugServer(
			env,
			env.PachdPodName,
			nil,
		))
		return nil
	}); err != nil {
		return err
	}
	txnEnv.Initialize(env, transactionAPIServer, authAPIServer, pfsAPIServer, ppsAPIServer)
	// The sidecar only needs to serve traffic on the peer port, as it only serves
	// traffic from the user container (the worker binary and occasionally user
	// pipelines)
	if _, err := server.ListenTCP("", env.PeerPort); err != nil {
		return err
	}
	return server.Wait()
}

func doFullMode(config interface{}) (retErr error) {
	defer func() {
		if retErr != nil {
			pprof.Lookup("goroutine").WriteTo(os.Stderr, 2)
		}
	}()
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

	// must run InstallJaegerTracer before InitWithKube/pach client initialization
	if endpoint := tracing.InstallJaegerTracerFromEnv(); endpoint != "" {
		log.Printf("connecting to Jaeger at %q", endpoint)
	} else {
		log.Printf("no Jaeger collector found (JAEGER_COLLECTOR_SERVICE_HOST not set)")
	}
	env := serviceenv.InitWithKube(serviceenv.NewConfiguration(config))
	debug.SetGCPercent(env.GCPercent)
	if env.EtcdPrefix == "" {
		env.EtcdPrefix = col.DefaultPrefix
	}

	// TODO: currently all pachds attempt to apply migrations, we should coordinate this
	if err := migrations.ApplyMigrations(context.Background(), env.GetDBClient(), migrations.Env{}, clusterstate.DesiredClusterState); err != nil {
		return err
	}
	if err := migrations.BlockUntil(context.Background(), env.GetDBClient(), clusterstate.DesiredClusterState); err != nil {
		return err
	}

	clusterID, err := getClusterID(env.GetEtcdClient())
	if err != nil {
		return errors.Wrapf(err, "getClusterID")
	}
	var reporter *metrics.Reporter
	if env.Metrics {
		reporter = metrics.NewReporter(clusterID, env)
	}
	etcdAddress := fmt.Sprintf("http://%s", net.JoinHostPort(env.EtcdHost, env.EtcdPort))
	ip, err := netutil.ExternalIP()
	if err != nil {
		return errors.Wrapf(err, "error getting pachd external ip")
	}
	address := net.JoinHostPort(ip, fmt.Sprintf("%d", env.PeerPort))
	kubeNamespace := env.Namespace
	requireNoncriticalServers := !env.RequireCriticalServersOnly

	// Setup External Pachd GRPC Server.
	authInterceptor := auth.NewInterceptor(env)
	externalServer, err := grpcutil.NewServer(
		context.Background(),
		true,
		grpc.ChainUnaryInterceptor(
			tracing.UnaryServerInterceptor(),
			authInterceptor.InterceptUnary,
		),
		grpc.ChainStreamInterceptor(
			tracing.StreamServerInterceptor(),
			authInterceptor.InterceptStream,
		),
	)

	if err != nil {
		return err
	}

	identityStorageProvider := identity_server.NewLazyPostgresStorage(
		env.PostgresServiceHost,
		env.IdentityServerDatabase,
		env.IdentityServerUser,
		env.IdentityServerPassword,
		env.PostgresServiceSSL,
		env.PostgresServicePort,
	)

	if err := logGRPCServerSetup("External Pachd", func() error {
		txnEnv := &txnenv.TransactionEnv{}
		var pfsAPIServer pfs_server.APIServer
		if err := logGRPCServerSetup("PFS API", func() error {
			pfsAPIServer, err = pfs_server.NewAPIServer(env, txnEnv, path.Join(env.EtcdPrefix, env.PFSEtcdPrefix), env.GetDBClient())
			if err != nil {
				return err
			}
			pfsclient.RegisterAPIServer(externalServer.Server, pfsAPIServer)
			return nil
		}); err != nil {
			return err
		}
		var ppsAPIServer pps_server.APIServer
		if err := logGRPCServerSetup("PPS API", func() error {
			ppsAPIServer, err = pps_server.NewAPIServer(
				env,
				txnEnv,
				path.Join(env.EtcdPrefix, env.PPSEtcdPrefix),
				kubeNamespace,
				env.WorkerImage,
				env.WorkerSidecarImage,
				env.WorkerImagePullPolicy,
				env.StorageRoot,
				env.StorageBackend,
				env.StorageHostPath,
				env.CacheRoot,
				env.IAMRole,
				env.ImagePullSecret,
				env.NoExposeDockerSocket,
				reporter,
				env.WorkerUsesRoot,
				env.PPSWorkerPort,
				env.Port,
				env.HTTPPort,
				env.PeerPort,
				env.GCPercent,
			)
			if err != nil {
				return err
			}
			ppsclient.RegisterAPIServer(externalServer.Server, ppsAPIServer)
			return nil
		}); err != nil {
			return err
		}

		if err := logGRPCServerSetup("Identity API", func() error {
			idAPIServer, err := identity_server.NewIdentityServer(
				env,
				identityStorageProvider,
				true,
				path.Join(env.EtcdPrefix, env.IdentityEtcdPrefix),
			)
			if err != nil {
				return err
			}
			identityclient.RegisterAPIServer(externalServer.Server, idAPIServer)
			return nil
		}); err != nil {
			return err
		}

		var authAPIServer authserver.APIServer
		if err := logGRPCServerSetup("Auth API", func() error {
			authAPIServer, err = authserver.NewAuthServer(
				env, txnEnv, path.Join(env.EtcdPrefix, env.AuthEtcdPrefix), true, requireNoncriticalServers)
			if err != nil {
				return err
			}
			authclient.RegisterAPIServer(externalServer.Server, authAPIServer)
			return nil
		}); err != nil {
			return err
		}
		var transactionAPIServer txnserver.APIServer
		if err := logGRPCServerSetup("Transaction API", func() error {
			transactionAPIServer, err = txnserver.NewAPIServer(
				env,
				txnEnv,
				path.Join(env.EtcdPrefix, env.PFSEtcdPrefix),
			)
			if err != nil {
				return err
			}
			transactionclient.RegisterAPIServer(externalServer.Server, transactionAPIServer)
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup("Enterprise API", func() error {
			enterpriseAPIServer, err := eprsserver.NewEnterpriseServer(
				env, path.Join(env.EtcdPrefix, env.EnterpriseEtcdPrefix))
			if err != nil {
				return err
			}
			eprsclient.RegisterAPIServer(externalServer.Server, enterpriseAPIServer)
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup("License API", func() error {
			licenseAPIServer, err := licenseserver.New(
				env, path.Join(env.EtcdPrefix, env.EnterpriseEtcdPrefix))
			if err != nil {
				return err
			}
			licenseclient.RegisterAPIServer(externalServer.Server, licenseAPIServer)
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup("Admin API", func() error {
			adminclient.RegisterAPIServer(externalServer.Server, adminserver.NewAPIServer(&adminclient.ClusterInfo{
				ID:           clusterID,
				DeploymentID: env.DeploymentID,
			}))
			return nil
		}); err != nil {
			return err
		}
		healthServer := health.NewHealthServer()
		if err := logGRPCServerSetup("Health", func() error {
			healthclient.RegisterHealthServer(externalServer.Server, healthServer)
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup("Version API", func() error {
			versionpb.RegisterAPIServer(externalServer.Server, version.NewAPIServer(version.Version, version.APIServerOptions{}))
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup("Debug", func() error {
			debugclient.RegisterDebugServer(externalServer.Server, debugserver.NewDebugServer(
				env,
				env.PachdPodName,
				nil,
			))
			return nil
		}); err != nil {
			return err
		}
		txnEnv.Initialize(env, transactionAPIServer, authAPIServer, pfsAPIServer, ppsAPIServer)
		if _, err := externalServer.ListenTCP("", env.Port); err != nil {
			return err
		}
		healthServer.Ready()
		return nil
	}); err != nil {
		return err
	}
	// Setup Internal Pachd GRPC Server.
	internalServer, err := grpcutil.NewServer(context.Background(), false, grpc.ChainUnaryInterceptor(tracing.UnaryServerInterceptor(), authInterceptor.InterceptUnary), grpc.StreamInterceptor(authInterceptor.InterceptStream))
	if err != nil {
		return err
	}
	if err := logGRPCServerSetup("Internal Pachd", func() error {
		txnEnv := &txnenv.TransactionEnv{}
		var pfsAPIServer pfs_server.APIServer
		if err := logGRPCServerSetup("PFS API", func() error {
			pfsAPIServer, err = pfs_server.NewAPIServer(
				env,
				txnEnv,
				path.Join(env.EtcdPrefix, env.PFSEtcdPrefix),
				env.GetDBClient(),
			)
			if err != nil {
				return err
			}
			pfsclient.RegisterAPIServer(internalServer.Server, pfsAPIServer)
			return nil
		}); err != nil {
			return err
		}
		var ppsAPIServer pps_server.APIServer
		if err := logGRPCServerSetup("PPS API", func() error {
			ppsAPIServer, err = pps_server.NewAPIServer(
				env,
				txnEnv,
				path.Join(env.EtcdPrefix, env.PPSEtcdPrefix),
				kubeNamespace,
				env.WorkerImage,
				env.WorkerSidecarImage,
				env.WorkerImagePullPolicy,
				env.StorageRoot,
				env.StorageBackend,
				env.StorageHostPath,
				env.CacheRoot,
				env.IAMRole,
				env.ImagePullSecret,
				env.NoExposeDockerSocket,
				reporter,
				env.WorkerUsesRoot,
				env.PPSWorkerPort,
				env.Port,
				env.HTTPPort,
				env.PeerPort,
				env.GCPercent,
			)
			if err != nil {
				return err
			}
			ppsclient.RegisterAPIServer(internalServer.Server, ppsAPIServer)
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup("Identity API", func() error {
			idAPIServer, err := identity_server.NewIdentityServer(
				env,
				identityStorageProvider,
				false,
				path.Join(env.EtcdPrefix, env.IdentityEtcdPrefix),
			)
			if err != nil {
				return err
			}
			identityclient.RegisterAPIServer(internalServer.Server, idAPIServer)
			return nil
		}); err != nil {
			return err
		}
		var authAPIServer authserver.APIServer
		if err := logGRPCServerSetup("Auth API", func() error {
			authAPIServer, err = authserver.NewAuthServer(
				env,
				txnEnv,
				path.Join(env.EtcdPrefix, env.AuthEtcdPrefix),
				false,
				requireNoncriticalServers,
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
		if err := logGRPCServerSetup("Transaction API", func() error {
			transactionAPIServer, err = txnserver.NewAPIServer(
				env,
				txnEnv,
				path.Join(env.EtcdPrefix, env.PFSEtcdPrefix),
			)
			if err != nil {
				return err
			}
			transactionclient.RegisterAPIServer(internalServer.Server, transactionAPIServer)
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup("Enterprise API", func() error {
			enterpriseAPIServer, err := eprsserver.NewEnterpriseServer(
				env, path.Join(env.EtcdPrefix, env.EnterpriseEtcdPrefix))
			if err != nil {
				return err
			}
			eprsclient.RegisterAPIServer(internalServer.Server, enterpriseAPIServer)
			return nil
		}); err != nil {
			return err
		}
		healthServer := health.NewHealthServer()
		if err := logGRPCServerSetup("Health", func() error {
			healthclient.RegisterHealthServer(internalServer.Server, healthServer)
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup("Version API", func() error {
			versionpb.RegisterAPIServer(internalServer.Server, version.NewAPIServer(version.Version, version.APIServerOptions{}))
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup("Admin API", func() error {
			adminclient.RegisterAPIServer(internalServer.Server, adminserver.NewAPIServer(&adminclient.ClusterInfo{
				ID:           clusterID,
				DeploymentID: env.DeploymentID,
			}))
			return nil
		}); err != nil {
			return err
		}
		txnEnv.Initialize(env, transactionAPIServer, authAPIServer, pfsAPIServer, ppsAPIServer)
		if _, err := internalServer.ListenTCP("", env.PeerPort); err != nil {
			return err
		}
		healthServer.Ready()
		return nil
	}); err != nil {
		return err
	}
	// Create the goroutines for the servers.
	// Any server error is considered critical and will cause Pachd to exit.
	// The first server that errors will have its error message logged.
	errChan := make(chan error, 1)
	go waitForError("External Pachd GRPC Server", errChan, true, func() error {
		return externalServer.Wait()
	})
	go waitForError("Internal Pachd GRPC Server", errChan, true, func() error {
		return internalServer.Wait()
	})
	go waitForError("HTTP Server", errChan, requireNoncriticalServers, func() error {
		httpServer, err := pach_http.NewHTTPServer(address)
		if err != nil {
			return err
		}
		server := http.Server{
			Addr:    fmt.Sprintf(":%v", env.HTTPPort),
			Handler: httpServer,
		}

		certPath, keyPath, err := tls.GetCertPaths()
		if err != nil {
			log.Warnf("pfs-over-HTTP - TLS disabled: %v", err)
			return server.ListenAndServe()
		}

		cLoader := tls.NewCertLoader(certPath, keyPath, tls.CertCheckFrequency)
		err = cLoader.LoadAndStart()
		if err != nil {
			return errors.Wrapf(err, "couldn't load TLS cert for pfs-over-http: %v", err)
		}

		server.TLSConfig = &gotls.Config{GetCertificate: cLoader.GetCertificate}

		return server.ListenAndServeTLS(certPath, keyPath)
	})
	go waitForError("Githook Server", errChan, requireNoncriticalServers, func() error {
		return githook.RunGitHookServer(address, etcdAddress, path.Join(env.EtcdPrefix, env.PPSEtcdPrefix))
	})
	go waitForError("S3 Server", errChan, requireNoncriticalServers, func() error {
		server, err := s3.Server(env.S3GatewayPort, s3.NewMasterDriver(), func() (*client.APIClient, error) {
			return client.NewFromAddress(fmt.Sprintf("localhost:%d", env.PeerPort))
		})
		if err != nil {
			return err
		}
		certPath, keyPath, err := tls.GetCertPaths()
		if err != nil {
			log.Warnf("s3gateway TLS disabled: %v", err)
			return server.ListenAndServe()
		}
		cLoader := tls.NewCertLoader(certPath, keyPath, tls.CertCheckFrequency)
		// Read TLS cert and key
		err = cLoader.LoadAndStart()
		if err != nil {
			return errors.Wrapf(err, "couldn't load TLS cert for s3gateway: %v", err)
		}
		server.TLSConfig = &gotls.Config{GetCertificate: cLoader.GetCertificate}
		return server.ListenAndServeTLS(certPath, keyPath)
	})
	go waitForError("Prometheus Server", errChan, requireNoncriticalServers, func() error {
		http.Handle("/metrics", promhttp.Handler())
		return http.ListenAndServe(fmt.Sprintf(":%v", assets.PrometheusPort), nil)
	})
	return <-errChan
}

const clusterIDKey = "cluster-id"

func getClusterID(client *etcd.Client) (string, error) {
	resp, err := client.Get(context.Background(),
		clusterIDKey)

	// if it's a key not found error then we create the key
	if resp.Count == 0 {
		// This might error if it races with another pachd trying to set the
		// cluster id so we ignore the error.
		client.Put(context.Background(), clusterIDKey, uuid.NewWithoutDashes())
	} else if err != nil {
		return "", err
	} else {
		// We expect there to only be one value for this key
		id := string(resp.Kvs[0].Value)
		return id, nil
	}

	return getClusterID(client)
}

func logGRPCServerSetup(name string, f func() error) (retErr error) {
	log.Printf("started setting up %v GRPC Server", name)
	defer func() {
		if retErr != nil {
			retErr = errors.Wrapf(retErr, "error setting up %v GRPC Server", name)
		} else {
			log.Printf("finished setting up %v GRPC Server", name)
		}
	}()
	return f()
}

func waitForError(name string, errChan chan error, required bool, f func() error) {
	if err := f(); !errors.Is(err, http.ErrServerClosed) {
		if !required {
			log.Errorf("error setting up and/or running %v: %v", name, err)
		} else {
			errChan <- errors.Wrapf(err, "error setting up and/or running %v (use --require-critical-servers-only deploy flag to ignore errors from noncritical servers)", name)
		}
	}
}
