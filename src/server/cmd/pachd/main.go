package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path"
	"runtime/debug"
	"runtime/pprof"
	"strconv"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
	units "github.com/docker/go-units"
	"github.com/pachyderm/pachyderm/src/client"
	adminclient "github.com/pachyderm/pachyderm/src/client/admin"
	authclient "github.com/pachyderm/pachyderm/src/client/auth"
	debugclient "github.com/pachyderm/pachyderm/src/client/debug"
	deployclient "github.com/pachyderm/pachyderm/src/client/deploy"
	eprsclient "github.com/pachyderm/pachyderm/src/client/enterprise"
	healthclient "github.com/pachyderm/pachyderm/src/client/health"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/discovery"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pkg/shard"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/client/version/versionpb"
	adminserver "github.com/pachyderm/pachyderm/src/server/admin/server"
	authserver "github.com/pachyderm/pachyderm/src/server/auth/server"
	debugserver "github.com/pachyderm/pachyderm/src/server/debug/server"
	deployserver "github.com/pachyderm/pachyderm/src/server/deploy"
	eprsserver "github.com/pachyderm/pachyderm/src/server/enterprise/server"
	"github.com/pachyderm/pachyderm/src/server/health"
	pach_http "github.com/pachyderm/pachyderm/src/server/http"
	pfs_server "github.com/pachyderm/pachyderm/src/server/pfs/server"
	cache_pb "github.com/pachyderm/pachyderm/src/server/pkg/cache/groupcachepb"
	cache_server "github.com/pachyderm/pachyderm/src/server/pkg/cache/server"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	"github.com/pachyderm/pachyderm/src/server/pkg/metrics"
	"github.com/pachyderm/pachyderm/src/server/pkg/netutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	pps_server "github.com/pachyderm/pachyderm/src/server/pps/server"
	"github.com/pachyderm/pachyderm/src/server/pps/server/githook"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	kube "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	defaultTreeCacheSize = 8
)

var mode string
var readiness bool

func init() {
	flag.StringVar(&mode, "mode", "full", "Pachd currently supports two modes: full and sidecar.  The former includes everything you need in a full pachd node.  The later runs only PFS, the Auth service, and a stripped-down version of PPS.")
	flag.BoolVar(&readiness, "readiness", false, "Run readiness check.")
	flag.Parse()
}

type appEnv struct {
	// Ports served by Pachd
	Port      uint16 `env:"PORT,default=650"`
	PProfPort uint16 `env:"PPROF_PORT,default=651"`
	HTTPPort  uint16 `env:"HTTP_PORT,default=652"`
	PeerPort  uint16 `env:"PEER_PORT,default=653"`

	NumShards             uint64 `env:"NUM_SHARDS,default=32"`
	StorageRoot           string `env:"PACH_ROOT,default=/pach"`
	StorageBackend        string `env:"STORAGE_BACKEND,default="`
	StorageHostPath       string `env:"STORAGE_HOST_PATH,default="`
	EtcdPrefix            string `env:"ETCD_PREFIX,default="`
	PPSEtcdPrefix         string `env:"PPS_ETCD_PREFIX,default=pachyderm_pps"`
	PFSEtcdPrefix         string `env:"PFS_ETCD_PREFIX,default=pachyderm_pfs"`
	AuthEtcdPrefix        string `env:"PACHYDERM_AUTH_ETCD_PREFIX,default=pachyderm_auth"`
	EnterpriseEtcdPrefix  string `env:"PACHYDERM_ENTERPRISE_ETCD_PREFIX,default=pachyderm_enterprise"`
	KubeAddress           string `env:"KUBERNETES_PORT_443_TCP_ADDR,required"`
	EtcdAddress           string `env:"ETCD_PORT_2379_TCP_ADDR,required"`
	Namespace             string `env:"NAMESPACE,default=default"`
	Metrics               bool   `env:"METRICS,default=true"`
	Init                  bool   `env:"INIT,default=false"`
	BlockCacheBytes       string `env:"BLOCK_CACHE_BYTES,default=1G"`
	PFSCacheSize          string `env:"PFS_CACHE_SIZE,default=0"`
	WorkerImage           string `env:"WORKER_IMAGE,default="`
	WorkerSidecarImage    string `env:"WORKER_SIDECAR_IMAGE,default="`
	WorkerImagePullPolicy string `env:"WORKER_IMAGE_PULL_POLICY,default="`
	LogLevel              string `env:"LOG_LEVEL,default=info"`
	IAMRole               string `env:"IAM_ROLE,default="`
	ImagePullSecret       string `env:"IMAGE_PULL_SECRET,default="`
	NoExposeDockerSocket  bool   `env:"NO_EXPOSE_DOCKER_SOCKET,default=false"`
	ExposeObjectAPI       bool   `env:"EXPOSE_OBJECT_API,default=false"`
	MemoryRequest         string `env:"PACHD_MEMORY_REQUEST,default=1T"`
}

