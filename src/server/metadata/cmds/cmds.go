package cmds

import (
	"fmt"
	"os"

	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachctl"
	"github.com/pachyderm/pachyderm/v2/src/metadata"

	"github.com/spf13/cobra"
)

// Cmds returns a slice containing metadata commands.
func Cmds(pachctlCfg *pachctl.Config) []*cobra.Command {
	var commands []*cobra.Command

	editMetadata := &cobra.Command{
		Short: "Edits an object's metadata",
		Long:  "Edits an object's metadata",
		Run: cmdutil.RunFixedArgsCmd(0, func(cmd *cobra.Command, args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(cmd.Context(), false)
			if err != nil {
				return err
			}
			defer c.Close()
			if _, err := c.MetadataClient.EditMetadata(c.Ctx(), &metadata.EditMetadataRequest{}); err != nil {
				return errors.Wrap(err, "invoke EditMetadata")
			}
			fmt.Fprintf(os.Stderr, "ok\n")
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(editMetadata, "edit metadata"))

	return commands
}
