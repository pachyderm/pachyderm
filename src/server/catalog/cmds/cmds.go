package cmds

import (
	"fmt"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/spf13/cobra"
)

func Cmds() []*cobra.Command {
	var commands []*cobra.Command
	query := &cobra.Command{
		Use:   "{{alias}} <query>",
		Short: "Query the data catalog.",
		Long:  "Query the data catalog.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer c.Close()

			results, err := c.Query(args[0])
			if err != nil {
				return err
			}
			for _, result := range results {
				fmt.Println(result)
			}
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(query, "query"))
	return commands
}
