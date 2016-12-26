package shell

import "github.com/urfave/cli"

// NewCommand returns group commands for repo.
func NewCommand(app *cli.App, opts ...Option) cli.Command {
	return cli.Command{
		Name:    "shell",
		Aliases: []string{"sh"},
		Usage:   "Enter interactive mode",
		Action: func(c *cli.Context) error {
			return newShell(app, opts...).run()
		},
	}
}
