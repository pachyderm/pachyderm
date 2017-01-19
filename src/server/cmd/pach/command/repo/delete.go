package repo

import (
	"github.com/pachyderm/pachyderm/src/client"

	"github.com/urfave/cli"
)

func newDeleteCommand() cli.Command {
	return cli.Command{
		Name:    "delete",
		Aliases: []string{"d"},
		Usage:   "Delete a repo",
		Action:  actDelete,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "force, f",
				Usage: "remove the repo regardless of errors; use with care",
			},
		},
	}
}

func actDelete(c *cli.Context) error {
	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "user")
	if err != nil {
		return err
	}
	return clnt.DeleteRepo(c.Args().First(), c.Bool("force"))
}
