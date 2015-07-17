package main

import (
	"fmt"
	"math"
	"net"
	"os"

	"net/http"
	_ "net/http/pprof"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/address"
	"github.com/pachyderm/pachyderm/src/pfs/dial"
	"github.com/pachyderm/pachyderm/src/pfs/drive/btrfs"
	"github.com/pachyderm/pachyderm/src/pfs/route"
	"github.com/pachyderm/pachyderm/src/pfs/server"
	"github.com/pachyderm/pachyderm/src/pfs/shard"
	"github.com/peter-edge/go-env"
	"google.golang.org/grpc"
)

const (
	defaultNumShards = 16
)

type appEnv struct {
	BtrfsRoot string `env:"PFS_BTRFS_ROOT,required"`
	NumShards int    `env:"PFS_NUM_SHARDS"`
	APIPort   int    `env:"PFS_API_PORT,required"`
	TracePort int    `env:"PFS_TRACE_PORT"`
}

func main() {
	if err := do(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func do() error {
	appEnv := &appEnv{}
	if err := env.Populate(appEnv, env.PopulateOptions{}); err != nil {
		return err
	}
	if appEnv.NumShards == 0 {
		appEnv.NumShards = defaultNumShards
	}
	a := fmt.Sprintf("0.0.0.0:%d", appEnv.APIPort)
	combinedAPIServer := server.NewCombinedAPIServer(
		shard.NewSharder(
			appEnv.NumShards,
		),
		route.NewRouter(
			address.NewSingleAddresser(
				a,
				appEnv.NumShards,
			),
			dial.NewDialer(),
			a,
		),
		btrfs.NewDriver(appEnv.BtrfsRoot),
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
