package commit

import (
	"bufio"
	"fmt"
	"os"

	"github.com/pachyderm/pachyderm/src/client"

	"github.com/urfave/cli"
)

func newDeleteCommand() cli.Command {
	return cli.Command{
		Name:        "delete",
		Aliases:     []string{"d"},
		Usage:       "Delete a commit.",
		ArgsUsage:   "repo-name commit-id",
		Description: "Delete a commit.  The commit needs to be 1) open and 2) the head of a branch.",
		Action:      actDelete,
	}
}

func actDelete(c *cli.Context) error {
	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "user")
	if err != nil {
		return err
	}
	fmt.Printf("delete-commit is a beta feature; specifically, it may race with concurrent start-commit on the same branch.  Are you sure you want to proceed? yN\n")
	r := bufio.NewReader(os.Stdin)
	bytes, err := r.ReadBytes('\n')
	if err != nil {
		return err
	}
	if bytes[0] == 'y' || bytes[0] == 'Y' {
		return clnt.DeleteCommit(c.Args().Get(0), c.Args().Get(1))
	}
	return nil
}
