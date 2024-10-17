// Package cmds implements pachctl commands for the admin service.
package cmds

import (
	"fmt"
	"os"

	"github.com/pachyderm/pachyderm/v2/src/admin"
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
		Run: cmdutil.RunFixedArgs(0, func(cmd *cobra.Command, args []string) (retErr error) {
			ctx := cmd.Context()
			c, err := pachctlCfg.NewOnUserMachine(ctx, false)
			if err != nil {
				return err
			}
			defer errors.Close(&retErr, c, "close client")
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

	var reason string
	restartPachyderm := &cobra.Command{
		Short: "Schedules all Pachyderm pods to restart as soon as possible.",
		Long: "Schedules all Pachyderm pods to restart as soon as possible.  " +
			"All ongoing work will be lost.",
		Run: cmdutil.RunFixedArgs(0, func(cmd *cobra.Command, args []string) (retErr error) {
			c, err := pachctlCfg.NewOnUserMachine(cmd.Context(), false)
			if err != nil {
				return err
			}
			defer errors.Close(&retErr, c, "close client")
			if _, err := c.AdminAPIClient.RestartPachyderm(c.Ctx(), &admin.RestartPachydermRequest{
				Reason: reason,
			}); err != nil {
				return errors.Wrap(err, "call RestartPachyderm")
			}
			fmt.Fprintf(os.Stderr, "restart requested ok\n")
			return nil
		}),
	}
	restartPachyderm.Flags().StringVarP(&reason, "reason", "r", "", "The reason that you're restarting Pachyderm.")
	commands = append(commands, cmdutil.CreateAlias(restartPachyderm, "restart cluster"))

	return commands
}
