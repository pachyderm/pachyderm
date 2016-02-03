package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/gengo/grpc-gateway/runtime"
	"github.com/pachyderm/pachyderm"
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
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
)

type appEnv struct {
	NumShards   uint64 `env:"PFS_NUM_SHARDS,default=16"`
	Address     string `env:"PFS_ADDRESS"`
	Port        uint16 `env:"PFS_PORT,default=650"`
	HTTPPort    uint16 `env:"PFS_HTTP_PORT,default=750"`
	EtcdAddress string `env:"ETCD_PORT_2379_TCP_ADDR"`
	ObjdAddress string `env:"OBJD_PORT_652_TCP_ADDR"`
}

func main() {
	env.Main(do, &appEnv{})
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	discoveryClient, err := getEtcdClient(appEnv)
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
	address = fmt.Sprintf("%s:%d", address, appEnv.Port)
	sharder := shard.NewSharder(
		discoveryClient,
		appEnv.NumShards,
		0,
		"namespace",
	)
	objdAddress, err := getObjdAddress(appEnv)
	if err != nil {
		return err
	}
	driver, err := drive.NewDriver(objdAddress)
	if err != nil {
		return err
	}
	apiServer := server.NewAPIServer(
		pfs.NewHasher(
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
	internalAPIServer := server.NewInternalAPIServer(
		pfs.NewHasher(
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
	go func() {
		if err := sharder.Register(nil, address, []shard.Server{internalAPIServer}); err != nil {
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

func getEtcdClient(env *appEnv) (discovery.Client, error) {
	etcdAddress, err := getEtcdAddress(env)
	if err != nil {
		return nil, err
	}
	return discovery.NewEtcdClient(etcdAddress), nil
}

func getEtcdAddress(env *appEnv) (string, error) {
	if env.EtcdAddress == "" {
		return "", errors.New("ETCD_PORT_2379_TCP_ADDR not set")
	}
	return fmt.Sprintf("http://%s:2379", env.EtcdAddress), nil
}

func getObjdAddress(env *appEnv) (string, error) {
	objdAddr := os.Getenv("OBJD_PORT_652_TCP_ADDR")
	if objdAddr == "" {
		return "", errors.New("OBJD_PORT_652_TCP_ADDR not set")
	}
	return fmt.Sprintf("%s:652", objdAddr), nil
}
