package cmds

import (
	"os"

	pachdclient "github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"

	"github.com/spf13/cobra"
)

func Cmds() []*cobra.Command {
	var commands []*cobra.Command

	var raw bool
	var output string
	outputFlags := cmdutil.OutputFlags(&raw, &output)

	listTask := &cobra.Command{
		Hidden: true, // don't show in the list of commands
		Use:    "{{alias}} <task/name/space>",
		Short:  "Return info about tasks in a namespace.",
		Long:   "Return info about tasks in a namespace.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer client.Close()

			taskInfos, err := client.ListTask(args[0])
			if err != nil {
				return err
			}

			// silently assume raw
			e := cmdutil.Encoder(output, os.Stdout)
			for _, taskInfo := range taskInfos {
				if err := e.EncodeProto(taskInfo); err != nil {
					return errors.EnsureStack(err)
				}
			}
			return nil
		}),
	}
	listTask.Flags().AddFlagSet(outputFlags)
	commands = append(commands, cmdutil.CreateAlias(listTask, "list task"))

	return commands
}
