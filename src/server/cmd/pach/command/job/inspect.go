package job

import (
	"github.com/pachyderm/pachyderm/src/client"
	pkgcmd "github.com/pachyderm/pachyderm/src/server/pkg/cmd"
	"github.com/pachyderm/pachyderm/src/server/pps/pretty"
	"github.com/urfave/cli"
)

func newInspectCommand() cli.Command {
	return cli.Command{
		Name:    "inspect",
		Aliases: []string{"i"},
		Usage:   "Return info about a Job.",
		Action:  actInspect,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "block, b",
				Usage: "block until the job has either succeeded or failed",
			},
		},
	}
}

func actInspect(c *cli.Context) error {
	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "user")
	if err != nil {
		return err
	}
	jobInfo, err := clnt.InspectJob(c.Args().First(), c.Bool("block"))
	if err != nil {
		pkgcmd.ErrorAndExit("error from InspectJob: %s", err.Error())
	}
	if jobInfo == nil {
		pkgcmd.ErrorAndExit("job %s not found.", c.Args().First())
	}
	return pretty.PrintDetailedJobInfo(jobInfo)
}
