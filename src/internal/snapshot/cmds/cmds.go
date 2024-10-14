package cmds

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachctl"
	txncmds "github.com/pachyderm/pachyderm/v2/src/server/transaction/cmds"
	"github.com/pachyderm/pachyderm/v2/src/snapshot"
	"github.com/spf13/cobra"
)

func Cmds(pachctlCfg *pachctl.Config) []*cobra.Command {
	var commands []*cobra.Command

	createSnapshot := &cobra.Command{
		Short: "create a new snapshot",
		Long:  "This command creates a snapshot for disaster recovery",
		Run: cmdutil.RunFixedArgs(0, func(cmd *cobra.Command, args []string) (retErr error) {
			ctx := cmd.Context()
			c, err := pachctlCfg.NewOnUserMachine(ctx, false)
			if err != nil {
				return err
			}
			defer errors.Close(&retErr, c, "close client")
			err = txncmds.WithActiveTransaction(c, func(c *client.APIClient) error {
				_, err = c.SnapshotAPIClient.CreateSnapshot(
					c.Ctx(),
					&snapshot.CreateSnapshotRequest{},
				)
				return errors.EnsureStack(err)
			})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(createSnapshot, "create snapshot"))

	return commands
}
