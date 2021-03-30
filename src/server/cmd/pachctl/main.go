package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/tracing"
	"github.com/pachyderm/pachyderm/src/server/cmd/pachctl/cmd"
	"github.com/spf13/pflag"

	"net/http"
	//_ "net/http/pprof"
)

func main() {
	go func() {
		fmt.Println("serving debug server on 0.0.0.0:8081")
		http.ListenAndServe("0.0.0.0:8081", nil)
		fmt.Println("server exited")
	}()

	// fmt.Println("writing cpu profile to prof")
	// fd, err := os.OpenFile("prof", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	// if err != nil {
	// 	panic(fmt.Sprintf("create prof: %v", err))
	// }
	// if err := pprof.StartCPUProfile(fd); err != nil {
	// 	panic(fmt.Sprintf("start profile: %v", err))
	// }

	// Remove kubernetes client flags from the spf13 flag set
	// (we link the kubernetes client, so otherwise they're in 'pachctl --help')
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	tracing.InstallJaegerTracerFromEnv()
	var duration time.Duration
	err := func() error {
		start := time.Now()
		defer tracing.CloseAndReportTraces()
		defer func() { duration = time.Since(start) }()
		return cmd.PachctlCmd().Execute()
	}()
	fmt.Printf("copy: %v bytes in %v (%v/s)\n", humanize.Bytes(uint64(client.BytesSent)), time.Duration(client.TimeTaken).String(), humanize.Bytes(uint64(float64(client.BytesSent)/time.Duration(client.TimeTaken).Seconds())))
	fmt.Printf("wall: %v bytes in %v (%v/s)\n", humanize.Bytes(uint64(client.BytesSent)), duration.String(), humanize.Bytes(uint64(float64(client.BytesSent)/duration.Seconds())))
	// pprof.StopCPUProfile()
	// if err := fd.Close(); err != nil {
	// 	panic(fmt.Sprintf("close prof: %v", err))
	// }
	// fmt.Println("wrote prof")
	if err != nil {
		if errString := strings.TrimSpace(err.Error()); errString != "" {
			fmt.Fprintf(os.Stderr, "%s\n", errString)
		}
		os.Exit(1)
	}
}
