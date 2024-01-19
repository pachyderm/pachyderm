package main

import (
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachdev"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/signals"
)

var verbose bool

func main() {
	rootCmd := &cobra.Command{
		Use:           "pachdev",
		Short:         "A CLI tool for Pachyderm development",
		SilenceUsage:  true, // This avoids printing the usage message on errors.
		SilenceErrors: true, // We print out the error ourselves.
	}
	// Common flags.
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "If true, show debug-level log messages.")

	// Subcommands.
	rootCmd.AddCommand(pachdev.DeleteClusterCmd())
	rootCmd.AddCommand(pachdev.CreateClusterCmd())
	rootCmd.AddCommand(pachdev.LoadImageCmd())

	// Run a command.
	log.InitPachctlLogger()
	ctx, c := signal.NotifyContext(pctx.Background(""), signals.TerminationSignals...)
	defer c()
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		log.Exit(ctx, err.Error())
	}
}
