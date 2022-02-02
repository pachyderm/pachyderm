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

	var group string
	listTask := &cobra.Command{
		Hidden: true, // don't show in the list of commands
		Use:    "{{alias}} <service> [<namespace>]",
		Short:  "Return info about tasks in a namespace.",
		Long:   "Return info about tasks in a namespace.",
		Run: cmdutil.RunBoundedArgs(1, 2, func(args []string) error {
			var namespace string
			if len(args) > 1 {
				namespace = args[1]
			}
			if namespace == "" && group != "" {
				return errors.Errorf("must set a task namespace to list a group")
			}

			client, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer client.Close()

			taskInfos, err := client.ListTask(args[0], namespace, group)
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
	listTask.Flags().StringVarP(&group, "group", "g", "", "The group to list within the namespace")
	commands = append(commands, cmdutil.CreateAlias(listTask, "list task"))

	return commands
}
