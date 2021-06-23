package cmds

import (
	"fmt"

	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"

	"github.com/spf13/cobra"
)

// Cmds returns a slice containing admin commands.
func Cmds() []*cobra.Command {
	var commands []*cobra.Command

	inspectCluster := &cobra.Command{
		Short: "Returns info about the pachyderm cluster",
		Long:  "Returns info about the pachyderm cluster",
		RunE: cmdutil.RunFixedArgs(0, func(args []string, env cmdutil.Env) error {
			ci, err := env.Client("user").InspectCluster()
			if err != nil {
				return err
			}
			fmt.Fprintln(env.Out(), ci.ID)
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(inspectCluster, "inspect cluster"))

	return commands
}
