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

func newListCommand() cli.Command {
	return cli.Command{
		Name:        "list",
		Aliases:     []string{"l"},
		Usage:       "Return all commits on a set of repos.",
		Description: descList,
		Action:      actList,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "all, a",
				Usage: "list all commits including cancelled and archived ones",
			},
			cli.BoolFlag{
				Name:  "block, b",
				Usage: "block until there are new commits since the from commits",
			},
			cli.StringSliceFlag{
				Name:  "exclude, x",
				Usage: "exclude the ancestors of this commit, or exclude the commits on this branch",
			},
			cli.StringSliceFlag{
				Name:  "provenance, p",
				Usage: "list only commits with the specified 'commit's provenance, commits are specified as RepoName/CommitID",
			},
		},
	}
}

var descList = `Return all commits on a set of repos.

   Examples:

     # return commits in repo "foo" and repo "bar"
     $ pachctl commit list foo bar

     # return commits in repo "foo" on branch "master"
     $ pachctl commit list foo/master

     # return commits in repo "foo" since commit master/2
     $ pachctl commit list foo/master -e foo/master/2

     # return commits in repo "foo" that have commits
     # "bar/master/3" and "baz/master/5" as provenance
     $ pachctl commit list foo -p bar/master/3 -p baz/master/5`

func actList(c *cli.Context) error {
	include, err := cmd.ParseCommits(c.Args())
	if err != nil {
		return err
	}
	exclude, err := cmd.ParseCommits(c.StringSlice("exclude"))
	if err != nil {
		return err
	}
	provenance, err := cmd.ParseCommits(c.GlobalStringSlice("provenance"))
	if err != nil {
		return err
	}
	status := pfsclient.CommitStatus_NORMAL
	if c.Bool("all") {
		status = pfsclient.CommitStatus_ALL
	}

	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "user")
	if err != nil {
		return err
	}

	commitInfos, err := clnt.ListCommit(exclude, include, provenance, client.CommitTypeNone, status, c.Bool("block"))
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
