package misc

import (
	"github.com/pachyderm/pachyderm/src/client"

	"github.com/urfave/cli"
)

func newArchiveAllCommand() cli.Command {
	return cli.Command{
		Name:        "archive-all",
		Usage:       "Archives all commits in all repos.",
		Description: "Archives all commits in all repos.",
		Action:      actArchiveAll,
	}
}

func actArchiveAll(c *cli.Context) error {
	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "user")
	if err != nil {
		return err
	}
	return clnt.ArchiveAll()
}
