package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/pachyderm/pachyderm/src/client/pkg/tracing"
	"github.com/pachyderm/pachyderm/src/server/cmd/pachctl/cmd"
	"github.com/spf13/pflag"
)

func main() {
	// Remove kubernetes client flags from the spf13 flag set
	// (we link the kubernetes client, so otherwise they're in 'pachctl --help')
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	jaegerEndpoint := tracing.InstallJaegerTracerFromEnv()
	if os.Args[1] != "version" && jaegerEndpoint != "" {
		log.Printf("using Jaeger collector endpoint: %s\n", jaegerEndpoint)
	}
	err := func() error {
		defer tracing.CloseAndReportTraces()
		rootCmd, err := cmd.PachctlCmd()
		if err != nil {
			return err
		}
		return rootCmd.Execute()
	}()
	if err != nil {
		if errString := strings.TrimSpace(err.Error()); errString != "" {
			fmt.Fprintf(os.Stderr, "%s\n", errString)
		}
		os.Exit(1)
	}
}
