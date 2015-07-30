package main

import (
	"fmt"
	"math"
	"net"
	"os"

	"net/http"
	//_ "net/http/pprof"

	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/server"
	"github.com/pachyderm/pachyderm/src/pps/store"
	"github.com/peter-edge/go-env"
	"google.golang.org/grpc"
)

const (
	defaultAPIPort = 651
)

type appEnv struct {
	APIPort   int `env:"PPS_API_PORT"`
	TracePort int `env:"PPS_TRACE_PORT"`
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
	if appEnv.APIPort == 0 {
		appEnv.APIPort = defaultAPIPort
	}
	//address := fmt.Sprintf("0.0.0.0:%d", appEnv.APIPort)
	s := grpc.NewServer(grpc.MaxConcurrentStreams(math.MaxUint32))
	pps.RegisterApiServer(s, server.NewAPIServer(store.NewInMemoryClient()))
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", appEnv.APIPort))
	if err != nil {
		return err
	}

	errC := make(chan error)
	go func() { errC <- s.Serve(listener) }()
	//go func() { errC <- http.ListenAndServe(":8080", nil) }()
	if appEnv.TracePort != 0 {
		go func() { errC <- http.ListenAndServe(fmt.Sprintf(":%d", appEnv.TracePort), nil) }()
	}
	return <-errC
}
