package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/pachyderm/pachyderm/src/server/cmd/pachctl/cmd"
	"github.com/spf13/pflag"
)

func main() {
	// Remove kubernetes client flags from the spf13 flag set
	// (we link the kubernetes client, so otherwise they're in 'pachctl --help')
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	err := func() error {
		rootCmd, err := cmd.PachctlCmd()
		if err != nil {
			return err
		}
		if err := rootCmd.Execute(); err != nil {
			return err
		}
		return nil
	}()
	if err != nil {
		if errString := strings.TrimSpace(err.Error()); errString != "" {
			fmt.Fprintf(os.Stderr, "%s\n", errString)
		}
		os.Exit(1)
	}
}
