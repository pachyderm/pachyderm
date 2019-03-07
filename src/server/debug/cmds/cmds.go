package cmds

import (
	"os"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// Cmds returns a slice containing debug commands.
func Cmds(noMetrics *bool, noPortForwarding *bool) []*cobra.Command {
	debugDump := &cobra.Command{
		Use:   "debug-dump",
		Short: "Return a dump of running goroutines.",
		Long:  "Return a dump of running goroutines.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			client, err := client.NewOnUserMachine(!*noMetrics, !*noPortForwarding, "debug-dump")
			if err != nil {
				return err
			}
			defer client.Close()
			return client.Dump(os.Stdout)
		}),
	}

	var duration time.Duration
	debugProfile := &cobra.Command{
		Use:   "debug-profile [profile]",
		Short: "Return a profile from the server.",
		Long:  "Return a profile from the server.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := client.NewOnUserMachine(!*noMetrics, !*noPortForwarding, "debug-dump")
			if err != nil {
				return err
			}
			defer client.Close()
			return client.Profile(args[0], duration, os.Stdout)
		}),
	}
	debugProfile.Flags().DurationVarP(&duration, "duration", "d", 0, "Duration to run a CPU profile for.")

	return []*cobra.Command{
		debugDump,
		debugProfile,
	}
}
