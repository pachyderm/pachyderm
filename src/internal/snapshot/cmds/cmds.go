// Package cmds implements commands for snapshot
package cmds

import (
	"os"
	"strconv"

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

const snapshots = "snapshots"

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
				return errors.Wrap(err, "create snapshot")
			})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	commands = append(commands, cmdutil.CreateAliases(createSnapshot, "create snapshot", snapshots))

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

			if raw {
				encoder := cmdutil.Encoder(output, os.Stdout)
				if err := grpcutil.ForEach(snapshotClient, func(res *snapshot.ListSnapshotResponse) error {
					if err := encoder.EncodeProto(res); err != nil {
						return errors.Wrap(err, "encode proto")
					}
					return nil
				}); err != nil {
					return grpcutil.ScrubGRPC(err)
				}

			} else {
				writer := tabwriter.NewWriter(os.Stdout, pretty.SnapshotHeader)
				defer errors.Invoke(&retErr, writer.Flush, "flush output")
				if err := grpcutil.ForEach(snapshotClient, func(res *snapshot.ListSnapshotResponse) error {
					pretty.PrintSnapshotInfo(writer, res.GetInfo())
					return nil
				}); err != nil {
					return grpcutil.ScrubGRPC(err)
				}
			}
			return nil
		}),
	}
	listSnapshot.Flags().AddFlagSet(outputFlags)
	commands = append(commands, cmdutil.CreateAliases(listSnapshot, "list snapshot", snapshots))

	var idStr string
	deleteSnapshot := &cobra.Command{
		Short: "Delete a snapshot",
		Long:  "This command deletes a snapshot",
		Run: cmdutil.RunBoundedArgs(0, 1, func(cmd *cobra.Command, args []string) (retErr error) {
			ctx := cmd.Context()
			c, err := pachctlCfg.NewOnUserMachine(ctx, false)
			if err != nil {
				return err
			}
			defer errors.Close(&retErr, c, "close client")

			request := &snapshot.DeleteSnapshotRequest{}
			if len(args) == 0 {
				return errors.Errorf("a snapshot id needs to be provided")
			}
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return errors.Wrap(err, "parse int")
			}
			request.Id = id

			err = txncmds.WithActiveTransaction(c, func(c *client.APIClient) error {
				_, err := c.SnapshotAPIClient.DeleteSnapshot(c.Ctx(), request)
				return errors.Wrap(err, "delete snapshot RPC")
			})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	deleteSnapshot.Flags().StringVar(&idStr, "snapshot", idStr, "Specify the snapshot (by id) to delete.")
	commands = append(commands, cmdutil.CreateAliases(deleteSnapshot, "delete snapshot", snapshots))

	inspectSnapshot := &cobra.Command{
		Short: "Return info about a snapshot.",
		Long:  "This command returns details of the snapshot such as: `ID`, `ChunksetID`, `Fileset`, `Version` and `Created`.",
		Run: cmdutil.RunFixedArgs(1, func(cmd *cobra.Command, args []string) (retErr error) {
			c, err := pachctlCfg.NewOnUserMachine(cmd.Context(), false)
			if err != nil {
				return err
			}
			defer errors.Close(&retErr, c, "close client")

			request := &snapshot.InspectSnapshotRequest{}
			if len(args) == 0 {
				return errors.Errorf("a snapshot id needs to be provided")
			}
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return errors.Wrap(err, "parse int")
			}
			request.Id = id

			var inspectResp *snapshot.InspectSnapshotResponse
			if err := txncmds.WithActiveTransaction(c, func(c *client.APIClient) error {
				resp, err := c.SnapshotAPIClient.InspectSnapshot(c.Ctx(), request)
				if err != nil {
					return errors.Wrap(err, "inspect snapshot RPC")
				}
				inspectResp = resp
				return nil
			}); err != nil {
				return grpcutil.ScrubGRPC(err)
			}

			if raw {
				return errors.Wrap(cmdutil.Encoder(output, os.Stdout).EncodeProto(inspectResp.Info), "encoder")
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}
			return pretty.PrintDetailedSnapshotInfo(inspectResp)
		}),
	}
	inspectSnapshot.Flags().AddFlagSet(outputFlags)
	commands = append(commands, cmdutil.CreateAliases(inspectSnapshot, "inspect snapshot", snapshots))
	return commands
}
