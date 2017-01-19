package commit

import (
	"os"
	"text/tabwriter"

	"github.com/pachyderm/pachyderm/src/client"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pfs/pretty"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmd"

	"github.com/urfave/cli"
)

func newFlushCommand() cli.Command {
	return cli.Command{
		Name:        "flush",
		Usage:       "Wait for all commits caused by the specified commits to finish and return them.",
		ArgsUsage:   "commit [commit ...]",
		Description: descFlush,
		Action:      actFlush,
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "repos, r",
				Usage: "Wait only for commits leading to a specific set of repos",
			},
		},
	}
}

var descFlush = `Wait for all commits caused by the specified commits to finish and return them.

   Examples:

     # return commits caused by foo/master/1 and bar/master/2
     $ pachctl commit flush foo/master/1 bar/master/2

     # return commits caused by foo/master/1 leading to repos bar and baz
     $ pachctl commit flush foo/master/1 -r bar -r baz`

func actFlush(c *cli.Context) error {
	commits, err := cmd.ParseCommits(c.Args())
	if err != nil {
		return err
	}

	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "user")
	if err != nil {
		return err
	}
	var toRepos []*pfsclient.Repo
	for _, repoName := range c.StringSlice("repos") {
		toRepos = append(toRepos, client.NewRepo(repoName))
	}

	commitInfos, err := clnt.FlushCommit(commits, toRepos)
	if err != nil {
		return err
	}

	writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
	pretty.PrintCommitInfoHeader(writer)
	for _, commitInfo := range commitInfos {
		pretty.PrintCommitInfo(writer, commitInfo)
	}
	return writer.Flush()
}
