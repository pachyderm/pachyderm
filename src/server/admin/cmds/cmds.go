package cmds

import (
	"context"
	"fmt"

	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachctl"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/version"

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
			cfg, err := config.Read(false, false)
			if err != nil {
				return errors.Wrap(err, "could not read config")
			}
			_, context, err := cfg.ActiveContext(true)
			if err != nil {
				return errors.Wrap(err, "could not get active context")
			}
			ci, err := c.InspectClusterWithVersionAndProject(version.Version, &pfs.Project{Name: context.GetProject()})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			fmt.Println(ci.Id)
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(inspectCluster, "inspect cluster"))

	return commands
}
