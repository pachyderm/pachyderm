package main

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/route"
	"github.com/pachyderm/pachyderm/src/pfs/server"
	"github.com/pachyderm/pachyderm/src/pkg/btrfs"
	"github.com/pachyderm/pachyderm/src/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/pkg/mainutil"
	"google.golang.org/grpc"
)

const (
	defaultNumShards = 16
	defaultAPIPort   = 650
)

var (
	defaultEnv = map[string]string{
		"PFS_NUM_SHARDS": "16",
		"PFS_API_PORT":   "650",
	}
)

type appEnv struct {
	BtrfsRoot string `env:"PFS_BTRFS_ROOT,required"`
	NumShards int    `env:"PFS_NUM_SHARDS"`
	APIPort   int    `env:"PFS_API_PORT"`
	TracePort int    `env:"PFS_TRACE_PORT"`
}

func main() {
	mainutil.Main(do, &appEnv{}, defaultEnv)
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	address := fmt.Sprintf("0.0.0.0:%d", appEnv.APIPort)
	combinedAPIServer := server.NewCombinedAPIServer(
		route.NewSharder(
			appEnv.NumShards,
		),
		route.NewRouter(
			route.NewSingleAddresser(
				address,
				appEnv.NumShards,
			),
			grpcutil.NewDialer(),
			address,
		),
		drive.NewBtrfsDriver(
			appEnv.BtrfsRoot,
			btrfs.NewFFIAPI(),
		),
	)
	return mainutil.GrpcDo(
		appEnv.APIPort,
		appEnv.TracePort,
		func(s *grpc.Server) {
			pfs.RegisterApiServer(s, combinedAPIServer)
			pfs.RegisterInternalApiServer(s, combinedAPIServer)
		},
	)
}
