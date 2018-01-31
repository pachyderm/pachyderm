package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
	units "github.com/docker/go-units"
	"github.com/pachyderm/pachyderm/src/client"
	authclient "github.com/pachyderm/pachyderm/src/client/auth"
	deployclient "github.com/pachyderm/pachyderm/src/client/deploy"
	eprsclient "github.com/pachyderm/pachyderm/src/client/enterprise"
	healthclient "github.com/pachyderm/pachyderm/src/client/health"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/discovery"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pkg/shard"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/client/version"
	authserver "github.com/pachyderm/pachyderm/src/server/auth/server"
	deployserver "github.com/pachyderm/pachyderm/src/server/deploy"
	eprsserver "github.com/pachyderm/pachyderm/src/server/enterprise/server"
	"github.com/pachyderm/pachyderm/src/server/health"
	pfs_server "github.com/pachyderm/pachyderm/src/server/pfs/server"
	cache_pb "github.com/pachyderm/pachyderm/src/server/pkg/cache/groupcachepb"
	cache_server "github.com/pachyderm/pachyderm/src/server/pkg/cache/server"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/metrics"
	"github.com/pachyderm/pachyderm/src/server/pkg/migration"
	"github.com/pachyderm/pachyderm/src/server/pkg/netutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsconsts"
	pps_server "github.com/pachyderm/pachyderm/src/server/pps/server"
	"github.com/pachyderm/pachyderm/src/server/pps/server/githook"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	kube "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var mode string
var readinessCheck bool
var migrate string

func init() {
	flag.StringVar(&mode, "mode", "full", "Pachd currently supports two modes: full and sidecar.  The former includes everything you need in a full pachd node.  The later runs only PFS, the Auth service, and a stripped-down version of PPS.")
	flag.BoolVar(&readinessCheck, "readiness-check", false, "Set to true when checking if local pod is ready")
	flag.StringVar(&migrate, "migrate", "", "Use the format FROM_VERSION-TO_VERSION; e.g. 1.4.8-1.5.0")
	flag.Parse()
}

type appEnv struct {
	Port                  uint16 `env:"PORT,default=650"`
	NumShards             uint64 `env:"NUM_SHARDS,default=32"`
	StorageRoot           string `env:"PACH_ROOT,default=/pach"`
	StorageBackend        string `env:"STORAGE_BACKEND,default="`
	StorageHostPath       string `env:"STORAGE_HOST_PATH,default="`
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
}

func main() {
	switch mode {
	case "full":
		cmdutil.Main(doFullMode, &appEnv{})
	case "sidecar":
		cmdutil.Main(doSidecarMode, &appEnv{})
	default:
		fmt.Printf("unrecognized mode: %s\n", mode)
	}
}

