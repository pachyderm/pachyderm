package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
	"github.com/pachyderm/pachyderm/v2/src/server/cmd/pachctl/cmd"
	"github.com/spf13/pflag"
)

func main() {
	log.InitPachctlLogger()
	log.SetLevel(log.InfoLevel)
	ctx := pctx.Background("pachctl")

	// Remove kubernetes client flags from the spf13 flag set
	// (we link the kubernetes client, so otherwise they're in 'pachctl --help')
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	tracing.InstallJaegerTracerFromEnv()
	err := func() error {
		defer tracing.CloseAndReportTraces()
		pachctl, err := cmd.PachctlCmd(ctx)
		if err != nil {
			return errors.Wrap(err, "could not create pachctl command")
		}
		return errors.EnsureStack(pachctl.ExecuteContext(ctx))
	}()
	if err != nil {
		if errString := strings.TrimSpace(err.Error()); errString != "" {
			fmt.Fprintf(os.Stderr, "%s\n", errString)
		}
		os.Exit(1)
	}
}
