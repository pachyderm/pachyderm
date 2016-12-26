package repo

import (
	"os"
	"text/tabwriter"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/server/pfs/pretty"

	"github.com/urfave/cli"
)

func newListCommand() cli.Command {
	return cli.Command{
		Name:    "list",
		Aliases: []string{"l"},
		Usage:   "Return all repos.",
		Action:  actList,
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "provenance, p",
				Usage: "List only repos with the specified repos provenance",
			},
		},
	}
}

func actList(c *cli.Context) error {
	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "user")
	if err != nil {
		return err
	}
	infos, err := clnt.ListRepo(c.StringSlice("provenance"))
	if err != nil {
		return err
	}
	writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
	pretty.PrintRepoHeader(writer)
	for _, i := range infos {
		pretty.PrintRepoInfo(writer, i)
	}
	return writer.Flush()
}
