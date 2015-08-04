package mainutil

import (
	"fmt"
	"math"
	"net"
	"net/http"
	"os"

	"google.golang.org/grpc"

	"github.com/peter-edge/go-env"
)

func Main(do func(interface{}) error, appEnv interface{}, defaultEnv map[string]string) {
	if err := env.Populate(appEnv, env.PopulateOptions{Defaults: defaultEnv}); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
	if err := do(appEnv); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func GrpcDo(
	port int,
	tracePort int,
	registerFunc func(*grpc.Server),
) error {
	s := grpc.NewServer(grpc.MaxConcurrentStreams(math.MaxUint32))
	registerFunc(s)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	errC := make(chan error)
	go func() { errC <- s.Serve(listener) }()
	if tracePort != 0 {
		go func() { errC <- http.ListenAndServe(fmt.Sprintf(":%d", tracePort), nil) }()
	}
	return <-errC
}
