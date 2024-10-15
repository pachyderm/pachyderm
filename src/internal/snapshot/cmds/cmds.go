// Package cmds implements commands for snapshot
package cmds

import (
	"os"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachctl"
	"github.com/pachyderm/pachyderm/v2/src/internal/snapshot/pretty"
	"github.com/pachyderm/pachyderm/v2/src/internal/tabwriter"
	txncmds "github.com/pachyderm/pachyderm/v2/src/server/transaction/cmds"
	"github.com/pachyderm/pachyderm/v2/src/snapshot"
	"github.com/spf13/cobra"
)

func Cmds(pachctlCfg *pachctl.Config) []*cobra.Command {
	var commands []*cobra.Command

	var raw bool
	var output string
	outputFlags := cmdutil.OutputFlags(&raw, &output)

	createSnapshot := &cobra.Command{
		Short: "Create a new snapshot",
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
	commands = append(commands, cmdutil.CreateAliases(createSnapshot, "create snapshot", "snapshots"))

	listSnapshot := &cobra.Command{
		Short: "Return all snapshots",
		Long:  "This command returns all snapshots",
		Run: cmdutil.RunFixedArgs(0, func(cmd *cobra.Command, args []string) (retErr error) {
			c, err := pachctlCfg.NewOnUserMachine(cmd.Context(), false)
			if err != nil {
				return err
			}
			defer errors.Close(&retErr, c, "close client")
			snapshotClient, err := c.SnapshotAPIClient.ListSnapshot(
				c.Ctx(),
				&snapshot.ListSnapshotRequest{},
			)
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			snapshots, err := grpcutil.Collect[*snapshot.ListSnapshotResponse](snapshotClient, 100)
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			if raw {
				encoder := cmdutil.Encoder(output, os.Stdout)
				for _, info := range snapshots {
					if err := encoder.EncodeProto(info.Info); err != nil {
						return errors.EnsureStack(err)
					}
				}
				return nil
			}
			header := pretty.SnapshotHeader
			writer := tabwriter.NewWriter(os.Stdout, header)
			for _, snapshotInfo := range snapshots {
				pretty.PrintSnapshotInfo(writer, snapshotInfo.Info)
			}
			return writer.Flush()
		}),
	}
	listSnapshot.Flags().AddFlagSet(outputFlags)
	commands = append(commands, cmdutil.CreateAliases(listSnapshot, "list snapshot", "snapshots"))

	return commands
}
