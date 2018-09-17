package cmds

import (
	"os"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// Cmds returns a slice containing debug commands.
func Cmds(noMetrics *bool) []*cobra.Command {
	metrics := !*noMetrics

	debugDump := &cobra.Command{
		Use:   "debug-dump",
		Short: "Return a dump of running goroutines.",
		Long:  "Return a dump of running goroutines.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			client, err := client.NewOnUserMachine(metrics, "debug-dump")
			if err != nil {
				return err
			}
			return client.Dump(os.Stdout)
		}),
	}

	return []*cobra.Command{
		debugDump,
	}
}
