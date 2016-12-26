package branch

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client"

	"github.com/urfave/cli"
)

func newForkCommand() cli.Command {
	return cli.Command{
		Name:        "fork",
		Aliases:     []string{"f"},
		Usage:       "Start a new commit with a given parent on a new branch.",
		ArgsUsage:   "repo-name parent-commit branch-name",
		Description: descFork,
		Action:      actFork,
	}
}

var descFork = `Start a new commit with parent-commit as the parent, on a new branch with the name branch-name.

   Examples:

     # Start a commit in repo "test" on a new branch "bar" with foo/2 as the parent
     $ pachctl branch fork test foo/2 bar`

func actFork(c *cli.Context) error {
	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "user")
	if err != nil {
		return err
	}
	commit, err := clnt.ForkCommit(c.Args().Get(0), c.Args().Get(1), c.Args().Get(2))
	if err != nil {
		return err
	}
	fmt.Println(commit.ID)
	return nil
}
