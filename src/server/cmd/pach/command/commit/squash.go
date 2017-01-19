package commit

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client"

	"github.com/urfave/cli"
)

func newSquashCommand() cli.Command {
	return cli.Command{
		Name:        "squash",
		Aliases:     []string{"q"},
		Usage:       "Squash a number of commits into a single commit.",
		ArgsUsage:   "repo-name commits to-commit",
		Description: descSquash,
		Action:      actSquash,
	}
}

var descSquash = `Squash a number of commits into a single commit.

   Examples:

     # squash commits foo/2 and foo/3 into bar/1 in repo "test"
     # note that bar/1 needs to be an open commit
     $ pachctl squash-commit test foo/2 foo/3 bar/1`

func actSquash(c *cli.Context) error {
	if c.NArg() < 3 {
		return fmt.Errorf("invalid arguments")
	}
	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "user")
	if err != nil {
		return err
	}
	args := c.Args()
	return clnt.SquashCommit(args[0], args[1:len(args)-1], args[len(args)-1])
}
