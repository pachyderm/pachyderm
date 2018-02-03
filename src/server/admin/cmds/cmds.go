package cmds

import (
	"os"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"

	"github.com/spf13/cobra"
)

// Cmds returns a slice containing admin commands.
func Cmds(noMetrics *bool) []*cobra.Command {
	metrics := !*noMetrics

	extract := &cobra.Command{
		Use:   "extract",
		Short: "Extract Pachyderm state to stdout.",
		Long:  "Extract Pachyderm state to stdout.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := client.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
			return c.ExtractWriter(os.Stdout)
		}),
	}
	restore := &cobra.Command{
		Use:   "restore",
		Short: "Restore Pachyderm state from stdin.",
		Long:  "Restore Pachyderm state from stdin.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := client.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
			return c.RestoreReader(os.Stdin)
		}),
	}
	return []*cobra.Command{extract, restore}
}
