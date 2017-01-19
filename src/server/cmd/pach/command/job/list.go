package job

import (
	"os"
	"sort"
	"text/tabwriter"

	"github.com/pachyderm/pachyderm/src/client"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	pkgcmd "github.com/pachyderm/pachyderm/src/server/pkg/cmd"
	"github.com/pachyderm/pachyderm/src/server/pps/pretty"
	"github.com/urfave/cli"
)

func newListCommand() cli.Command {
	return cli.Command{
		Name:        "list",
		Aliases:     []string{"l"},
		Usage:       "Return info about jobs.",
		ArgsUsage:   "[-p pipeline-name] [commits]",
		Description: descList,
		Action:      actList,
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "pipeline, p",
				Usage: "Limit to jobs made by pipeline.",
			},
		},
	}
}

var descList = `Return info about jobs.

   Examples:

     # return all jobs
     $ pachctl job list

     # return all jobs in pipeline foo
     $ pachctl job list -p foo

     # return all jobs whose input commits include foo/master/1 and bar/master/2
     $ pachctl job list foo/master/1 bar/master/2

     # return all jobs in pipeline foo and whose input commits include bar/master/2
     $ pachctl job list -p foo bar/master/2`

// ByCreationTime is an implementation of sort.Interface which
// sorts pps job info by creation time, ascending.
type ByCreationTime []*ppsclient.JobInfo

func (arr ByCreationTime) Len() int { return len(arr) }

func (arr ByCreationTime) Swap(i, j int) { arr[i], arr[j] = arr[j], arr[i] }

func (arr ByCreationTime) Less(i, j int) bool {
	if arr[i].Started == nil || arr[j].Started == nil {
		return false
	}
	if arr[i].Started.Seconds < arr[j].Started.Seconds {
		return true
	} else if arr[i].Started.Seconds == arr[j].Started.Seconds {
		return arr[i].Started.Nanos < arr[j].Started.Nanos
	}
	return false
}

func actList(c *cli.Context) error {
	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "user")
	if err != nil {
		return err
	}
	commits, err := pkgcmd.ParseCommits(c.Args())
	if err != nil {
		pkgcmd.ErrorAndExit("error from InspectJob: %v", sanitizeErr(err))
	}

	jobInfos, err := clnt.ListJob(c.String("pipeline"), commits)
	if err != nil {
		pkgcmd.ErrorAndExit("error from InspectJob: %v", sanitizeErr(err))
	}

	// Display newest jobs first
	sort.Sort(sort.Reverse(ByCreationTime(jobInfos)))

	writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
	pretty.PrintJobHeader(writer)
	for _, jobInfo := range jobInfos {
		pretty.PrintJobInfo(writer, jobInfo)
	}

	if err := writer.Flush(); err != nil {
		pkgcmd.ErrorAndExit("error from InspectJob: %v", sanitizeErr(err))
	}
	return nil
}
