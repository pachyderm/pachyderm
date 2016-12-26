package pipeline

import "github.com/urfave/cli"

// NewCommand returns group commands for repo.
func NewCommand() cli.Command {
	return cli.Command{
		Name:        "pipeline",
		Aliases:     []string{"p"},
		Usage:       "Docs for pipelines.",
		Description: desc,
		Subcommands: []cli.Command{
			newListCommand(),
			newInspectCommand(),
			newCreateCommand(),
			newUpdateCommand(),
			newDeleteCommand(),
			newStartCommand(),
			newStopCommand(),
			newRunCommand(),
		},
	}
}

var desc = `Pipelines are a powerful abstraction for automating jobs.

   Pipelines take a set of repos as inputs, rather than the set of commits that
   jobs take. Pipelines then subscribe to commits on those repos and launches a job
   to process each incoming commit.
   Creating a pipeline will also create a repo of the same name.
   All jobs created by a pipeline will create commits in the pipeline's repo.`