func doSidecarMode(appEnvObj interface{}) error {
	go func() {
		log.Println(http.ListenAndServe(":651", nil))
	}()
	appEnv := appEnvObj.(*appEnv)
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

	etcdAddress := fmt.Sprintf("http://%s:2379", appEnv.EtcdAddress)
	etcdClientV2 := getEtcdClient(etcdAddress)

	clusterID, err := getClusterID(etcdClientV2)
	if err != nil {
		return err
	}
	kubeClient, err := getKubeClient(appEnv)
	if err != nil {
		return err
	}
	var reporter *metrics.Reporter
	if appEnv.Metrics {
		reporter = metrics.NewReporter(clusterID, kubeClient)
	}
	address, err := netutil.ExternalIP()
	if err != nil {
		return err
	}
	address = fmt.Sprintf("%s:%d", address, appEnv.Port)
	pfsCacheSize, err := strconv.Atoi(appEnv.PFSCacheSize)
	if err != nil {
		return err
	}
	pfsAPIServer, err := pfs_server.NewAPIServer(address, []string{etcdAddress}, appEnv.PFSEtcdPrefix, int64(pfsCacheSize))
	if err != nil {
		return err
	}
	ppsAPIServer, err := pps_server.NewSidecarAPIServer(
		etcdAddress,
		appEnv.PPSEtcdPrefix,
		address,
		appEnv.IAMRole,
		reporter,
	)
	if err != nil {
		return err
	}
	blockCacheBytes, err := units.RAMInBytes(appEnv.BlockCacheBytes)
	if err != nil {
		return err
	}
	blockAPIServer, err := pfs_server.NewBlockAPIServer(appEnv.StorageRoot, blockCacheBytes, appEnv.StorageBackend, etcdAddress)
	if err != nil {
		return err
	}
	healthServer := health.NewHealthServer()
	authAPIServer, err := authserver.NewAuthServer(address, etcdAddress, appEnv.AuthEtcdPrefix)
	if err != nil {
		return err
	}
	enterpriseAPIServer, err := eprsserver.NewEnterpriseServer(etcdAddress, appEnv.EnterpriseEtcdPrefix)
	if err != nil {
		return err
	}
	return grpcutil.Serve(
		func(s *grpc.Server) {
			pfsclient.RegisterAPIServer(s, pfsAPIServer)
			pfsclient.RegisterObjectAPIServer(s, blockAPIServer)
			ppsclient.RegisterAPIServer(s, ppsAPIServer)
			healthclient.RegisterHealthServer(s, healthServer)
			authclient.RegisterAPIServer(s, authAPIServer)
			eprsclient.RegisterAPIServer(s, enterpriseAPIServer)
		},
		grpcutil.ServeOptions{
			Version:    version.Version,
			MaxMsgSize: grpcutil.MaxMsgSize,
		},
		grpcutil.ServeEnv{
			GRPCPort: appEnv.Port,
		},
	)
}

