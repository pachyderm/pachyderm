package commit

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client"

	"github.com/urfave/cli"
)

func newReplayCommand() cli.Command {
	return cli.Command{
		Name:        "replay",
		Aliases:     []string{"r"},
		Usage:       "Replay a number of commits onto a branch.",
		ArgsUsage:   "repo-name commits branch",
		Description: descReplay,
		Action:      actReplay,
	}
}

var descReplay = `Replay a number of commits onto a branch

   Examples:

	 # replay unique commits on branch "foo" to branch "bar".  The common commits on
	 # these branches won't be replayed.
	 $ pachctl replay-commit test foo bar`

func actReplay(c *cli.Context) error {
	args := c.Args()
	if len(args) < 3 {
		fmt.Println("invalid arguments")
		return nil
	}
	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "user")
	if err != nil {
		return err
	}
	commits, err := clnt.ReplayCommit(args[0], args[1:len(args)-1], args[len(args)-1])
	if err != nil {
		return err
	}
	for _, commit := range commits {
		fmt.Println(commit.ID)
	}
	return nil
}
