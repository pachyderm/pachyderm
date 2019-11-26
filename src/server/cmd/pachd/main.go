package main

import (
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path"
	"runtime/debug"
	"runtime/pprof"
	"strconv"

	etcd "github.com/coreos/etcd/clientv3"
	units "github.com/docker/go-units"
	adminclient "github.com/pachyderm/pachyderm/src/client/admin"
	authclient "github.com/pachyderm/pachyderm/src/client/auth"
	debugclient "github.com/pachyderm/pachyderm/src/client/debug"
	eprsclient "github.com/pachyderm/pachyderm/src/client/enterprise"
	healthclient "github.com/pachyderm/pachyderm/src/client/health"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/discovery"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pkg/shard"
	"github.com/pachyderm/pachyderm/src/client/pkg/tracing"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	transactionclient "github.com/pachyderm/pachyderm/src/client/transaction"
	"github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/client/version/versionpb"
	adminserver "github.com/pachyderm/pachyderm/src/server/admin/server"
	authserver "github.com/pachyderm/pachyderm/src/server/auth/server"
	debugserver "github.com/pachyderm/pachyderm/src/server/debug/server"
	eprsserver "github.com/pachyderm/pachyderm/src/server/enterprise/server"
	"github.com/pachyderm/pachyderm/src/server/health"
	pach_http "github.com/pachyderm/pachyderm/src/server/http"
	"github.com/pachyderm/pachyderm/src/server/pfs/s3"
	pfs_server "github.com/pachyderm/pachyderm/src/server/pfs/server"
	cache_pb "github.com/pachyderm/pachyderm/src/server/pkg/cache/groupcachepb"
	cache_server "github.com/pachyderm/pachyderm/src/server/pkg/cache/server"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	logutil "github.com/pachyderm/pachyderm/src/server/pkg/log"
	"github.com/pachyderm/pachyderm/src/server/pkg/metrics"
	"github.com/pachyderm/pachyderm/src/server/pkg/netutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
	txnenv "github.com/pachyderm/pachyderm/src/server/pkg/transactionenv"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	pps_server "github.com/pachyderm/pachyderm/src/server/pps/server"
	"github.com/pachyderm/pachyderm/src/server/pps/server/githook"
	txnserver "github.com/pachyderm/pachyderm/src/server/transaction/server"

	"github.com/pachyderm/pachyderm/src/client/pkg/tls"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
)

