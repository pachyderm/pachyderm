package main

import (
	"github.com/pachyderm/pachyderm"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/drive/server"
	"github.com/pachyderm/pachyderm/src/pkg/netutil"
	"go.pedge.io/env"
	"go.pedge.io/proto/server"
	"google.golang.org/grpc"
)

type appEnv struct {
	StorageRoot string `env:"OBJ_ROOT,required"`
	Address     string `env:"OBJ_ADDRESS"`
	Port        int    `env:"OBJ_PORT,default=652"`
	HTTPPort    int    `env:"OBJ_HTTP_PORT,default=752"`
	DebugPort   int    `env:"OBJ_TRACE_PORT,default=1050"`
}

func main() {
	env.Main(do, &appEnv{})
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	var err error
	address := appEnv.Address
	if address == "" {
		address, err = netutil.ExternalIP()
		if err != nil {
			return err
		}
	}
	apiServer, err := server.NewLocalAPIServer(appEnv.StorageRoot)
	if err != nil {
		return err
	}

	return protoserver.Serve(
		uint16(appEnv.Port),
		func(s *grpc.Server) {
			drive.RegisterAPIServer(s, apiServer)
		},
		protoserver.ServeOptions{
			HTTPPort:  uint16(appEnv.HTTPPort),
			DebugPort: uint16(appEnv.DebugPort),
			Version:   pachyderm.Version,
		},
	)
}
