package cmds

import (
	"encoding/json"
	"os"

	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachctl"
	"github.com/pachyderm/pachyderm/v2/src/task"

	"github.com/spf13/cobra"
)

func Cmds(pachctlCfg *pachctl.Config) []*cobra.Command {
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
		Run: cmdutil.RunFixedArgsCmd(1, func(cmd *cobra.Command, args []string) error {
			if namespace == "" && group != "" {
				return errors.Errorf("must set a task namespace to list a group")
			}

			client, err := pachctlCfg.NewOnUserMachine(cmd.Context(), false)
			if err != nil {
				return err
			}
			defer client.Close()

			// silently assume raw
			e := cmdutil.Encoder(output, os.Stdout)
			return client.ListTask(args[0], namespace, group, func(ti *task.TaskInfo) error {
				return errors.EnsureStack(e.EncodeProtoTransform(ti, func(in map[string]interface{}) error {
					// the input data field is transmitted as a JSON string, unmarshal it for easier reading
					if inputData, ok := in["input_data"]; ok {
						if inputString, ok := inputData.(string); ok {
							// if not a string, preserve as is rather than error
							var holder map[string]interface{}
							if err := json.Unmarshal([]byte(inputString), &holder); err != nil {
								return errors.EnsureStack(err)
							}
							in["input_data"] = holder
						}
					}
					return nil
				}))
			})
		}),
	}
	listTask.Flags().AddFlagSet(outputFlags)
	listTask.Flags().StringVarP(&namespace, "namespace", "n", "", "The namespace to restrict to")
	listTask.Flags().StringVarP(&group, "group", "g", "", "The group to list within the namespace")
	commands = append(commands, cmdutil.CreateAlias(listTask, "list task"))

	return commands
}
