package job

import (
	"github.com/pachyderm/pachyderm/src/client"
	pkgcmd "github.com/pachyderm/pachyderm/src/server/pkg/cmd"
	"github.com/urfave/cli"
)

func newDeleteCommand() cli.Command {
	return cli.Command{
		Name:        "delete",
		Aliases:     []string{"d"},
		Usage:       "Delete a job.",
		ArgsUsage:   "job-id",
		Description: "Delete a job.",
		Action:      actDelete,
	}
}

func actDelete(c *cli.Context) error {
	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "user")
	if err != nil {
		return err
	}
	if err := clnt.DeleteJob(c.Args().First()); err != nil {
		pkgcmd.ErrorAndExit("error from DeleteJob: %s", err.Error())
	}
	return nil
}
