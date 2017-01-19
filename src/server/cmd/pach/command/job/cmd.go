package job

import "github.com/urfave/cli"

// NewCommand returns group commands for jobs.
func NewCommand() cli.Command {
	return cli.Command{
		Name:        "job",
		Aliases:     []string{"j"},
		Usage:       "Docs for jobs.",
		Description: desc,
		Subcommands: []cli.Command{
			newListCommand(),
			newInspectCommand(),
			newCreateCommand(),
			newDeleteCommand(),
			newLogCommand(),
		},
	}
}

var desc = `Jobs are the basic unit of computation in Pachyderm.

   Jobs run a containerized workload over a set of finished input commits.
   Creating a job will also create a new repo and a commit in that repo which
   contains the output of the job. Unless the job is created with another job as a
   parent. If the job is created with a parent it will use the same repo as its
   parent job and the commit it creates will use the parent job's commit as a
   parent.
   If the job fails the commit it creates will not be finished.
   The increase the throughput of a job increase the Shard paremeter.`
