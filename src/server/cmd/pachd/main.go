package main

import (
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"

	"github.com/pachyderm/pachyderm/src/client"
	healthclient "github.com/pachyderm/pachyderm/src/client/health"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/discovery"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pkg/shard"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/server/health"
	pfs_driver "github.com/pachyderm/pachyderm/src/server/pfs/drive"
	pfs_server "github.com/pachyderm/pachyderm/src/server/pfs/server"
	cache_pb "github.com/pachyderm/pachyderm/src/server/pkg/cache/groupcachepb"
	cache_server "github.com/pachyderm/pachyderm/src/server/pkg/cache/server"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/metrics"
	"github.com/pachyderm/pachyderm/src/server/pkg/netutil"
	ppsserver "github.com/pachyderm/pachyderm/src/server/pps"
	pps_server "github.com/pachyderm/pachyderm/src/server/pps/server"

	flag "github.com/spf13/pflag"
	"go.pedge.io/lion"
	"go.pedge.io/lion/proto"
	"google.golang.org/grpc"
	"k8s.io/kubernetes/pkg/api"
	kube_client "k8s.io/kubernetes/pkg/client/restclient"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
)

var readinessCheck bool
var migrate string

func init() {
	flag.BoolVar(&readinessCheck, "readiness-check", false, "Set to true when checking if local pod is ready")
	flag.StringVar(&migrate, "migrate", "", "Use the format FROM_VERSION-TO_VERSION; e.g. 1.2.4-1.3.0")
	flag.Parse()
}

type appEnv struct {
	Port                  uint16 `env:"PORT,default=650"`
	NumShards             uint64 `env:"NUM_SHARDS,default=32"`
	StorageRoot           string `env:"PACH_ROOT,default=/pach"`
	StorageBackend        string `env:"STORAGE_BACKEND,default="`
	PPSEtcdPrefix         string `env:"PPS_ETCD_PREFIX,default=pachyderm_pps"`
	PFSEtcdPrefix         string `env:"PFS_ETCD_PREFIX,default=pachyderm_pfs"`
	KubeAddress           string `env:"KUBERNETES_PORT_443_TCP_ADDR,required"`
	EtcdAddress           string `env:"ETCD_PORT_2379_TCP_ADDR,required"`
	Namespace             string `env:"NAMESPACE,default=default"`
	Metrics               bool   `env:"METRICS,default=true"`
	Init                  bool   `env:"INIT,default=false"`
	BlockCacheBytes       int64  `env:"BLOCK_CACHE_BYTES,default=5368709120"` //default = 1 gigabyte
	WorkerShimImage       string `env:"WORKER_SHIM_IMAGE,default="`
	WorkerImagePullPolicy string `env:"WORKER_IMAGE_PULL_POLICY,default="`
	LogLevel              string `env:"LOG_LEVEL,default=info"`
}

func main() {
	cmdutil.Main(do, &appEnv{})
}

func do(appEnvObj interface{}) error {
	go func() {
		lion.Println(http.ListenAndServe(":651", nil))
	}()
	appEnv := appEnvObj.(*appEnv)
	switch appEnv.LogLevel {
	case "debug":
		lion.SetLevel(lion.LevelDebug)
	case "info":
		lion.SetLevel(lion.LevelInfo)
	case "error":
		lion.SetLevel(lion.LevelError)
	default:
		lion.Errorf("Unrecognized log level %s, falling back to default of \"info\"", appEnv.LogLevel)
		lion.SetLevel(lion.LevelInfo)
	}
	etcdAddress := fmt.Sprintf("http://%s:2379", appEnv.EtcdAddress)
	etcdClient := getEtcdClient(etcdAddress)
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

	clusterID, err := getClusterID(etcdClient)
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
		etcdClient,
		appEnv.NumShards,
		appEnv.Namespace,
	)
	go func() {
		if err := sharder.AssignRoles(address, nil); err != nil {
			protolion.Printf("error from sharder.AssignRoles: %s", sanitizeErr(err))
		}
	}()
	driver, err := pfs_driver.NewDriver(address, []string{etcdAddress}, appEnv.PFSEtcdPrefix)
	//	driver, err := drive.NewDriver(address)
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
	pfsAPIServer := pfs_server.NewAPIServer(driver, reporter)
	ppsAPIServer, err := pps_server.NewAPIServer(
		etcdAddress,
		appEnv.PPSEtcdPrefix,
		ppsserver.NewHasher(appEnv.NumShards, appEnv.NumShards),
		address,
		kubeClient,
		getNamespace(),
		appEnv.WorkerShimImage,
		appEnv.WorkerImagePullPolicy,
		reporter,
	)
	if err != nil {
		return err
	}
	go func() {
		if err := sharder.RegisterFrontends(nil, address, []shard.Frontend{cacheServer}); err != nil {
			protolion.Printf("error from sharder.RegisterFrontend %s", sanitizeErr(err))
		}
	}()
	go func() {
		if err := sharder.Register(nil, address, []shard.Server{ppsAPIServer, cacheServer}); err != nil {
			protolion.Printf("error from sharder.Register %s", sanitizeErr(err))
		}
	}()
	blockAPIServer, err := pfs_server.NewBlockAPIServer(appEnv.StorageRoot, appEnv.BlockCacheBytes, appEnv.StorageBackend)
	if err != nil {
		return err
	}
	healthServer := health.NewHealthServer()
	fmt.Println("pachd is ready to serve!")
	return grpcutil.Serve(
		func(s *grpc.Server) {
			pfsclient.RegisterAPIServer(s, pfsAPIServer)
			pfsclient.RegisterBlockAPIServer(s, blockAPIServer)
			ppsclient.RegisterAPIServer(s, ppsAPIServer)
			cache_pb.RegisterGroupCacheServer(s, cacheServer)
			healthclient.RegisterHealthServer(s, healthServer)
		},
		grpcutil.ServeOptions{
			Version:    version.Version,
			MaxMsgSize: client.MaxMsgSize,
		},
		grpcutil.ServeEnv{
			GRPCPort: appEnv.Port,
		},
	)
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

func getKubeClient(env *appEnv) (*kube.Client, error) {
	kubeClient, err := kube.NewInCluster()
	if err != nil {
		protolion.Errorf("falling back to insecure kube client due to error from NewInCluster: %s", sanitizeErr(err))
	} else {
		return kubeClient, err
	}
	config := &kube_client.Config{
		Host:     fmt.Sprintf("%s:443", env.KubeAddress),
		Insecure: true,
	}
	return kube.New(config)
}

// getNamespace returns the kubernetes namespace that this pachd pod runs in
func getNamespace() string {
	namespace := os.Getenv("PACHD_POD_NAMESPACE")
	if namespace != "" {
		return namespace
	}
	return api.NamespaceDefault
}

func sanitizeErr(err error) error {
	if err == nil {
		return nil
	}

	return errors.New(grpc.ErrorDesc(err))
}
