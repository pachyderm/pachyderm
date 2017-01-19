package pipeline

import (
	"github.com/pachyderm/pachyderm/src/client"
	pkgcmd "github.com/pachyderm/pachyderm/src/server/pkg/cmd"

	"github.com/urfave/cli"
)

func newDeleteCommand() cli.Command {
	return cli.Command{
		Name:        "delete",
		Aliases:     []string{"d"},
		Usage:       "Delete a pipeline.",
		ArgsUsage:   "pipeline-name",
		Description: "Delete a pipeline.",
		Action:      actDelete,
	}
}

func actDelete(c *cli.Context) error {
	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "user")
	if err != nil {
		return err
	}
	if err := clnt.DeletePipeline(c.Args().First()); err != nil {
		pkgcmd.ErrorAndExit("error from DeletePipeline: %s", err.Error())
	}
	return nil
}
