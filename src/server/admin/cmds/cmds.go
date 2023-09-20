package cmds

import (
	"context"
	"fmt"
	"os"

	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachctl"

	"github.com/spf13/cobra"
)

// Cmds returns a slice containing admin commands.
func Cmds(mainCtx context.Context, pachctlCfg *pachctl.Config) []*cobra.Command {
	var commands []*cobra.Command

	inspectCluster := &cobra.Command{
		Short: "Returns info about the pachyderm cluster",
		Long:  "Returns info about the pachyderm cluster",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer c.Close()
			if ci, ok := c.ClusterInfo(); ok {
				fmt.Println(ci.Id)
			} else {
				// This should never happen; NewOnUserMachine will error.
				fmt.Fprintf(os.Stderr, "no ClusterInfo received after creating client")
			}
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(inspectCluster, "inspect cluster"))

	return commands
}
