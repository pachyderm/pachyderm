package commit

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/server/pfs/pretty"

	"github.com/urfave/cli"
)

func newInspectCommand() cli.Command {
	return cli.Command{
		Name:      "inspect",
		Aliases:   []string{"i"},
		Usage:     "Return info about a commit.",
		ArgsUsage: "repo-name commit-id",
		Action:    actInspect,
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
	commitInfo, err := clnt.InspectCommit(c.Args().Get(0), c.Args().Get(1))
	if err != nil {
		return err
	}
	if commitInfo == nil {
		return fmt.Errorf("commit %s not found", c.Args().Get(1))
	}
	return pretty.PrintDetailedCommitInfo(commitInfo)
}
