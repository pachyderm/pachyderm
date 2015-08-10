package main

import (
	"github.com/pachyderm/pachyderm"
	"github.com/pachyderm/pachyderm/src/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/pkg/grpcversion"
	"github.com/pachyderm/pachyderm/src/pkg/mainutil"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/server"
	"github.com/pachyderm/pachyderm/src/pps/store"
	"google.golang.org/grpc"
)

var (
	defaultEnv = map[string]string{
		"PPS_API_PORT": "651",
	}
)

type appEnv struct {
	APIPort   int `env:"PPS_API_PORT"`
	TracePort int `env:"PPS_TRACE_PORT"`
}

func main() {
	mainutil.Main(do, &appEnv{}, defaultEnv)
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	return grpcutil.GrpcDo(
		appEnv.APIPort,
		appEnv.TracePort,
		&grpcversion.Version{
			Major:      pachyderm.MajorVersion,
			Minor:      pachyderm.MinorVersion,
			Micro:      pachyderm.MicroVersion,
			Additional: pachyderm.AdditionalVersion,
		},
		func(s *grpc.Server) {
			pps.RegisterApiServer(s, server.NewAPIServer(store.NewInMemoryClient()))
		},
	)
}
