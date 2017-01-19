package commit

import (
	"github.com/pachyderm/pachyderm/src/client"

	"github.com/urfave/cli"
)

func newFinishCommand() cli.Command {
	return cli.Command{
		Name:        "finish",
		Aliases:     []string{"f"},
		Usage:       "Finish a started commit.",
		ArgsUsage:   "repo-name commit-id",
		Description: "Finish a started commit. Commit-id must be a writeable commit.",
		Action:      actFinish,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "cancel, c",
				Usage: "cancel the commit",
			},
		},
	}
}

func actFinish(c *cli.Context) error {
	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "user")
	if err != nil {
		return err
	}
	if c.Bool("cancel") {
		return clnt.CancelCommit(c.Args().Get(0), c.Args().Get(1))
	}
	return clnt.FinishCommit(c.Args().Get(0), c.Args().Get(1))
}
