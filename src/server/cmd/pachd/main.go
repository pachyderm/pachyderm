package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
	units "github.com/docker/go-units"
	"github.com/pachyderm/pachyderm/src/client"
	adminclient "github.com/pachyderm/pachyderm/src/client/admin"
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
	adminserver "github.com/pachyderm/pachyderm/src/server/admin/server"
	authserver "github.com/pachyderm/pachyderm/src/server/auth/server"
	deployserver "github.com/pachyderm/pachyderm/src/server/deploy"
	eprsserver "github.com/pachyderm/pachyderm/src/server/enterprise/server"
	"github.com/pachyderm/pachyderm/src/server/health"
	pfs_server "github.com/pachyderm/pachyderm/src/server/pfs/server"
	cache_pb "github.com/pachyderm/pachyderm/src/server/pkg/cache/groupcachepb"
	cache_server "github.com/pachyderm/pachyderm/src/server/pkg/cache/server"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/metrics"
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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var mode string
var readiness bool

func init() {
	flag.StringVar(&mode, "mode", "full", "Pachd currently supports two modes: full and sidecar.  The former includes everything you need in a full pachd node.  The later runs only PFS, the Auth service, and a stripped-down version of PPS.")
	flag.BoolVar(&readiness, "readiness", false, "Run readiness check.")
	flag.Parse()
}

type appEnv struct {
	Port                  uint16 `env:"PORT,default=650"`
	NumShards             uint64 `env:"NUM_SHARDS,default=32"`
	StorageRoot           string `env:"PACH_ROOT,default=/pach"`
	StorageBackend        string `env:"STORAGE_BACKEND,default="`
	StorageHostPath       string `env:"STORAGE_HOST_PATH,default="`
	EtcdPrefix            string `env:"ETCD_PREFIX,default=pachyderm/1.7.0"`
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
	pachClient, err := client.NewFromAddress(fmt.Sprintf("%s:%d", address, appEnv.Port))
	if err != nil {
		return err
	}
	return pachClient.Health()
}

func doSidecarMode(appEnvObj interface{}) error {
	debug.SetGCPercent(50)
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
	pfsAPIServer, err := pfs_server.NewAPIServer(address, []string{etcdAddress}, path.Join(appEnv.EtcdPrefix, appEnv.PFSEtcdPrefix), int64(pfsCacheSize))
	if err != nil {
		return err
	}
	ppsAPIServer, err := pps_server.NewSidecarAPIServer(
		etcdAddress,
		path.Join(appEnv.EtcdPrefix, appEnv.PPSEtcdPrefix),
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
	authAPIServer, err := authserver.NewAuthServer(address, etcdAddress, path.Join(appEnv.EtcdPrefix, appEnv.AuthEtcdPrefix))
	if err != nil {
		return err
	}
	enterpriseAPIServer, err := eprsserver.NewEnterpriseServer(etcdAddress, path.Join(appEnv.EtcdPrefix, appEnv.EnterpriseEtcdPrefix))
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
	pfsAPIServer, err := pfs_server.NewAPIServer(address, []string{etcdAddress}, path.Join(appEnv.EtcdPrefix, appEnv.PFSEtcdPrefix), int64(pfsCacheSize))
	if err != nil {
		return err
	}
	kubeNamespace := getNamespace()
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

	authAPIServer, err := authserver.NewAuthServer(address, etcdAddress, path.Join(appEnv.EtcdPrefix, appEnv.AuthEtcdPrefix))
	if err != nil {
		return err
	}
	// Get the PPS token and put it in etcd.
	// This might emit an error, if the token has already been created or auth
	// has already been activated (in which case the token should also have
	// already been created). But in either  case we want to ignore the error and
	// re-use the existing token
	tokenResp, err := authAPIServer.GetAuthToken(context.Background(), &authclient.GetAuthTokenRequest{})
	if err == nil {
		_, err := etcdClientV3.Put(context.Background(),
			path.Join(appEnv.EtcdPrefix, appEnv.PPSEtcdPrefix, ppsconsts.PPSTokenKey),
			tokenResp.Token)
		if err != nil {
			return err
		}
	}
	enterpriseAPIServer, err := eprsserver.NewEnterpriseServer(etcdAddress, path.Join(appEnv.EtcdPrefix, appEnv.EnterpriseEtcdPrefix))
	if err != nil {
		return err
	}
	adminAPIServer := adminserver.NewAPIServer(address)

	healthServer := health.NewHealthServer()

	deployServer := deployserver.NewDeployServer(kubeClient, kubeNamespace)

	httpServer, err := pfs_server.NewHTTPServer(address, []string{etcdAddress}, path.Join(appEnv.EtcdPrefix, appEnv.PFSEtcdPrefix), blockCacheBytes)
	if err != nil {
		return err
	}
	var eg errgroup.Group
	eg.Go(func() error {
		return http.ListenAndServe(fmt.Sprintf(":%v", pfs_server.HTTPPort), httpServer)
	})
	eg.Go(func() error {
		return githook.RunGitHookServer(address, etcdAddress, path.Join(appEnv.EtcdPrefix, appEnv.PPSEtcdPrefix))
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
				adminclient.RegisterAPIServer(s, adminAPIServer)
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
	if err := migrate(address, kubeClient); err != nil {
		return err
	}
	healthServer.Ready()
	return eg.Wait()
}

func migrate(address string, kubeClient *kube.Clientset) error {
	endpoints, err := kubeClient.CoreV1().Endpoints(getNamespace()).Get("pachd", metav1.GetOptions{})
	if err != nil {
		return err
	}
	if len(endpoints.Subsets) != 1 {
		return fmt.Errorf("unexpected number of subsets: %d", len(endpoints.Subsets))
	}
	if len(endpoints.Subsets[0].Addresses) == 0 {
		// bail if there's no ready addresses
		return nil
	}
	externalPachClient, err := client.NewInCluster()
	if err != nil {
		return err
	}
	internalPachClient, err := client.NewFromAddress(address)
	if err != nil {
		return err
	}
	return internalPachClient.RestoreFrom(false, externalPachClient)
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
