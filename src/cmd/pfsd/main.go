package main

import (
	"fmt"
	"math"
	"net"

	"net/http"
	//_ "net/http/pprof"

	"github.com/pachyderm/pachyderm/src/common"
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/route"
	"github.com/pachyderm/pachyderm/src/pfs/server"
	"github.com/pachyderm/pachyderm/src/pkg/btrfs"
	"github.com/pachyderm/pachyderm/src/pkg/grpcutil"
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
	common.Main(do, &appEnv{}, defaultEnv)
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	btrfsAPI := btrfs.NewFFIAPI()
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
			btrfsAPI,
		),
	)
	server := grpc.NewServer(grpc.MaxConcurrentStreams(math.MaxUint32))
	pfs.RegisterApiServer(server, combinedAPIServer)
	pfs.RegisterInternalApiServer(server, combinedAPIServer)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", appEnv.APIPort))
	if err != nil {
		return err
	}

	errC := make(chan error)
	go func() { errC <- server.Serve(listener) }()
	//go func() { errC <- http.ListenAndServe(":8080", nil) }()
	if appEnv.TracePort != 0 {
		go func() { errC <- http.ListenAndServe(fmt.Sprintf(":%d", appEnv.TracePort), nil) }()
	}
	return <-errC
}
