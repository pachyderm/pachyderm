package main

import (
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"

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
	pfs_persist "github.com/pachyderm/pachyderm/src/server/pfs/db"
	"github.com/pachyderm/pachyderm/src/server/pfs/drive"
	pfs_server "github.com/pachyderm/pachyderm/src/server/pfs/server"
	cache_pb "github.com/pachyderm/pachyderm/src/server/pkg/cache/groupcachepb"
	cache_server "github.com/pachyderm/pachyderm/src/server/pkg/cache/server"
	"github.com/pachyderm/pachyderm/src/server/pkg/metrics"
	"github.com/pachyderm/pachyderm/src/server/pkg/netutil"
	ppsserver "github.com/pachyderm/pachyderm/src/server/pps"
	"github.com/pachyderm/pachyderm/src/server/pps/persist"
	persist_server "github.com/pachyderm/pachyderm/src/server/pps/persist/server"
	pps_server "github.com/pachyderm/pachyderm/src/server/pps/server"

	flag "github.com/spf13/pflag"
	"go.pedge.io/env"
	"go.pedge.io/lion"
	"go.pedge.io/lion/proto"
	"go.pedge.io/proto/server"
	"google.golang.org/grpc"
	"k8s.io/kubernetes/pkg/api"
	kube_client "k8s.io/kubernetes/pkg/client/restclient"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
)

var readinessCheck bool

func init() {
	flag.BoolVar(&readinessCheck, "readiness-check", false, "Set to true when checking if local pod is ready")
	flag.Parse()
}

type appEnv struct {
	Port               uint16 `env:"PORT,default=650"`
	NumShards          uint64 `env:"NUM_SHARDS,default=32"`
	StorageRoot        string `env:"PACH_ROOT,required"`
	StorageBackend     string `env:"STORAGE_BACKEND,default="`
	DatabaseAddress    string `env:"RETHINK_PORT_28015_TCP_ADDR,required"`
	PPSDatabaseName    string `env:"DATABASE_NAME,default=pachyderm_pps"`
	PFSDatabaseName    string `env:"DATABASE_NAME,default=pachyderm_pfs"`
	KubeAddress        string `env:"KUBERNETES_PORT_443_TCP_ADDR,required"`
	EtcdAddress        string `env:"ETCD_PORT_2379_TCP_ADDR,required"`
	Namespace          string `env:"NAMESPACE,default=default"`
	Metrics            bool   `env:"METRICS,default=true"`
	Init               bool   `env:"INIT,default=false"`
	BlockCacheBytes    int64  `env:"BLOCK_CACHE_BYTES,default=1073741824"` //default = 1 gigabyte
	JobShimImage       string `env:"JOB_SHIM_IMAGE,default="`
	JobImagePullPolicy string `env:"JOB_IMAGE_PULL_POLICY,default="`
}

func main() {
	env.Main(do, &appEnv{})
}

func do(appEnvObj interface{}) error {
	go func() {
		lion.Println(http.ListenAndServe(":651", nil))
	}()
	appEnv := appEnvObj.(*appEnv)
	etcdClient := getEtcdClient(appEnv)
	if appEnv.Init {
		if err := setClusterID(etcdClient); err != nil {
			return fmt.Errorf("error connecting to etcd, if this error persists it likely indicates that kubernetes services are not working correctly. See https://github.com/pachyderm/pachyderm/blob/master/SETUP.md#pachd-or-pachd-init-crash-loop-with-error-connecting-to-etcd for more info")
		}
		rethinkAddress := fmt.Sprintf("%s:28015", appEnv.DatabaseAddress)
		if err := persist_server.InitDBs(rethinkAddress, appEnv.PPSDatabaseName); err != nil {
			return err
		}
		return pfs_persist.InitDB(rethinkAddress, appEnv.PFSDatabaseName)
	}
	if readinessCheck {
		//c, err := client.NewInCluster()
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
	if appEnv.Metrics {
		go metrics.ReportMetrics(clusterID, kubeClient)
	}
	rethinkAPIServer, err := getRethinkAPIServer(appEnv)
	if err != nil {
		return err
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
	driver, err := getPFSDriver(address, appEnv)
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
	go func() {
		if err := sharder.RegisterFrontends(nil, address, []shard.Frontend{cacheServer}); err != nil {
			protolion.Printf("error from sharder.RegisterFrontend %s", sanitizeErr(err))
		}
	}()
	apiServer := pfs_server.NewAPIServer(driver)
	ppsAPIServer := pps_server.NewAPIServer(
		ppsserver.NewHasher(appEnv.NumShards, appEnv.NumShards),
		address,
		kubeClient,
		getNamespace(),
		appEnv.JobShimImage,
		appEnv.JobImagePullPolicy,
	)
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
	return protoserver.Serve(
		func(s *grpc.Server) {
			pfsclient.RegisterAPIServer(s, apiServer)
			pfsclient.RegisterBlockAPIServer(s, blockAPIServer)
			ppsclient.RegisterAPIServer(s, ppsAPIServer)
			ppsserver.RegisterInternalJobAPIServer(s, ppsAPIServer)
			persist.RegisterAPIServer(s, rethinkAPIServer)
			cache_pb.RegisterGroupCacheServer(s, cacheServer)
			healthclient.RegisterHealthServer(s, healthServer)
		},
		protoserver.ServeOptions{
			Version: version.Version,
		},
		protoserver.ServeEnv{
			GRPCPort: appEnv.Port,
		},
	)
}

func getEtcdClient(env *appEnv) discovery.Client {
	return discovery.NewEtcdClient(fmt.Sprintf("http://%s:2379", env.EtcdAddress))
}

const clusterIDKey = "cluster-id"

func setClusterID(client discovery.Client) error {
	return client.Set(clusterIDKey, uuid.NewWithoutDashes(), 0)
}

func getClusterID(client discovery.Client) (string, error) {
	id, err := client.Get(clusterIDKey)
	if err != nil {
		return "", err
	}
	if id == "" {
		return "", fmt.Errorf("clusterID not yet set")
	}
	return id, nil
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

func getPFSDriver(address string, env *appEnv) (drive.Driver, error) {
	rethinkAddress := fmt.Sprintf("%s:28015", env.DatabaseAddress)
	return pfs_persist.NewDriver(address, rethinkAddress, env.PFSDatabaseName)
}

func getRethinkAPIServer(env *appEnv) (persist.APIServer, error) {
	if err := persist_server.CheckDBs(fmt.Sprintf("%s:28015", env.DatabaseAddress), env.PPSDatabaseName); err != nil {
		return nil, err
	}
	return persist_server.NewRethinkAPIServer(fmt.Sprintf("%s:28015", env.DatabaseAddress), env.PPSDatabaseName)
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
