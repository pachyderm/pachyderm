package repo

import (
	"github.com/pachyderm/pachyderm/src/client"

	"github.com/urfave/cli"
)

func newCreateCommand() cli.Command {
	return cli.Command{
		Name:    "create",
		Aliases: []string{"c"},
		Usage:   "Create a new repo.",
		Action:  actCreate,
	}
}

func actCreate(c *cli.Context) error {
	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "user")
	if err != nil {
		return err
	}
	return clnt.CreateRepo(c.Args().First())
}