func doFullMode(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	if migrate != "" {
		parts := strings.Split(migrate, "-")
		if len(parts) != 2 {
			return fmt.Errorf("the migration flag needs to be of the format FROM_VERSION-TO_VERSION; e.g. 1.4.8-1.5.0")
		}

		if err := migration.Run(appEnv.EtcdAddress, appEnv.PFSEtcdPrefix, appEnv.PPSEtcdPrefix, parts[0], parts[1]); err != nil {
			return fmt.Errorf("error from migration: %v", err)
		}
		return nil
	}

	if err := migration.OneSixToOneSeven(appEnv.EtcdAddress, appEnv.PFSEtcdPrefix, appEnv.PPSEtcdPrefix); err != nil {
		return err
	}

	go func() {
		log.Println(http.ListenAndServe(":651", nil))
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
	etcdAddress := fmt.Sprintf("http://%s:2379", appEnv.EtcdAddress)
	etcdClientV2 := getEtcdClient(etcdAddress)
	etcdClientV3, err := etcd.New(etcd.Config{
		Endpoints:   []string{etcdAddress},
		DialOptions: append(client.EtcdDialOptions(), grpc.WithTimeout(5*time.Minute)),
	})

	// Check if Pachd pods are ready
	if readinessCheck {
		c, err := client.NewFromAddress("127.0.0.1:650")
		if err != nil {
			return err
		}

		// We want to use a PPS API instead of a PFS API because PFS APIs
		// typically talk to every node, but the point of the readiness probe
		// is that it checks to see if this particular node is functioning,
		// and removing it from the service if it's not.  So if we use a PFS
		// API such as ListRepo for readiness probe, then the failure of any
		// node will result in the failures of all readiness probes, causing
		// all nodes to be removed from the pachd service.
		_, err = c.ListPipeline()
		if err != nil {
			return err
		}

		os.Exit(0)

		return nil
	}

	clusterID, err := getClusterID(etcdClientV2)
	if err != nil {
		return err
	}
	kubeClient, err := getKubeClient(appEnv)
	if err != nil {
		return err
	}
	var reporter *metrics.Reporter
	if appEnv.Metrics {
		reporter = metrics.NewReporter(clusterID, kubeClient)
	}
	address, err := netutil.ExternalIP()
	if err != nil {
		return err
	}
	address = fmt.Sprintf("%s:%d", address, appEnv.Port)
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
	pfsCacheSize, err := strconv.Atoi(appEnv.PFSCacheSize)
	if err != nil {
		return err
	}
	router := shard.NewRouter(
		sharder,
		grpcutil.NewDialer(
			grpc.WithInsecure(),
		),
		address,
	)
	cacheServer := cache_server.NewCacheServer(router, appEnv.NumShards)
	pfsAPIServer, err := pfs_server.NewAPIServer(address, []string{etcdAddress}, appEnv.PFSEtcdPrefix, int64(pfsCacheSize))
	if err != nil {
		return err
	}
	kubeNamespace := getNamespace()
	ppsAPIServer, err := pps_server.NewAPIServer(
		etcdAddress,
		appEnv.PPSEtcdPrefix,
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
		reporter,
	)
	if err != nil {
		return err
	}
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
	blockCacheBytes, err := units.RAMInBytes(appEnv.BlockCacheBytes)
	if err != nil {
		return err
	}
	blockAPIServer, err := pfs_server.NewBlockAPIServer(appEnv.StorageRoot, blockCacheBytes, appEnv.StorageBackend, etcdAddress)
	if err != nil {
		return err
	}

	authAPIServer, err := authserver.NewAuthServer(address, etcdAddress, appEnv.AuthEtcdPrefix)
	if err != nil {
		return err
	}
	// Get the PPS token and put it in etcd.
	// This might emit an error, if the token has already been created or auth
	// has already been activated (in which case the token should also have
	// already been created). But in either  case we want to ignore the error and
	// re-use the existing token
	capabilityResp, err := authAPIServer.GetCapability(context.Background(), &authclient.GetCapabilityRequest{})
	if err == nil {
		_, err := etcdClientV3.Put(context.Background(),
			path.Join(appEnv.PPSEtcdPrefix, ppsconsts.PPSTokenKey),
			capabilityResp.Capability)
		if err != nil {
			return err
		}
	}
	enterpriseAPIServer, err := eprsserver.NewEnterpriseServer(etcdAddress, appEnv.EnterpriseEtcdPrefix)
	if err != nil {
		return err
	}

	healthServer := health.NewHealthServer()

	deployServer := deployserver.NewDeployServer(kubeClient, kubeNamespace)

	httpServer, err := pfs_server.NewHTTPServer(address, []string{etcdAddress}, appEnv.PFSEtcdPrefix, blockCacheBytes)
	if err != nil {
		return err
	}
	var eg errgroup.Group
	eg.Go(func() error {
		return http.ListenAndServe(fmt.Sprintf(":%v", pfs_server.HTTPPort), httpServer)
	})
	eg.Go(func() error {
		return githook.RunGitHookServer(address, etcdAddress, appEnv.PPSEtcdPrefix)
	})
	eg.Go(func() error {
		return grpcutil.Serve(
			func(s *grpc.Server) {
				healthclient.RegisterHealthServer(s, healthServer)
				pfsclient.RegisterAPIServer(s, pfsAPIServer)
				pfsclient.RegisterObjectAPIServer(s, blockAPIServer)
				ppsclient.RegisterAPIServer(s, ppsAPIServer)
				cache_pb.RegisterGroupCacheServer(s, cacheServer)
				authclient.RegisterAPIServer(s, authAPIServer)
				eprsclient.RegisterAPIServer(s, enterpriseAPIServer)
				deployclient.RegisterAPIServer(s, deployServer)
			},
			grpcutil.ServeOptions{
				Version:    version.Version,
				MaxMsgSize: grpcutil.MaxMsgSize,
			},
			grpcutil.ServeEnv{
				GRPCPort: appEnv.Port,
			},
		)
	})
	return eg.Wait()
}

func getEtcdClient(etcdAddress string) discovery.Client {
	return discovery.NewEtcdClient(etcdAddress)
}

const clusterIDKey = "cluster-id"

func getClusterID(client discovery.Client) (string, error) {
	id, err := client.Get(clusterIDKey)
	// if it's a key not found error then we create the key
	if err != nil && strings.Contains(err.Error(), "not found") {
		// This might error if it races with another pachd trying to set the
		// cluster id so we ignore the error.
		client.Set(clusterIDKey, uuid.NewWithoutDashes(), 0)
	} else if err != nil {
		return "", err
	} else {
		return id, nil
	}
	return client.Get(clusterIDKey)
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
