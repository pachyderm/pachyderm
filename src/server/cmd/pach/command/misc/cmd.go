package misc

import "github.com/urfave/cli"

// NewCommand returns delete-all command.
func NewCommand() cli.Command {
	return cli.Command{
		Name:        "misc",
		Usage:       "Archives all commits in all repos.",
		Description: "Archives all commits in all repos.",
		Subcommands: []cli.Command{
			newArchiveAllCommand(),
			newDeleteAllCommand(),
			newPortForwardCommand(),
		},
	}
}
