package pipeline

import (
	"github.com/pachyderm/pachyderm/src/client"
	pkgcmd "github.com/pachyderm/pachyderm/src/server/pkg/cmd"
	"github.com/pachyderm/pachyderm/src/server/pps/pretty"

	"github.com/urfave/cli"
)

func newInspectCommand() cli.Command {
	return cli.Command{
		Name:        "inspect",
		Aliases:     []string{"i"},
		Usage:       "Return info about a pipeline.",
		ArgsUsage:   "pipeline-name",
		Description: "Return info about a pipeline.",
		Action:      actInspect,
	}
}

func actInspect(c *cli.Context) error {
	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "user")
	if err != nil {
		return err
	}
	pipelineInfo, err := clnt.InspectPipeline(c.Args().First())
	if err != nil {
		pkgcmd.ErrorAndExit("error from InspectPipeline: %s", err.Error())
	}
	if pipelineInfo == nil {
		pkgcmd.ErrorAndExit("pipeline %s not found.", c.Args().First())
	}
	return pretty.PrintDetailedPipelineInfo(pipelineInfo)
}
