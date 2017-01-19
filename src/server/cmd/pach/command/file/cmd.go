package file

import "github.com/urfave/cli"

// NewCommand returns group command for file.
func NewCommand() cli.Command {
	return cli.Command{
		Name:        "file",
		Aliases:     []string{"f"},
		Usage:       "Files are the lowest level data object in Pachyderm.",
		Description: desc,
		Subcommands: []cli.Command{
			newListCommand(),
			newInspectCommand(),
			newPutCommand(),
			newGetCommand(),
			newDeleteCommand(),
		},
	}
}

var desc = `Files are the lowest level data object in Pachyderm.

  Files can be written to started (but not finished) commits with put-file.
  Files can be read from finished commits with get-file.`
