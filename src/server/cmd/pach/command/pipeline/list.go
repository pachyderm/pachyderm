package pipeline

import (
	"os"
	"text/tabwriter"

	"github.com/pachyderm/pachyderm/src/client"
	pkgcmd "github.com/pachyderm/pachyderm/src/server/pkg/cmd"
	"github.com/pachyderm/pachyderm/src/server/pps/pretty"

	"github.com/urfave/cli"
)

func newListCommand() cli.Command {
	return cli.Command{
		Name:        "list",
		Aliases:     []string{"l"},
		Usage:       "Return info about all pipelines.",
		Description: "Return info about all pipelines.",
		Action:      actList,
	}
}

func actList(c *cli.Context) error {
	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "user")
	if err != nil {
		return err
	}
	pipelineInfos, err := clnt.ListPipeline()
	if err != nil {
		pkgcmd.ErrorAndExit("error from ListPipeline: %s", err.Error())
	}
	writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
	pretty.PrintPipelineHeader(writer)
	for _, pipelineInfo := range pipelineInfos {
		pretty.PrintPipelineInfo(writer, pipelineInfo)
	}
	return writer.Flush()
}
