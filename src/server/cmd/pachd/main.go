package main

import (
	"fmt"
	"os"

	"github.com/pachyderm/pachyderm/src/client"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/discovery"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pkg/shard"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps" //SJ: bad name conflict w below
	"github.com/pachyderm/pachyderm/src/client/version"
	pfsmodel "github.com/pachyderm/pachyderm/src/server/pfs" // SJ: really bad name conflict. Normally I was making the non pfsclient stuff all under pfs server
	"github.com/pachyderm/pachyderm/src/server/pfs/drive"
	pfs_server "github.com/pachyderm/pachyderm/src/server/pfs/server"
	"github.com/pachyderm/pachyderm/src/server/pkg/metrics"
	"github.com/pachyderm/pachyderm/src/server/pkg/netutil"
	ppsserver "github.com/pachyderm/pachyderm/src/server/pps" //SJ: cant name this server per the refactor convention because of the import below
	"github.com/pachyderm/pachyderm/src/server/pps/persist"
	persist_server "github.com/pachyderm/pachyderm/src/server/pps/persist/server"
	pps_server "github.com/pachyderm/pachyderm/src/server/pps/server"

	flag "github.com/spf13/pflag"
	"go.pedge.io/env"
	"go.pedge.io/lion/proto"
	"go.pedge.io/proto/server"
	"google.golang.org/grpc"
	"k8s.io/kubernetes/pkg/api"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
)

var readinessCheck bool

func init() {
	flag.BoolVar(&readinessCheck, "readiness-check", false, "Set to true when checking if local pod is ready")
	flag.Parse()
}

type appEnv struct {
	Port            uint16 `env:"PORT,default=650"`
	NumShards       uint64 `env:"NUM_SHARDS,default=32"`
	StorageRoot     string `env:"PACH_ROOT,required"`
	StorageBackend  string `env:"STORAGE_BACKEND,default="`
	DatabaseAddress string `env:"RETHINK_PORT_28015_TCP_ADDR,required"`
	DatabaseName    string `env:"DATABASE_NAME,default=pachyderm"`
	KubeAddress     string `env:"KUBERNETES_PORT_443_TCP_ADDR,required"`
	EtcdAddress     string `env:"ETCD_PORT_2379_TCP_ADDR,required"`
	Namespace       string `env:"NAMESPACE,default=default"`
	Metrics         bool   `env:"METRICS,default=true"`
	Init            bool   `env:"INIT,default=false"`
}

func main() {
	env.Main(do, &appEnv{})
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	etcdClient := getEtcdClient(appEnv)
	if appEnv.Init {
		if err := setClusterID(etcdClient); err != nil {
			return err
		}
		if err := persist_server.InitDBs(fmt.Sprintf("%s:28015", appEnv.DatabaseAddress), appEnv.DatabaseName); err != nil {
			return err
		}
		return nil
	}
	if readinessCheck {
		//c, err := client.NewInCluster()
		c, err := client.NewFromAddress("127.0.0.1:650")
		if err != nil {
			return err
		}

		_, err = c.ListRepo(nil)
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
			protolion.Printf("Error from sharder.AssignRoles: %s", err.Error())
		}
	}()
	driver, err := drive.NewDriver(address)
	if err != nil {
		return err
	}
	apiServer := pfs_server.NewAPIServer(
		pfsmodel.NewHasher(
			appEnv.NumShards,
			1,
		),
		shard.NewRouter(
			sharder,
			grpcutil.NewDialer(
				grpc.WithInsecure(),
			),
			address,
		),
	)
	go func() {
		if err := sharder.RegisterFrontends(nil, address, []shard.Frontend{apiServer}); err != nil {
			protolion.Printf("Error from sharder.RegisterFrontend %s", err.Error())
		}
	}()
	internalAPIServer := pfs_server.NewInternalAPIServer(
		pfsmodel.NewHasher(
			appEnv.NumShards,
			1,
		),
		shard.NewRouter(
			sharder,
			grpcutil.NewDialer(
				grpc.WithInsecure(),
			),
			address,
		),
		driver,
	)
	ppsAPIServer := pps_server.NewAPIServer(
		ppsserver.NewHasher(appEnv.NumShards, appEnv.NumShards),
		address,
		kubeClient,
		getNamespace(),
	)
	go func() {
		if err := sharder.Register(nil, address, []shard.Server{internalAPIServer, ppsAPIServer}); err != nil {
			protolion.Printf("Error from sharder.Register %s", err.Error())
		}
	}()
	blockAPIServer, err := pfs_server.NewBlockAPIServer(appEnv.StorageRoot, appEnv.StorageBackend)
	if err != nil {
		return err
	}
	return protoserver.Serve(
		func(s *grpc.Server) {
			pfsclient.RegisterAPIServer(s, apiServer)
			pfsclient.RegisterInternalAPIServer(s, internalAPIServer)
			pfsclient.RegisterBlockAPIServer(s, blockAPIServer)
			ppsclient.RegisterAPIServer(s, ppsAPIServer)
			ppsserver.RegisterInternalJobAPIServer(s, ppsAPIServer)
			persist.RegisterAPIServer(s, rethinkAPIServer)
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
		protolion.Errorf("Falling back to insecure kube client due to error from NewInCluster: %s", err.Error())
	} else {
		return kubeClient, err
	}
	config := &kube.Config{
		Host:     fmt.Sprintf("%s:443", env.KubeAddress),
		Insecure: true,
	}
	return kube.New(config)
}

func getRethinkAPIServer(env *appEnv) (persist.APIServer, error) {
	if err := persist_server.CheckDBs(fmt.Sprintf("%s:28015", env.DatabaseAddress), env.DatabaseName); err != nil {
		return nil, err
	}
	return persist_server.NewRethinkAPIServer(fmt.Sprintf("%s:28015", env.DatabaseAddress), env.DatabaseName)
}

// getNamespace returns the kubernetes namespace that this pachd pod runs in
func getNamespace() string {
	namespace := os.Getenv("PACHD_POD_NAMESPACE")
	if namespace != "" {
		return namespace
	}
	return api.NamespaceDefault
}
