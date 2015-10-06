package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/net/context"
	"golang.org/x/net/trace"

	"go.pedge.io/env"
	"go.pedge.io/proto/server"
	"go.pedge.io/protolog/logrus"

	"github.com/gengo/grpc-gateway/runtime"
	"go.pachyderm.com/pachyderm"
	"go.pachyderm.com/pachyderm/src/pfs"
	"go.pachyderm.com/pachyderm/src/pfs/drive"
	"go.pachyderm.com/pachyderm/src/pfs/drive/btrfs"
	"go.pachyderm.com/pachyderm/src/pfs/route"
	"go.pachyderm.com/pachyderm/src/pfs/server"
	"go.pachyderm.com/pachyderm/src/pkg/discovery"
	"go.pachyderm.com/pachyderm/src/pkg/grpcutil"
	"go.pachyderm.com/pachyderm/src/pkg/netutil"
	"google.golang.org/grpc"
)

var (
	defaultEnv = map[string]string{
		"PFS_NUM_SHARDS":   "16",
		"PFS_NUM_REPLICAS": "0",
		"PFS_PORT":         "650",
		"PFS_HTTP_PORT":    "750",
		"PFS_TRACE_PORT":   "1050",
		"PFS_DRIVER_TYPE":  "btrfs",
	}
)

type appEnv struct {
	DriverRoot  string `env:"PFS_DRIVER_ROOT,required"`
	DriverType  string `env:"PFS_DRIVER_TYPE"`
	NumShards   uint64 `env:"PFS_NUM_SHARDS"`
	NumReplicas uint64 `env:"PFS_NUM_REPLICAS"`
	Address     string `env:"PFS_ADDRESS"`
	Port        int    `env:"PFS_PORT"`
	HTTPPort    int    `env:"PFS_HTTP_PORT"`
	DebugPort   int    `env:"PFS_TRACE_PORT"`
}

func main() {
	env.Main(do, &appEnv{}, defaultEnv)
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	logrus.Register()
	discoveryClient, err := getEtcdClient()
	if err != nil {
		return err
	}
	address := appEnv.Address
	if address == "" {
		address, err = netutil.ExternalIP()
		if err != nil {
			return err
		}
	}
	sharder := route.NewSharder(appEnv.NumShards, appEnv.NumReplicas)
	address = fmt.Sprintf("%s:%d", address, appEnv.Port)
	addresser := route.NewDiscoveryAddresser(
		discoveryClient,
		sharder,
		"namespace",
	)
	var driver drive.Driver
	switch appEnv.DriverType {
	case "btrfs":
		driver, err = btrfs.NewDriver(appEnv.DriverRoot, "")
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown value for PFS_DRIVER_TYPE: %s", appEnv.DriverType)
	}
	apiServer := server.NewAPIServer(
		route.NewSharder(
			appEnv.NumShards,
			0,
		),
		route.NewRouter(
			addresser,
			grpcutil.NewDialer(
				grpc.WithInsecure(),
			),
			address,
		),
	)
	internalAPIServer := server.NewInternalAPIServer(
		route.NewSharder(
			appEnv.NumShards,
			0,
		),
		route.NewRouter(
			addresser,
			grpcutil.NewDialer(
				grpc.WithInsecure(),
			),
			address,
		),
		driver,
	)
	go func() {
		if err := addresser.Register(nil, "id", address, internalAPIServer); err != nil {
			log.Print(err)
		}
	}()
	// TODO(pedge): no!
	trace.AuthRequest = func(_ *http.Request) (bool, bool) {
		return true, true
	}
	return protoserver.Serve(
		uint16(appEnv.Port),
		func(s *grpc.Server) {
			pfs.RegisterApiServer(s, apiServer)
			pfs.RegisterInternalApiServer(s, internalAPIServer)
		},
		protoserver.ServeOptions{
			HTTPPort:  uint16(appEnv.HTTPPort),
			DebugPort: uint16(appEnv.DebugPort),
			Version:   pachyderm.Version,
			HTTPRegisterFunc: func(ctx context.Context, mux *runtime.ServeMux, clientConn *grpc.ClientConn) error {
				return pfs.RegisterApiHandler(ctx, mux, clientConn)
			},
		},
	)
}

func getEtcdClient() (discovery.Client, error) {
	etcdAddress, err := getEtcdAddress()
	if err != nil {
		return nil, err
	}
	return discovery.NewEtcdClient(etcdAddress), nil
}

func getEtcdAddress() (string, error) {
	etcdAddr := os.Getenv("ETCD_PORT_2379_TCP_ADDR")
	if etcdAddr == "" {
		return "", errors.New("ETCD_PORT_2379_TCP_ADDR not set")
	}
	return fmt.Sprintf("http://%s:2379", etcdAddr), nil
}