func main() {
	switch {
	case readiness:
		cmdutil.Main(doReadinessCheck, &appEnv{})
	case mode == "full":
		cmdutil.Main(doFullMode, &appEnv{})
	case mode == "sidecar":
		cmdutil.Main(doSidecarMode, &appEnv{})
	default:
		fmt.Printf("unrecognized mode: %s\n", mode)
	}
}

func doReadinessCheck(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	address, err := netutil.ExternalIP()
	if err != nil {
		return err
	}
	pachClient, err := client.NewFromAddress(fmt.Sprintf("%s:%d", address, appEnv.PeerPort))
	if err != nil {
		return err
	}
	return pachClient.Health()
}

func doSidecarMode(appEnvObj interface{}) (retErr error) {
	defer func() {
		if retErr != nil {
			pprof.Lookup("goroutine").WriteTo(os.Stderr, 2)
		}
	}()
	appEnv := appEnvObj.(*appEnv)
	debug.SetGCPercent(50)
	go func() {
		log.Println(http.ListenAndServe(fmt.Sprintf(":%d", appEnv.PProfPort), nil))
	}()
	switch appEnv.LogLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.Errorf("Unrecognized log level %s, falling back to default of \"info\"", appEnv.LogLevel)
		log.SetLevel(log.InfoLevel)
	}
	if appEnv.EtcdPrefix == "" {
		appEnv.EtcdPrefix = col.DefaultPrefix
	}

	etcdAddress := fmt.Sprintf("http://%s:2379", appEnv.EtcdAddress)
	etcdClientV3, err := etcd.New(etcd.Config{
		Endpoints:   []string{etcdAddress},
		DialOptions: client.DefaultDialOptions(),
	})
	if err != nil {
		return err
	}

	clusterID, err := getClusterID(etcdClientV3)
	if err != nil {
		return fmt.Errorf("getClusterID: %v", err)
	}
	kubeClient, err := getKubeClient(appEnv)
	if err != nil {
		return fmt.Errorf("getKubeClient: %v", err)
	}
	var reporter *metrics.Reporter
	if appEnv.Metrics {
		reporter = metrics.NewReporter(clusterID, kubeClient)
	}
	address, err := netutil.ExternalIP()
	if err != nil {
		return fmt.Errorf("ExternalIP: %v", err)
	}
	address = fmt.Sprintf("%s:%d", address, appEnv.PeerPort)

	pfsCacheSize, err := strconv.Atoi(appEnv.PFSCacheSize)
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
	// The sidecar only needs to serve traffic on the peer port, as it only serves
	// traffic from the user container (the worker binary and occasionally user
	// pipelines)
	return grpcutil.Serve(
		grpcutil.ServerOptions{
			Port:       appEnv.PeerPort,
			MaxMsgSize: grpcutil.MaxMsgSize,
			RegisterFunc: func(s *grpc.Server) error {
				blockCacheBytes, err := units.RAMInBytes(appEnv.BlockCacheBytes)
				if err != nil {
					return fmt.Errorf("units.RAMInBytes: %v", err)
				}
				blockAPIServer, err := pfs_server.NewBlockAPIServer(appEnv.StorageRoot, blockCacheBytes, appEnv.StorageBackend, etcdAddress)
				if err != nil {
					return fmt.Errorf("pfs.NewBlockAPIServer: %v", err)
				}
				pfsclient.RegisterObjectAPIServer(s, blockAPIServer)

				memoryRequestBytes, err := units.RAMInBytes(appEnv.MemoryRequest)
				if err != nil {
					return err
				}
				pfsAPIServer, err := pfs_server.NewAPIServer(address, []string{etcdAddress}, path.Join(appEnv.EtcdPrefix, appEnv.PFSEtcdPrefix), treeCache, appEnv.StorageRoot, memoryRequestBytes)
				if err != nil {
					return fmt.Errorf("pfs.NewAPIServer: %v", err)
				}
				pfsclient.RegisterAPIServer(s, pfsAPIServer)

				ppsAPIServer, err := pps_server.NewSidecarAPIServer(
					etcdAddress,
					path.Join(appEnv.EtcdPrefix, appEnv.PPSEtcdPrefix),
					address,
					appEnv.IAMRole,
					reporter,
				)
				if err != nil {
					return fmt.Errorf("pps.NewSidecarAPIServer: %v", err)
				}
				ppsclient.RegisterAPIServer(s, ppsAPIServer)

				authAPIServer, err := authserver.NewAuthServer(
					address, etcdAddress, path.Join(appEnv.EtcdPrefix, appEnv.AuthEtcdPrefix),
					false)
				if err != nil {
					return fmt.Errorf("NewAuthServer: %v", err)
				}
				authclient.RegisterAPIServer(s, authAPIServer)

				enterpriseAPIServer, err := eprsserver.NewEnterpriseServer(
					address, etcdAddress, path.Join(appEnv.EtcdPrefix, appEnv.EnterpriseEtcdPrefix))
				if err != nil {
					return fmt.Errorf("NewEnterpriseServer: %v", err)
				}
				eprsclient.RegisterAPIServer(s, enterpriseAPIServer)

				healthclient.RegisterHealthServer(s, health.NewHealthServer())
				debugclient.RegisterDebugServer(s, debugserver.NewDebugServer(
					"", // no name for pachd servers
					etcdClientV3,
					path.Join(appEnv.EtcdPrefix, appEnv.PPSEtcdPrefix),
				))
				return nil
			},
		},
	)
}

