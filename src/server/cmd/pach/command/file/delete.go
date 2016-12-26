package file

import (
	"github.com/pachyderm/pachyderm/src/client"

	"github.com/urfave/cli"
)

func newDeleteCommand() cli.Command {
	return cli.Command{
		Name:        "delete",
		Aliases:     []string{"d"},
		Usage:       "Delete a file.",
		ArgsUsage:   "repo-name commit-id path/to/file",
		Description: "Delete a file.",
		Action:      Delete,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "force, f",
				Usage: "remove the repo regardless of errors; use with care",
			},
		},
	}
}

func Delete(c *cli.Context) error {
	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "user")
	if err != nil {
		return err
	}
	args := c.Args()
	return clnt.DeleteFile(args[0], args[1], args[2])
}
