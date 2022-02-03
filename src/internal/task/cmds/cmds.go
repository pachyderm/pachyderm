package cmds

import (
	"os"

	pachdclient "github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/task"

	"github.com/spf13/cobra"
)

func Cmds() []*cobra.Command {
	var commands []*cobra.Command

	var raw bool
	var output string
	outputFlags := cmdutil.OutputFlags(&raw, &output)

	var namespace, group string
	listTask := &cobra.Command{
		Hidden: true, // don't show in the list of commands
		Use:    "{{alias}} <service>",
		Short:  "Return info about tasks from a service.",
		Long:   "Return info about tasks from a service.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			if namespace == "" && group != "" {
				return errors.Errorf("must set a task namespace to list a group")
			}

			client, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer client.Close()

			// silently assume raw
			e := cmdutil.Encoder(output, os.Stdout)
			return client.ListTask(args[0], namespace, group, func(ti *task.TaskInfo) error {
				return errors.EnsureStack(e.EncodeProto(ti))
			})
		}),
	}
	listTask.Flags().AddFlagSet(outputFlags)
	listTask.Flags().StringVarP(&namespace, "namespace", "n", "", "The namespace to restrict to")
	listTask.Flags().StringVarP(&group, "group", "g", "", "The group to list within the namespace")
	commands = append(commands, cmdutil.CreateAlias(listTask, "list task"))

	return commands
}
