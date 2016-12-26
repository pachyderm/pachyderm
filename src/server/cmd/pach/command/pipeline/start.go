package pipeline

import (
	"github.com/pachyderm/pachyderm/src/client"
	pkgcmd "github.com/pachyderm/pachyderm/src/server/pkg/cmd"

	"github.com/urfave/cli"
)

func newStartCommand() cli.Command {
	return cli.Command{
		Name:        "start",
		Usage:       "Restart a stopped pipeline.",
		ArgsUsage:   "pipeline-name",
		Description: "Restart a stopped pipeline.",
		Aliases:     []string{"s"},
		Action:      actStart,
	}
}

func actStart(c *cli.Context) error {
	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "user")
	if err != nil {
		return err
	}
	if err := clnt.StartPipeline(c.Args().First()); err != nil {
		pkgcmd.ErrorAndExit("error from StartPipeline: %s", err.Error())
	}
	return nil
}
