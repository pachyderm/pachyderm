package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
	"github.com/pachyderm/pachyderm/v2/src/server/cmd/mount-server/cmd"
	"github.com/spf13/pflag"
)

func main() {
	// Remove kubernetes client flags from the spf13 flag set
	// (we link the kubernetes client, so otherwise they're in 'pachctl --help')
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	log.InitPachctlLogger()
	tracing.InstallJaegerTracerFromEnv()
	err := func() error {
		defer tracing.CloseAndReportTraces()
		return errors.EnsureStack(cmd.MountServerCmd().Execute())
	}()
	if err != nil {
		if errString := strings.TrimSpace(err.Error()); errString != "" {
			fmt.Fprintf(os.Stderr, "%s\n", errString)
		}
		os.Exit(1)
	}
}
