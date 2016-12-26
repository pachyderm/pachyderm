package branch

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"

	"github.com/urfave/cli"
)

func newListCommand() cli.Command {
	return cli.Command{
		Name:        "list",
		Aliases:     []string{"l"},
		Usage:       "Return all branches on a repo.",
		ArgsUsage:   "repo-name",
		Description: "Return all branches on a repo.",
		Action:      actList,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "all, a",
				Usage: "list all branches including cancelled and archived ones",
			},
		},
	}
}

func actList(c *cli.Context) error {
	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "user")
	if err != nil {
		return err
	}
	status := pfsclient.CommitStatus_NORMAL
	if c.Bool("all") {
		status = pfsclient.CommitStatus_ALL
	}
	branches, err := clnt.ListBranch(c.Args().First(), status)
	if err != nil {
		return err
	}
	for _, branch := range branches {
		fmt.Println(branch)
	}
	return nil
}
