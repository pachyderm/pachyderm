package cmds

import (
	"os"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// Cmds returns a slice containing debug commands.
func Cmds(metrics bool, portForwarding bool) []*cobra.Command {
	debugDump := &cobra.Command{
		Use:   "debug-dump",
		Short: "Return a dump of running goroutines.",
		Long:  "Return a dump of running goroutines.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			client, err := client.NewOnUserMachine(metrics, portForwarding, "debug-dump")
			if err != nil {
				return err
			}
			defer client.Close()
			return client.Dump(os.Stdout)
		}),
	}

	return []*cobra.Command{
		debugDump,
	}
}
