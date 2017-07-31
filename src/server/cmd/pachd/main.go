package main

import (
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"

	units "github.com/docker/go-units"
	"github.com/pachyderm/pachyderm/src/client"
	authclient "github.com/pachyderm/pachyderm/src/client/auth"
	healthclient "github.com/pachyderm/pachyderm/src/client/health"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/discovery"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pkg/shard"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/client/version"
	authserver "github.com/pachyderm/pachyderm/src/server/auth/server"
	"github.com/pachyderm/pachyderm/src/server/health"
	pfs_server "github.com/pachyderm/pachyderm/src/server/pfs/server"
	cache_pb "github.com/pachyderm/pachyderm/src/server/pkg/cache/groupcachepb"
	cache_server "github.com/pachyderm/pachyderm/src/server/pkg/cache/server"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/metrics"
	"github.com/pachyderm/pachyderm/src/server/pkg/migration"
	"github.com/pachyderm/pachyderm/src/server/pkg/netutil"
	pps_server "github.com/pachyderm/pachyderm/src/server/pps/server"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"k8s.io/kubernetes/pkg/api"
	kube_client "k8s.io/kubernetes/pkg/client/restclient"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
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
	KubeAddress           string `env:"KUBERNETES_PORT_443_TCP_ADDR,required"`
	EtcdAddress           string `env:"ETCD_PORT_2379_TCP_ADDR,required"`
	Namespace             string `env:"NAMESPACE,default=default"`
	Metrics               bool   `env:"METRICS,default=true"`
	Init                  bool   `env:"INIT,default=false"`
	BlockCacheBytes       string `env:"BLOCK_CACHE_BYTES,default=1G"`
	PFSCacheBytes         string `env:"PFS_CACHE_BYTES,default=500M"`
	WorkerImage           string `env:"WORKER_IMAGE,default="`
	WorkerSidecarImage    string `env:"WORKER_SIDECAR_IMAGE,default="`
	WorkerImagePullPolicy string `env:"WORKER_IMAGE_PULL_POLICY,default="`
	LogLevel              string `env:"LOG_LEVEL,default=info"`
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
	etcdClient := getEtcdClient(etcdAddress)

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
	pfsCacheBytes, err := units.RAMInBytes(appEnv.PFSCacheBytes)
	if err != nil {
		return err
	}
	pfsAPIServer, err := pfs_server.NewAPIServer(address, []string{etcdAddress}, appEnv.PFSEtcdPrefix, pfsCacheBytes)
	if err != nil {
		return err
	}
	ppsAPIServer, err := pps_server.NewSidecarAPIServer(
		etcdAddress,
		appEnv.PPSEtcdPrefix,
		address,
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
	authAPIServer, err := authserver.NewAuthServer(etcdAddress, appEnv.PFSEtcdPrefix)
	return grpcutil.Serve(
		func(s *grpc.Server) {
			pfsclient.RegisterAPIServer(s, pfsAPIServer)
			pfsclient.RegisterObjectAPIServer(s, blockAPIServer)
			ppsclient.RegisterAPIServer(s, ppsAPIServer)
			healthclient.RegisterHealthServer(s, healthServer)
			authclient.RegisterAPIServer(s, authAPIServer)
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
			log.Printf("error from sharder.AssignRoles: %s", sanitizeErr(err))
		}
	}()
	pfsCacheBytes, err := units.RAMInBytes(appEnv.PFSCacheBytes)
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
	pfsAPIServer, err := pfs_server.NewAPIServer(address, []string{etcdAddress}, appEnv.PFSEtcdPrefix, pfsCacheBytes)
	if err != nil {
		return err
	}
	ppsAPIServer, err := pps_server.NewAPIServer(
		etcdAddress,
		appEnv.PPSEtcdPrefix,
		address,
		kubeClient,
		getNamespace(),
		appEnv.WorkerImage,
		appEnv.WorkerSidecarImage,
		appEnv.WorkerImagePullPolicy,
		appEnv.StorageRoot,
		appEnv.StorageBackend,
		appEnv.StorageHostPath,
		reporter,
	)
	if err != nil {
		return err
	}
	go func() {
		if err := sharder.RegisterFrontends(nil, address, []shard.Frontend{cacheServer}); err != nil {
			log.Printf("error from sharder.RegisterFrontend %s", sanitizeErr(err))
		}
	}()
	go func() {
		if err := sharder.Register(nil, address, []shard.Server{cacheServer}); err != nil {
			log.Printf("error from sharder.Register %s", sanitizeErr(err))
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

	authAPIServer, err := authserver.NewAuthServer(etcdAddress, appEnv.PFSEtcdPrefix)
	if err != nil {
		return err
	}

	healthServer := health.NewHealthServer()

	httpServer, err := pfs_server.NewHTTPServer(address, []string{etcdAddress}, appEnv.PFSEtcdPrefix, blockCacheBytes)
	if err != nil {
		return err
	}
	var eg errgroup.Group
	eg.Go(func() error {
		return http.ListenAndServe(fmt.Sprintf(":%v", pfs_server.HTTPPort), httpServer)
	})
	eg.Go(func() error {
		return grpcutil.Serve(
			func(s *grpc.Server) {
				pfsclient.RegisterAPIServer(s, pfsAPIServer)
				pfsclient.RegisterObjectAPIServer(s, blockAPIServer)
				ppsclient.RegisterAPIServer(s, ppsAPIServer)
				cache_pb.RegisterGroupCacheServer(s, cacheServer)
				authclient.RegisterAPIServer(s, authAPIServer)
				healthclient.RegisterHealthServer(s, healthServer)
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

func getKubeClient(env *appEnv) (*kube.Client, error) {
	kubeClient, err := kube.NewInCluster()
	if err != nil {
		log.Errorf("falling back to insecure kube client due to error from NewInCluster: %s", sanitizeErr(err))
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
