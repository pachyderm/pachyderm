package repo

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/server/pfs/pretty"

	"github.com/urfave/cli"
)

func newInspectCommand() cli.Command {
	return cli.Command{
		Name:    "inspect",
		Aliases: []string{"i"},
		Usage:   "Return info about a repo.",
		Action:  actInspect,
	}
}

func actInspect(c *cli.Context) error {
	if c.NArg() != 1 {
		return fmt.Errorf("invalid argument")
	}
	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "user")
	if err != nil {
		return err
	}
	info, err := clnt.InspectRepo(c.Args().First())
	if err != nil {
		return err
	}
	if info == nil {
		return fmt.Errorf("repo %s not found", c.Args().First())
	}
	return pretty.PrintDetailedRepoInfo(info)
}
