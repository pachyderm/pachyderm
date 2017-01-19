package job

import (
	"os"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/urfave/cli"
)

func newLogCommand() cli.Command {
	return cli.Command{
		Name:        "log",
		Aliases:     []string{"d"},
		Usage:       "Return logs from a job.",
		ArgsUsage:   "get-logs job-id",
		Description: "Return logs from a job.",
		Action:      actLog,
	}
}

func actLog(c *cli.Context) error {
	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "user")
	if err != nil {
		return err
	}
	return clnt.GetLogs(c.Args().First(), os.Stdout)
}
