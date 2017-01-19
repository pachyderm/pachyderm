package misc

import (
	"bufio"
	"fmt"
	"os"

	"github.com/pachyderm/pachyderm/src/client"

	"github.com/urfave/cli"
)

func newDeleteAllCommand() cli.Command {
	return cli.Command{
		Name:        "delete-all",
		Usage:       "Delete everything.",
		Description: descDeleteAll,
		Action:      actDeleteAll,
	}
}

var descDeleteAll = `Delete all repos, commits, files, pipelines and jobs.

   This resets the cluster to its initial state.`

func actDeleteAll(c *cli.Context) error {
	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "user")
	if err != nil {
		return err
	}
	fmt.Printf("Are you sure you want to delete all repos, commits, files, pipelines and jobs? yN\n")
	r := bufio.NewReader(os.Stdin)
	bytes, err := r.ReadBytes('\n')
	if err != nil {
		return err
	}
	if bytes[0] == 'y' || bytes[0] == 'Y' {
		return clnt.DeleteAll()
	}
	return nil
}
