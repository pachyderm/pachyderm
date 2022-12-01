package cmds

import (
	"fmt"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"

	"github.com/spf13/cobra"
)

// Cmds returns a slice containing admin commands.
func Cmds() []*cobra.Command {
	var commands []*cobra.Command

	inspectCluster := &cobra.Command{
		Short: "Returns info about the pachyderm cluster",
		Long:  "Returns info about the pachyderm cluster",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer c.Close()
			ci, err := c.InspectCluster()
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			fmt.Println(ci.ID)
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(inspectCluster, "inspect cluster"))

	return commands
}