func doFullMode(appEnvObj interface{}) (retErr error) {
	defer func() {
		if retErr != nil {
			pprof.Lookup("goroutine").WriteTo(os.Stderr, 2)
		}
	}()
	appEnv := appEnvObj.(*appEnv)
	go func() {
		log.Println(http.ListenAndServe(fmt.Sprintf(":%d", appEnv.PProfPort), nil))
	}()
	switch appEnv.LogLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.Errorf("Unrecognized log level %s, falling back to default of \"info\"", appEnv.LogLevel)
		log.SetLevel(log.InfoLevel)
	}

	if appEnv.EtcdPrefix == "" {
		appEnv.EtcdPrefix = col.DefaultPrefix
	}
	etcdAddress := fmt.Sprintf("http://%s:2379", appEnv.EtcdAddress)
	etcdClientV2 := getEtcdClient(etcdAddress)
	etcdClientV3, err := etcd.New(etcd.Config{
		Endpoints:   []string{etcdAddress},
		DialOptions: append(client.DefaultDialOptions(), grpc.WithTimeout(5*time.Minute)),
	})
	if err != nil {
		return err
	}
	clusterID, err := getClusterID(etcdClientV3)
	if err != nil {
		return fmt.Errorf("getClusterID: %v", err)
	}
	kubeClient, err := getKubeClient(appEnv)
	if err != nil {
		return fmt.Errorf("getKubeClient: %v", err)
	}
	var reporter *metrics.Reporter
	if appEnv.Metrics {
		reporter = metrics.NewReporter(clusterID, kubeClient)
	}
	address, err := netutil.ExternalIP()
	if err != nil {
		return fmt.Errorf("ExternalIP: %v", err)
	}
	address = fmt.Sprintf("%s:%d", address, appEnv.PeerPort)
	sharder := shard.NewSharder(
		etcdClientV2,
		appEnv.NumShards,
		appEnv.Namespace,
	)
	go func() {
		if err := sharder.AssignRoles(address, nil); err != nil {
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
	pfsCacheSize, err := strconv.Atoi(appEnv.PFSCacheSize)
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
		err = http.ListenAndServe(fmt.Sprintf(":%v", appEnv.HTTPPort), httpServer)
		if err != nil {
			log.Printf("error starting http server %v\n", err)
		}
		return fmt.Errorf("ListenAndServe: %v", err)
	})
	eg.Go(func() error {
		err := githook.RunGitHookServer(address, etcdAddress, path.Join(appEnv.EtcdPrefix, appEnv.PPSEtcdPrefix))
		if err != nil {
			log.Printf("error starting githook server %v\n", err)
		}
		return fmt.Errorf("RunGitHookServer: %v", err)
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
		err := grpcutil.Serve(
			grpcutil.ServerOptions{
				Port:                 appEnv.Port,
				MaxMsgSize:           grpcutil.MaxMsgSize,
				PublicPortTLSAllowed: true,
				RegisterFunc: func(s *grpc.Server) error {
					memoryRequestBytes, err := units.RAMInBytes(appEnv.MemoryRequest)
					if err != nil {
						return err
					}
					pfsAPIServer, err := pfs_server.NewAPIServer(address, []string{etcdAddress}, path.Join(appEnv.EtcdPrefix, appEnv.PFSEtcdPrefix), treeCache, appEnv.StorageRoot, memoryRequestBytes)
					if err != nil {
						return fmt.Errorf("pfs.NewAPIServer: %v", err)
					}
					pfsclient.RegisterAPIServer(s, pfsAPIServer)

					ppsAPIServer, err := pps_server.NewAPIServer(
						etcdAddress,
						path.Join(appEnv.EtcdPrefix, appEnv.PPSEtcdPrefix),
						address,
						kubeClient,
						kubeNamespace,
						appEnv.WorkerImage,
						appEnv.WorkerSidecarImage,
						appEnv.WorkerImagePullPolicy,
						appEnv.StorageRoot,
						appEnv.StorageBackend,
						appEnv.StorageHostPath,
						appEnv.IAMRole,
						appEnv.ImagePullSecret,
						appEnv.NoExposeDockerSocket,
						reporter,
					)
					if err != nil {
						return fmt.Errorf("pps.NewAPIServer: %v", err)
					}
					ppsclient.RegisterAPIServer(s, ppsAPIServer)

					if appEnv.ExposeObjectAPI {
						// Generally the object API should not be exposed publicly, but
						// TestGarbageCollection uses it and it may help with debugging
						blockAPIServer, err := pfs_server.NewBlockAPIServer(
							appEnv.StorageRoot,
							0 /* = blockCacheBytes (disable cache) */, appEnv.StorageBackend,
							etcdAddress)
						if err != nil {
							return fmt.Errorf("pfs.NewBlockAPIServer: %v", err)
						}
						pfsclient.RegisterObjectAPIServer(s, blockAPIServer)
					}

					authAPIServer, err := authserver.NewAuthServer(
						address, etcdAddress, path.Join(appEnv.EtcdPrefix, appEnv.AuthEtcdPrefix),
						true)
					if err != nil {
						return fmt.Errorf("NewAuthServer: %v", err)
					}
					authclient.RegisterAPIServer(s, authAPIServer)

					enterpriseAPIServer, err := eprsserver.NewEnterpriseServer(
						address, etcdAddress, path.Join(appEnv.EtcdPrefix, appEnv.EnterpriseEtcdPrefix))
					if err != nil {
						return fmt.Errorf("NewEnterpriseServer: %v", err)
					}
					eprsclient.RegisterAPIServer(s, enterpriseAPIServer)

					deployclient.RegisterAPIServer(s, deployserver.NewDeployServer(kubeClient, kubeNamespace))
					adminclient.RegisterAPIServer(s, adminserver.NewAPIServer(address, &adminclient.ClusterInfo{ID: clusterID}))
					healthclient.RegisterHealthServer(s, publicHealthServer)
					versionpb.RegisterAPIServer(s, version.NewAPIServer(version.Version, version.APIServerOptions{}))
					debugclient.RegisterDebugServer(s, debugserver.NewDebugServer(
						"", // no name for pachd servers
						etcdClientV3,
						path.Join(appEnv.EtcdPrefix, appEnv.PPSEtcdPrefix),
					))
					return nil
				},
			},
		)
		if err != nil {
			log.Printf("error starting grpc server %v\n", err)
		}
		return err
	})
	// Unfortunately, calling Register___Server(x) twice on the same
	// struct x doesn't work--x will only serve requests from the first
	// grpc.Server it was registered with. So we create a second set of
	// APIServer structs here so we can serve the Pachyderm API on the
	// peer port
	eg.Go(func() error {
		err := grpcutil.Serve(
			grpcutil.ServerOptions{
				Port:       appEnv.PeerPort,
				MaxMsgSize: grpcutil.MaxMsgSize,
				RegisterFunc: func(s *grpc.Server) error {
					cacheServer := cache_server.NewCacheServer(router, appEnv.NumShards)
					go func() {
						if err := sharder.RegisterFrontends(nil, address, []shard.Frontend{cacheServer}); err != nil {
							log.Printf("error from sharder.RegisterFrontend %s", grpcutil.ScrubGRPC(err))
						}
					}()
					go func() {
						if err := sharder.Register(nil, address, []shard.Server{cacheServer}); err != nil {
							log.Printf("error from sharder.Register %s", grpcutil.ScrubGRPC(err))
						}
					}()
					cache_pb.RegisterGroupCacheServer(s, cacheServer)

					blockCacheBytes, err := units.RAMInBytes(appEnv.BlockCacheBytes)
					if err != nil {
						return fmt.Errorf("units.RAMInBytes: %v", err)
					}
					blockAPIServer, err := pfs_server.NewBlockAPIServer(
						appEnv.StorageRoot, blockCacheBytes, appEnv.StorageBackend, etcdAddress)
					if err != nil {
						return fmt.Errorf("pfs.NewBlockAPIServer: %v", err)
					}
					pfsclient.RegisterObjectAPIServer(s, blockAPIServer)

					memoryRequestBytes, err := units.RAMInBytes(appEnv.MemoryRequest)
					if err != nil {
						return err
					}
					pfsAPIServer, err := pfs_server.NewAPIServer(
						address, []string{etcdAddress}, path.Join(appEnv.EtcdPrefix, appEnv.PFSEtcdPrefix), treeCache, appEnv.StorageRoot, memoryRequestBytes)
					if err != nil {
						return fmt.Errorf("pfs.NewAPIServer: %v", err)
					}
					pfsclient.RegisterAPIServer(s, pfsAPIServer)

					ppsAPIServer, err := pps_server.NewAPIServer(
						etcdAddress,
						path.Join(appEnv.EtcdPrefix, appEnv.PPSEtcdPrefix),
						address,
						kubeClient,
						kubeNamespace,
						appEnv.WorkerImage,
						appEnv.WorkerSidecarImage,
						appEnv.WorkerImagePullPolicy,
						appEnv.StorageRoot,
						appEnv.StorageBackend,
						appEnv.StorageHostPath,
						appEnv.IAMRole,
						appEnv.ImagePullSecret,
						appEnv.NoExposeDockerSocket,
						reporter,
					)
					if err != nil {
						return fmt.Errorf("pps.NewAPIServer: %v", err)
					}
					ppsclient.RegisterAPIServer(s, ppsAPIServer)

					authAPIServer, err := authserver.NewAuthServer(
						address, etcdAddress, path.Join(appEnv.EtcdPrefix, appEnv.AuthEtcdPrefix),
						false)
					if err != nil {
						return fmt.Errorf("NewAuthServer: %v", err)
					}
					authclient.RegisterAPIServer(s, authAPIServer)

					enterpriseAPIServer, err := eprsserver.NewEnterpriseServer(
						address, etcdAddress, path.Join(appEnv.EtcdPrefix, appEnv.EnterpriseEtcdPrefix))
					if err != nil {
						return fmt.Errorf("NewEnterpriseServer: %v", err)
					}
					eprsclient.RegisterAPIServer(s, enterpriseAPIServer)

					deployclient.RegisterAPIServer(s, deployserver.NewDeployServer(kubeClient, kubeNamespace))
					healthclient.RegisterHealthServer(s, peerHealthServer)
					versionpb.RegisterAPIServer(s, version.NewAPIServer(version.Version, version.APIServerOptions{}))
					adminclient.RegisterAPIServer(s, adminserver.NewAPIServer(address, &adminclient.ClusterInfo{ID: clusterID}))
					return nil
				},
			},
		)
		if err != nil {
			log.Printf("error starting grpc server %v\n", err)
		}
		return err
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

func getKubeClient(env *appEnv) (*kube.Clientset, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	kubeClient, err := kube.NewForConfig(cfg)
	if err != nil {
		log.Errorf("falling back to insecure kube client due to error from NewInCluster: %s", grpcutil.ScrubGRPC(err))
	} else {
		return kubeClient, err
	}
	config := &rest.Config{
		Host: fmt.Sprintf("%s:443", env.KubeAddress),
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}
	return kube.NewForConfig(config)
}

// getNamespace returns the kubernetes namespace that this pachd pod runs in
func getNamespace() string {
	namespace := os.Getenv("PACHD_POD_NAMESPACE")
	if namespace != "" {
		return namespace
	}
	return v1.NamespaceDefault
}
