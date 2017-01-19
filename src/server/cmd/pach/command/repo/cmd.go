package repo

import "github.com/urfave/cli"

// NewCommand returns group commands for repo.
func NewCommand() cli.Command {
	return cli.Command{
		Name:    "repo",
		Aliases: []string{"r"},
		Usage:   "Repos, short for repository, are the top level data object in Pachyderm.",
		Subcommands: []cli.Command{
			newListCommand(),
			newInspectCommand(),
			newCreateCommand(),
			newDeleteCommand(),
		},
	}
}
