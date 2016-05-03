package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.pedge.io/pkg/cobra"
)

func RunFixedArgs(numArgs int, run func([]string) error) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		if len(args) != numArgs {
			fmt.Printf("Expected %d arguments, got %d.\n\n", numArgs, len(args))
			cmd.Usage()
		} else {
			if err := run(args); err != nil {
				pkgcobra.ErrorAndExit("%v", err)
			}
		}
	}
}

func RunBoundedArgs(min int, max int, run func([]string) error) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		if len(args) < min || len(args) > max {
			fmt.Printf("Expected %d to %d arguments, got %d.\n\n", min, max, len(args))
			cmd.Usage()
		} else {
			if err := run(args); err != nil {
				pkgcobra.ErrorAndExit("%v", err)
			}
		}
	}
}
