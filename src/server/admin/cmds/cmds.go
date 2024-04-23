package cmds

import (
	"fmt"
	"os"

	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachctl"

	"github.com/spf13/cobra"
)

// Cmds returns a slice containing admin commands.
func Cmds(pachctlCfg *pachctl.Config) []*cobra.Command {
	var commands []*cobra.Command

	var raw bool
	var output string
	outputFlags := cmdutil.OutputFlags(&raw, &output)

	inspectCluster := &cobra.Command{
		Short: "Returns info about the Pachyderm cluster",
		Long:  "Returns info about the Pachyderm cluster",
		Run: cmdutil.RunFixedArgs(0, func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			c, err := pachctlCfg.NewOnUserMachine(ctx, false)
			if err != nil {
				return err
			}
			defer c.Close()
			if ci, ok := c.ClusterInfo(); ok {
				if raw {
					return errors.EnsureStack(cmdutil.Encoder(output, os.Stdout).EncodeProto(ci))
				} else {
					fmt.Println(ci.Id)
				}
			} else {
				// This should never happen; NewOnUserMachine will error.
				fmt.Fprintf(os.Stderr, "no ClusterInfo received after creating client")
			}
			return nil
		}),
	}
	inspectCluster.Flags().AddFlagSet(outputFlags)
	commands = append(commands, cmdutil.CreateAlias(inspectCluster, "inspect cluster"))

	return commands
}
