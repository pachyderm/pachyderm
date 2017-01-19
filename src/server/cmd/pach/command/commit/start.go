package commit

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client"

	"github.com/urfave/cli"
)

func newStartCommand() cli.Command {
	return cli.Command{
		Name:        "start",
		Aliases:     []string{"s"},
		Usage:       "Start a new commit.",
		ArgsUsage:   "repo-name [parent-commit | branch]",
		Description: descStart,
		Action:      actStart,
	}
}

var descStart = `Start a new commit with parent-commit as the parent, or start a commit on the given branch; if the branch does not exist, it will be createcmdD.

   Examples:

     # Start a commit in repo "foo" on branch "bar"
     $ pachctl start-commit foo bar

     # Start a commit with master/3 as the parent in repo foo
     $ pachctl start-commit foo master/3`

func actStart(c *cli.Context) error {
	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "user")
	if err != nil {
		return err
	}
	commit, err := clnt.StartCommit(c.Args().Get(0), c.Args().Get(1))
	if err != nil {
		return err
	}
	fmt.Println(commit.ID)
	return nil
}