const (
	defaultTreeCacheSize = 8
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
	// must run InstallJaegerTracer before InitWithKube (otherwise InitWithKube
	// may create a pach client before tracing is active, not install the Jaeger
	// gRPC interceptor in the client, and not propagate traces)
	if endpoint := tracing.InstallJaegerTracerFromEnv(); endpoint != "" {
		log.Printf("connecting to Jaeger at %q", endpoint)
	} else {
		log.Printf("no Jaeger collector found (JAEGER_COLLECTOR_SERVICE_HOST not set)")
	}
	env := serviceenv.InitWithKube(serviceenv.NewConfiguration(config))
	debug.SetGCPercent(50)
	go func() {
		log.Println(http.ListenAndServe(fmt.Sprintf(":%d", env.PProfPort), nil))
	}()
	switch env.LogLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.Errorf("Unrecognized log level %s, falling back to default of \"info\"", env.LogLevel)
		log.SetLevel(log.InfoLevel)
	}
	if env.EtcdPrefix == "" {
		env.EtcdPrefix = col.DefaultPrefix
	}
	clusterID, err := getClusterID(env.GetEtcdClient())
	if err != nil {
		return fmt.Errorf("getClusterID: %v", err)
	}
	var reporter *metrics.Reporter
	if env.Metrics {
		reporter = metrics.NewReporter(clusterID, env)
	}

	pfsCacheSize, err := strconv.Atoi(env.PFSCacheSize)
	if err != nil {
		return fmt.Errorf("Atoi: %v", err)
	}
	if pfsCacheSize == 0 {
		pfsCacheSize = defaultTreeCacheSize
	}
	treeCache, err := hashtree.NewCache(pfsCacheSize)
	if err != nil {
		return fmt.Errorf("lru.New: %v", err)
	}
	server, err := grpcutil.NewServer(context.Background(), false)
	if err != nil {
		return err
	}

	txnEnv := &txnenv.TransactionEnv{}
	blockCacheBytes, err := units.RAMInBytes(env.BlockCacheBytes)
	if err != nil {
		return fmt.Errorf("units.RAMInBytes: %v", err)
	}
	blockAPIServer, err := pfs_server.NewBlockAPIServer(env.StorageRoot, blockCacheBytes, env.StorageBackend, net.JoinHostPort(env.EtcdHost, env.EtcdPort), false)
	if err != nil {
		return fmt.Errorf("pfs.NewBlockAPIServer: %v", err)
	}
	pfsclient.RegisterObjectAPIServer(server.Server, blockAPIServer)

	memoryRequestBytes, err := units.RAMInBytes(env.MemoryRequest)
	if err != nil {
		return err
	}
	pfsAPIServer, err := pfs_server.NewAPIServer(
		env,
		txnEnv,
		path.Join(env.EtcdPrefix, env.PFSEtcdPrefix),
		treeCache,
		env.StorageRoot,
		memoryRequestBytes,
	)
	if err != nil {
		return fmt.Errorf("pfs.NewAPIServer: %v", err)
	}
	pfsclient.RegisterAPIServer(server.Server, pfsAPIServer)

	ppsAPIServer, err := pps_server.NewSidecarAPIServer(
		env,
		path.Join(env.EtcdPrefix, env.PPSEtcdPrefix),
		env.IAMRole,
		reporter,
		env.PPSWorkerPort,
		env.PProfPort,
		env.HTTPPort,
		env.PeerPort,
	)
	if err != nil {
		return fmt.Errorf("pps.NewSidecarAPIServer: %v", err)
	}
	ppsclient.RegisterAPIServer(server.Server, ppsAPIServer)

	authAPIServer, err := authserver.NewAuthServer(
		env,
		txnEnv,
		path.Join(env.EtcdPrefix, env.AuthEtcdPrefix),
		false,
	)
	if err != nil {
		return fmt.Errorf("NewAuthServer: %v", err)
	}
	authclient.RegisterAPIServer(server.Server, authAPIServer)

	transactionAPIServer, err := txnserver.NewAPIServer(
		env,
		txnEnv,
		path.Join(env.EtcdPrefix, env.PFSEtcdPrefix),
	)
	if err != nil {
		return fmt.Errorf("transaction.NewAPIServer: %v", err)
	}
	transactionclient.RegisterAPIServer(server.Server, transactionAPIServer)

	enterpriseAPIServer, err := eprsserver.NewEnterpriseServer(
		env, path.Join(env.EtcdPrefix, env.EnterpriseEtcdPrefix))
	if err != nil {
		return fmt.Errorf("NewEnterpriseServer: %v", err)
	}
	eprsclient.RegisterAPIServer(server.Server, enterpriseAPIServer)

	healthclient.RegisterHealthServer(server.Server, health.NewHealthServer())
	debugclient.RegisterDebugServer(server.Server, debugserver.NewDebugServer(
		"", // no name for pachd servers
		env.GetEtcdClient(),
		path.Join(env.EtcdPrefix, env.PPSEtcdPrefix),
		env.PPSWorkerPort,
	))
	txnEnv.Initialize(env, transactionAPIServer, authAPIServer, pfsAPIServer)

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
	// must run InstallJaegerTracer before InitWithKube
	if endpoint := tracing.InstallJaegerTracerFromEnv(); endpoint != "" {
		log.Printf("connecting to Jaeger at %q", endpoint)
	} else {
		log.Printf("no Jaeger collector found (JAEGER_COLLECTOR_SERVICE_HOST not set)")
	}
	env := serviceenv.InitWithKube(serviceenv.NewConfiguration(config))
	debug.SetGCPercent(50)
	go func() {
		log.Println(http.ListenAndServe(fmt.Sprintf(":%d", env.PProfPort), nil))
	}()
	switch env.LogLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.Errorf("Unrecognized log level %s, falling back to default of \"info\"", env.LogLevel)
		log.SetLevel(log.InfoLevel)
	}
	if env.EtcdPrefix == "" {
		env.EtcdPrefix = col.DefaultPrefix
	}
	clusterID, err := getClusterID(env.GetEtcdClient())
	if err != nil {
		return fmt.Errorf("getClusterID: %v", err)
	}
	var reporter *metrics.Reporter
	if env.Metrics {
		reporter = metrics.NewReporter(clusterID, env)
	}
	// (bryce) Do we have to use etcd client v2 here for sharder? Might want to re-visit this later.
	etcdAddress := fmt.Sprintf("http://%s", net.JoinHostPort(env.EtcdHost, env.EtcdPort))
	etcdClientV2 := getEtcdClient(etcdAddress)
	ip, err := netutil.ExternalIP()
	if err != nil {
		return fmt.Errorf("error getting pachd external ip: %v", err)
	}
	address := net.JoinHostPort(ip, fmt.Sprintf("%d", env.PeerPort))
	sharder := shard.NewSharder(
		etcdClientV2,
		env.NumShards,
		env.Namespace,
	)
	go func() {
		if err := sharder.AssignRoles(address); err != nil {
			log.Printf("error from sharder.AssignRoles: %s", grpcutil.ScrubGRPC(err))
		}
	}()
	router := shard.NewRouter(
		sharder,
		grpcutil.NewDialer(
			grpc.WithInsecure(),
		),
		address,
	)
	pfsCacheSize, err := strconv.Atoi(env.PFSCacheSize)
	if err != nil {
		return fmt.Errorf("Atoi: %v", err)
	}
	if pfsCacheSize == 0 {
		pfsCacheSize = defaultTreeCacheSize
	}
	treeCache, err := hashtree.NewCache(pfsCacheSize)
	if err != nil {
		return fmt.Errorf("lru.New: %v", err)
	}
	kubeNamespace := getNamespace()
	publicHealthServer := health.NewHealthServer()
	peerHealthServer := health.NewHealthServer()

	// TODO(msteffen): We should not use an errorgroup here. Errorgroup waits
	// until *all* goroutines have run and then returns, but we want pachd to halt
	// if *any* of these serve functions returns an arror.
	var eg errgroup.Group
	eg.Go(func() error {
		httpServer, err := pach_http.NewHTTPServer(address)
		if err != nil {
			return err
		}
		err = http.ListenAndServe(fmt.Sprintf(":%v", env.HTTPPort), httpServer)
		if err != nil {
			log.Printf("error starting http server %v\n", err)
		}
		return fmt.Errorf("ListenAndServe: %v", err)
	})
	eg.Go(func() error {
		err := githook.RunGitHookServer(address, etcdAddress, path.Join(env.EtcdPrefix, env.PPSEtcdPrefix))
		if err != nil {
			log.Printf("error starting githook server %v\n", err)
		}
		return fmt.Errorf("RunGitHookServer: %v", err)
	})
	eg.Go(func() error {
		server, err := s3.Server(env.S3GatewayPort, env.Port)
		if err != nil {
			return fmt.Errorf("s3gateway server: %v", err)
		}
		certPath, keyPath, err := tls.GetCertPaths()
		if err != nil {
			log.Warnf("s3gateway TLS disabled: %v", err)
			if err := server.ListenAndServe(); err != http.ErrServerClosed {
				return fmt.Errorf("s3gateway listen: %v", err)
			}
		} else {
			if err := server.ListenAndServeTLS(certPath, keyPath); err != http.ErrServerClosed {
				return fmt.Errorf("s3gateway listen: %v", err)
			}
		}
		return nil
	})
	eg.Go(func() error {
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(fmt.Sprintf(":%v", assets.PrometheusPort), nil)
		if err != nil {
			log.Printf("error starting prometheus server %v\n", err)
		}
		return fmt.Errorf("ListenAndServe: %v", err)
	})
	eg.Go(func() error {
		// start public pachd server
		server, err := grpcutil.NewServer(context.Background(), true)
		if err != nil {
			return err
		}

		txnEnv := &txnenv.TransactionEnv{}
		memoryRequestBytes, err := units.RAMInBytes(env.MemoryRequest)
		if err != nil {
			return err
		}
		pfsAPIServer, err := pfs_server.NewAPIServer(env, txnEnv, path.Join(env.EtcdPrefix, env.PFSEtcdPrefix), treeCache, env.StorageRoot, memoryRequestBytes)
		if err != nil {
			return fmt.Errorf("pfs.NewAPIServer: %v", err)
		}
		pfsclient.RegisterAPIServer(server.Server, pfsAPIServer)

		ppsAPIServer, err := pps_server.NewAPIServer(
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
			env.IAMRole,
			env.ImagePullSecret,
			env.NoExposeDockerSocket,
			reporter,
			env.WorkerUsesRoot,
			env.PPSWorkerPort,
			env.Port,
			env.PProfPort,
			env.HTTPPort,
			env.PeerPort,
		)
		if err != nil {
			return fmt.Errorf("pps.NewAPIServer: %v", err)
		}
		ppsclient.RegisterAPIServer(server.Server, ppsAPIServer)

		if env.ExposeObjectAPI {
			// Generally the object API should not be exposed publicly, but
			// TestGarbageCollection uses it and it may help with debugging
			blockAPIServer, err := pfs_server.NewBlockAPIServer(
				env.StorageRoot,
				0 /* = blockCacheBytes (disable cache) */, env.StorageBackend,
				etcdAddress,
				true /* duplicate */)
			if err != nil {
				return fmt.Errorf("pfs.NewBlockAPIServer: %v", err)
			}
			pfsclient.RegisterObjectAPIServer(server.Server, blockAPIServer)
		}

		authAPIServer, err := authserver.NewAuthServer(
			env, txnEnv, path.Join(env.EtcdPrefix, env.AuthEtcdPrefix), true)
		if err != nil {
			return fmt.Errorf("NewAuthServer: %v", err)
		}
		authclient.RegisterAPIServer(server.Server, authAPIServer)

		transactionAPIServer, err := txnserver.NewAPIServer(
			env,
			txnEnv,
			path.Join(env.EtcdPrefix, env.PFSEtcdPrefix),
		)
		if err != nil {
			return fmt.Errorf("transaction.NewAPIServer: %v", err)
		}
		transactionclient.RegisterAPIServer(server.Server, transactionAPIServer)

		enterpriseAPIServer, err := eprsserver.NewEnterpriseServer(
			env, path.Join(env.EtcdPrefix, env.EnterpriseEtcdPrefix))
		if err != nil {
			return fmt.Errorf("NewEnterpriseServer: %v", err)
		}
		eprsclient.RegisterAPIServer(server.Server, enterpriseAPIServer)

		adminclient.RegisterAPIServer(server.Server, adminserver.NewAPIServer(address, env.StorageRoot, &adminclient.ClusterInfo{ID: clusterID}))
		healthclient.RegisterHealthServer(server.Server, publicHealthServer)
		versionpb.RegisterAPIServer(server.Server, version.NewAPIServer(version.Version, version.APIServerOptions{}))
		debugclient.RegisterDebugServer(server.Server, debugserver.NewDebugServer(
			"", // no name for pachd servers
			env.GetEtcdClient(),
			path.Join(env.EtcdPrefix, env.PPSEtcdPrefix),
			env.PPSWorkerPort,
		))
		txnEnv.Initialize(env, transactionAPIServer, authAPIServer, pfsAPIServer)

		if _, err := server.ListenTCP("", env.Port); err != nil {
			return err
		}
		return server.Wait()
	})
	eg.Go(func() error {
		// start internal pachd server
		server, err := grpcutil.NewServer(context.Background(), false)
		if err != nil {
			return err
		}

		txnEnv := &txnenv.TransactionEnv{}
		cacheServer := cache_server.NewCacheServer(router, env.NumShards)
		go func() {
			if err := sharder.RegisterFrontends(address, []shard.Frontend{cacheServer}); err != nil {
				log.Printf("error from sharder.RegisterFrontend %s", grpcutil.ScrubGRPC(err))
			}
		}()
		go func() {
			if err := sharder.Register(address, []shard.Server{cacheServer}); err != nil {
				log.Printf("error from sharder.Register %s", grpcutil.ScrubGRPC(err))
			}
		}()
		cache_pb.RegisterGroupCacheServer(server.Server, cacheServer)

		blockCacheBytes, err := units.RAMInBytes(env.BlockCacheBytes)
		if err != nil {
			return fmt.Errorf("units.RAMInBytes: %v", err)
		}
		blockAPIServer, err := pfs_server.NewBlockAPIServer(
			env.StorageRoot, blockCacheBytes, env.StorageBackend, etcdAddress, false)
		if err != nil {
			return fmt.Errorf("pfs.NewBlockAPIServer: %v", err)
		}
		pfsclient.RegisterObjectAPIServer(server.Server, blockAPIServer)

		memoryRequestBytes, err := units.RAMInBytes(env.MemoryRequest)
		if err != nil {
			return err
		}
		pfsAPIServer, err := pfs_server.NewAPIServer(
			env,
			txnEnv,
			path.Join(env.EtcdPrefix, env.PFSEtcdPrefix),
			treeCache,
			env.StorageRoot,
			memoryRequestBytes,
		)
		if err != nil {
			return fmt.Errorf("pfs.NewAPIServer: %v", err)
		}
		pfsclient.RegisterAPIServer(server.Server, pfsAPIServer)

		ppsAPIServer, err := pps_server.NewAPIServer(
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
			env.IAMRole,
			env.ImagePullSecret,
			env.NoExposeDockerSocket,
			reporter,
			env.WorkerUsesRoot,
			env.PPSWorkerPort,
			env.Port,
			env.PProfPort,
			env.HTTPPort,
			env.PeerPort,
		)
		if err != nil {
			return fmt.Errorf("pps.NewAPIServer: %v", err)
		}
		ppsclient.RegisterAPIServer(server.Server, ppsAPIServer)

		authAPIServer, err := authserver.NewAuthServer(
			env,
			txnEnv,
			path.Join(env.EtcdPrefix, env.AuthEtcdPrefix),
			false,
		)
		if err != nil {
			return fmt.Errorf("NewAuthServer: %v", err)
		}
		authclient.RegisterAPIServer(server.Server, authAPIServer)

		transactionAPIServer, err := txnserver.NewAPIServer(
			env,
			txnEnv,
			path.Join(env.EtcdPrefix, env.PFSEtcdPrefix),
		)
		if err != nil {
			return fmt.Errorf("transaction.NewAPIServer: %v", err)
		}
		transactionclient.RegisterAPIServer(server.Server, transactionAPIServer)

		enterpriseAPIServer, err := eprsserver.NewEnterpriseServer(
			env, path.Join(env.EtcdPrefix, env.EnterpriseEtcdPrefix))
		if err != nil {
			return fmt.Errorf("NewEnterpriseServer: %v", err)
		}
		eprsclient.RegisterAPIServer(server.Server, enterpriseAPIServer)

		healthclient.RegisterHealthServer(server.Server, peerHealthServer)
		versionpb.RegisterAPIServer(server.Server, version.NewAPIServer(version.Version, version.APIServerOptions{}))
		adminclient.RegisterAPIServer(server.Server, adminserver.NewAPIServer(address, env.StorageRoot, &adminclient.ClusterInfo{ID: clusterID}))
		txnEnv.Initialize(env, transactionAPIServer, authAPIServer, pfsAPIServer)

		if _, err := server.ListenTCP("", env.PeerPort); err != nil {
			return err
		}
		return server.Wait()
	})

	// TODO(msteffen): Is it really necessary to indicate that the peer service is
	// healthy? Presumably migrate() will call the peer service no matter what.
	publicHealthServer.Ready()
	peerHealthServer.Ready()
	return eg.Wait()
}

func getEtcdClient(etcdAddress string) discovery.Client {
	return discovery.NewEtcdClient(etcdAddress)
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

// getNamespace returns the kubernetes namespace that this pachd pod runs in
func getNamespace() string {
	namespace := os.Getenv("PACHD_POD_NAMESPACE")
	if namespace != "" {
		return namespace
	}
	return v1.NamespaceDefault
}
