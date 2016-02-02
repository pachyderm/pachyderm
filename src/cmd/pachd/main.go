package main

import (
	"fmt"

	"github.com/gengo/grpc-gateway/runtime"
	"github.com/pachyderm/pachyderm"
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/route"
	"github.com/pachyderm/pachyderm/src/pfs/server"
	"github.com/pachyderm/pachyderm/src/pkg/discovery"
	"github.com/pachyderm/pachyderm/src/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/pkg/netutil"
	"github.com/pachyderm/pachyderm/src/pkg/shard"
	"go.pedge.io/env"
	"go.pedge.io/lion/proto"
	"go.pedge.io/pkg/http"
	"go.pedge.io/proto/server"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
)

type appEnv struct {
	Port            uint16 `env:"PORT,default=650"`
	HTTPPort        uint16 `env:"HTTP_PORT,default=750"`
	NumShards       uint64 `env:"NUM_SHARDS,default=32"`
	StorageRoot     string `env:"OBJ_ROOT,required"`
	DatabaseAddress string `env:"RETHINK_PORT_28015_TCP_ADDR,required"`
	DatabaseName    string `env:"DATABASE_NAME,default=pachyderm"`
	KubeAddress     string `env:"KUBERNETES_PORT_443_TCP_ADDR,required"`
	EtcdAddress     string `env:"ETCD_PORT_2379_TCP_ADDR,required"`
	Namespace       string `env:NAMESPACE,default=default"`
}

func main() {
	env.Main(do, &appEnv{})
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	etcdClient := getEtcdClient(appEnv)
	_, err := getKubeClient(appEnv)
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
		0,
		appEnv.Namespace,
	)
	driver, err := drive.NewDriver(address)
	if err != nil {
		return err
	}
	apiServer := server.NewAPIServer(
		route.NewSharder(
			appEnv.NumShards,
			1,
		),
		route.NewRouter(
			sharder,
			grpcutil.NewDialer(
				grpc.WithInsecure(),
			),
			address,
		),
	)
	go func() {
		if err := sharder.RegisterFrontend(nil, address, apiServer); err != nil {
			protolion.Printf("Error from sharder.RegisterFrontend %s", err.Error())
		}
	}()
	internalAPIServer := server.NewInternalAPIServer(
		route.NewSharder(
			appEnv.NumShards,
			1,
		),
		route.NewRouter(
			sharder,
			grpcutil.NewDialer(
				grpc.WithInsecure(),
			),
			address,
		),
		driver,
	)
	go func() {
		if err := sharder.Register(nil, address, internalAPIServer); err != nil {
			protolion.Printf("Error from sharder.Register %s", err.Error())
		}
	}()
	return protoserver.ServeWithHTTP(
		func(s *grpc.Server) {
			pfs.RegisterAPIServer(s, apiServer)
			pfs.RegisterInternalAPIServer(s, internalAPIServer)
		},
		func(ctx context.Context, mux *runtime.ServeMux, clientConn *grpc.ClientConn) error {
			return pfs.RegisterAPIHandler(ctx, mux, clientConn)
		},
		protoserver.ServeWithHTTPOptions{
			ServeOptions: protoserver.ServeOptions{
				Version: pachyderm.Version,
			},
		},
		protoserver.ServeEnv{
			GRPCPort: appEnv.Port,
		},
		pkghttp.HandlerEnv{
			Port: appEnv.HTTPPort,
		},
	)
}

func getEtcdClient(env *appEnv) discovery.Client {
	return discovery.NewEtcdClient(fmt.Sprintf("http://%s:2379", env.EtcdAddress))
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
