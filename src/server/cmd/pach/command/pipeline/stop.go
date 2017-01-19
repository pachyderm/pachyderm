package pipeline

import (
	"github.com/pachyderm/pachyderm/src/client"
	pkgcmd "github.com/pachyderm/pachyderm/src/server/pkg/cmd"

	"github.com/urfave/cli"
)

func newStopCommand() cli.Command {
	return cli.Command{
		Name:        "Stop",
		Aliases:     []string{"t"},
		ArgsUsage:   "pipeline-name",
		Usage:       "Stop a running pipeline.",
		Description: "Stop a running pipeline.",
		Action:      actStop,
	}
}

func actStop(c *cli.Context) error {
	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "user")
	if err != nil {
		return err
	}
	if err := clnt.StopPipeline(c.Args().First()); err != nil {
		pkgcmd.ErrorAndExit("error from StopPipeline: %s", err.Error())
	}
	return nil
}
