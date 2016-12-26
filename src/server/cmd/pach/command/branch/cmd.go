package branch

import "github.com/urfave/cli"

// NewCommand returns group commands for repo.
func NewCommand() cli.Command {
	return cli.Command{
		Name:    "branch",
		Aliases: []string{"b"},
		Usage:   "Branches represents independent lines of commits.",
		Subcommands: []cli.Command{
			newListCommand(),
			newForkCommand(),
		},
	}
}
