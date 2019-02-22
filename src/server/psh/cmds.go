package psh

import (
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func Cmds(noMetrics *bool, noPortForwarding *bool) []*cobra.Command {
	shell := &cobra.Command{
		Use:   "shell",
		Short: "Run the Pachyderm Shell.",
		Long:  "Run the Pachyderm Shell.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			return Shell()
		}),
	}
	return []*cobra.Command{shell}
}
