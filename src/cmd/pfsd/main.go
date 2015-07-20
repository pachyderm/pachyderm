package main

import (
	"fmt"
	"math"
	"net"
	"os"
	"runtime"

	"net/http"
	_ "net/http/pprof"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/route"
	"github.com/pachyderm/pachyderm/src/pfs/server"
	"github.com/pachyderm/pachyderm/src/pkg/btrfs"
	"github.com/peter-edge/go-env"
	"google.golang.org/grpc"
)

const (
	defaultNumShards = 16
)

type appEnv struct {
	BtrfsDriverType string `env:"PFS_BTRFS_DRIVER_TYPE"`
	BtrfsRoot       string `env:"PFS_BTRFS_ROOT,required"`
	NumShards       int    `env:"PFS_NUM_SHARDS"`
	APIPort         int    `env:"PFS_API_PORT,required"`
	TracePort       int    `env:"PFS_TRACE_PORT"`
}

func main() {
	if err := do(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func do() error {
	runtime.GOMAXPROCS(runtime.NumCPU())

	appEnv := &appEnv{}
	if err := env.Populate(appEnv, env.PopulateOptions{}); err != nil {
		return err
	}
	if appEnv.NumShards == 0 {
		appEnv.NumShards = defaultNumShards
	}
	var btrfsAPI btrfs.API
	switch appEnv.BtrfsDriverType {
	case "exec":
		btrfsAPI = btrfs.NewExecAPI()
	case "ffi":
		fallthrough
	default:
		btrfsAPI = btrfs.NewFFIAPI()
	}
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
			route.NewDialer(),
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
	go func() { errC <- http.ListenAndServe(":8080", nil) }()
	if appEnv.TracePort != 0 {
		go func() { errC <- http.ListenAndServe(fmt.Sprintf(":%d", appEnv.TracePort), nil) }()
	}
	return <-errC
}
