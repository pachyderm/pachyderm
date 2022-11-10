package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
	"github.com/pachyderm/pachyderm/v2/src/server/cmd/pachctl/cmd"
	"github.com/spf13/pflag"
)

func main() {
	// Remove kubernetes client flags from the spf13 flag set
	// (we link the kubernetes client, so otherwise they're in 'pachctl --help')
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	tracing.InstallJaegerTracerFromEnv()
	for i, arg := range os.Args {
		fmt.Fprintf(os.Stderr, "ARG NO.%d is %s\n", i, arg)
	}
	err := func() error {
		defer tracing.CloseAndReportTraces()
		return errors.EnsureStack(cmd.PachctlCmd().Execute())
	}()
	if err != nil {
		if errString := strings.TrimSpace(err.Error()); errString != "" {
			fmt.Fprintf(os.Stderr, "%s\n", errString)
		}
		os.Exit(1)
	}
}
