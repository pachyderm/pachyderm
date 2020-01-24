package cmds

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	pachdclient "github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pkg/tracing/extended"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/cmd/pachctl/shell"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/pager"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/serde"
	"github.com/pachyderm/pachyderm/src/server/pkg/tabwriter"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/pps/pretty"

	prompt "github.com/c-bata/go-prompt"
	units "github.com/docker/go-units"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/gogo/protobuf/proto"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/net/context"
)

// encoder creates an encoder that writes data structures to w[0] (or os.Stdout
// if no 'w' is passed) in the serialization format 'format'. If more than one
// writer is passed, all writers after the first are silently ignored (rather
// than returning an error), and if the 'format' passed is unrecognized
// (currently, 'format' must be 'json' or 'yaml') then pachctl exits
// immediately. Ignoring errors or crashing simplifies the type signature of
// 'encoder' and allows it to be used inline.
func encoder(format string, w ...io.Writer) serde.Encoder {
	format = strings.ToLower(format)
	if format == "" {
		format = "json"
	}
	var output io.Writer = os.Stdout
	if len(w) > 0 {
		output = w[0]
	}
	e, err := serde.GetEncoder(format, output,
		serde.WithIndent(2),
		serde.WithOrigName(true),
	)
	if err != nil {
		cmdutil.ErrorAndExit(err.Error())
	}
	return e
}

// Cmds returns a slice containing pps commands.
func Cmds() []*cobra.Command {
	var commands []*cobra.Command

	raw := false
	var output string
	outputFlags := pflag.NewFlagSet("", pflag.ExitOnError)
	outputFlags.BoolVar(&raw, "raw", false, "Disable pretty printing; serialize data structures to an encoding such as json or yaml")
	// --output is empty by default, so that we can print an error if a user
	// explicitly sets --output without --raw, but the effective default is set in
	// encode(), which assumes "json" if 'format' is empty.
	// Note: because of how spf13/flags works, no other StringVarP that sets
	// 'output' can have a default value either
	outputFlags.StringVarP(&output, "output", "o", "", "Output format when --raw is set: \"json\" or \"yaml\" (default \"json\")")

	fullTimestamps := false
	fullTimestampsFlags := pflag.NewFlagSet("", pflag.ContinueOnError)
	fullTimestampsFlags.BoolVar(&fullTimestamps, "full-timestamps", false, "Return absolute timestamps (as opposed to the default, relative timestamps).")

	noPager := false
	noPagerFlags := pflag.NewFlagSet("", pflag.ContinueOnError)
	noPagerFlags.BoolVar(&noPager, "no-pager", false, "Don't pipe output into a pager (i.e. less).")

	jobDocs := &cobra.Command{
		Short: "Docs for jobs.",
		Long: `Jobs are the basic units of computation in Pachyderm.

Jobs run a containerized workload over a set of finished input commits. Jobs are
created by pipelines and will write output to a commit in the pipeline's output
repo. A job can have multiple datums, each processed independently and the
results will be merged together at the end.

If the job fails, the output commit will not be populated with data.`,
	}
	commands = append(commands, cmdutil.CreateDocsAlias(jobDocs, "job", " job$"))

	var block bool
	inspectJob := &cobra.Command{
		Use:   "{{alias}} <job>",
		Short: "Return info about a job.",
		Long:  "Return info about a job.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer client.Close()
			jobInfo, err := client.InspectJob(args[0], block)
			if err != nil {
				cmdutil.ErrorAndExit("error from InspectJob: %s", err.Error())
			}
			if jobInfo == nil {
				cmdutil.ErrorAndExit("job %s not found.", args[0])
			}
			if raw {
				return encoder(output).EncodeProto(jobInfo)
			} else if output != "" {
				cmdutil.ErrorAndExit("cannot set --output (-o) without --raw")
			}
			ji := &pretty.PrintableJobInfo{
				JobInfo:        jobInfo,
				FullTimestamps: fullTimestamps,
			}
			return pretty.PrintDetailedJobInfo(ji)
		}),
	}
	inspectJob.Flags().BoolVarP(&block, "block", "b", false, "block until the job has either succeeded or failed")
	inspectJob.Flags().AddFlagSet(outputFlags)
	inspectJob.Flags().AddFlagSet(fullTimestampsFlags)
	shell.RegisterCompletionFunc(inspectJob, shell.JobCompletion)
	commands = append(commands, cmdutil.CreateAlias(inspectJob, "inspect job"))

	var pipelineName string
	var outputCommitStr string
	var inputCommitStrs []string
	var history string
	listJob := &cobra.Command{
		Short: "Return info about jobs.",
		Long:  "Return info about jobs.",
		Example: `
# Return all jobs
$ {{alias}}

# Return all jobs from the most recent version of pipeline "foo"
$ {{alias}} -p foo

# Return all jobs from all versions of pipeline "foo"
$ {{alias}} -p foo --history all

# Return all jobs whose input commits include foo@XXX and bar@YYY
$ {{alias}} -i foo@XXX -i bar@YYY

# Return all jobs in pipeline foo and whose input commits include bar@YYY
$ {{alias}} -p foo -i bar@YYY`,
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			commits, err := cmdutil.ParseCommits(inputCommitStrs)
			if err != nil {
				return err
			}
			history, err := cmdutil.ParseHistory(history)
			if err != nil {
				return fmt.Errorf("error parsing history flag: %v", err)
			}
			var outputCommit *pfs.Commit
			if outputCommitStr != "" {
				outputCommit, err = cmdutil.ParseCommit(outputCommitStr)
				if err != nil {
					return err
				}
			}

			client, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer client.Close()

			return pager.Page(noPager, os.Stdout, func(w io.Writer) error {
				if raw {
					e := encoder(output)
					return client.ListJobF(pipelineName, commits, outputCommit, history, true, func(ji *ppsclient.JobInfo) error {
						return e.EncodeProto(ji)
					})
				} else if output != "" {
					cmdutil.ErrorAndExit("cannot set --output (-o) without --raw")
				}
				writer := tabwriter.NewWriter(w, pretty.JobHeader)
				if err := client.ListJobF(pipelineName, commits, outputCommit, history, false, func(ji *ppsclient.JobInfo) error {
					pretty.PrintJobInfo(writer, ji, fullTimestamps)
					return nil
				}); err != nil {
					return err
				}
				return writer.Flush()
			})
		}),
	}
	listJob.Flags().StringVarP(&pipelineName, "pipeline", "p", "", "Limit to jobs made by pipeline.")
	listJob.MarkFlagCustom("pipeline", "__pachctl_get_pipeline")
	listJob.Flags().StringVarP(&outputCommitStr, "output", "o", "", "List jobs with a specific output commit. format: <repo>@<branch-or-commit>")
	listJob.MarkFlagCustom("output", "__pachctl_get_repo_commit")
	listJob.Flags().StringSliceVarP(&inputCommitStrs, "input", "i", []string{}, "List jobs with a specific set of input commits. format: <repo>@<branch-or-commit>")
	listJob.MarkFlagCustom("input", "__pachctl_get_repo_commit")
	listJob.Flags().AddFlagSet(outputFlags)
	listJob.Flags().AddFlagSet(fullTimestampsFlags)
	listJob.Flags().AddFlagSet(noPagerFlags)
	listJob.Flags().StringVar(&history, "history", "none", "Return jobs from historical versions of pipelines.")
	shell.RegisterCompletionFunc(listJob, func(flag, text string, maxCompletions int64) []prompt.Suggest {
		if flag == "-p" || flag == "--pipeline" {
			return shell.PipelineCompletion(flag, text, maxCompletions)
		}
		return nil
	})
	commands = append(commands, cmdutil.CreateAlias(listJob, "list job"))

	var pipelines cmdutil.RepeatedStringArg
	flushJob := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit> ...",
		Short: "Wait for all jobs caused by the specified commits to finish and return them.",
		Long:  "Wait for all jobs caused by the specified commits to finish and return them.",
		Example: `
# Return jobs caused by foo@XXX and bar@YYY.
$ {{alias}} foo@XXX bar@YYY

# Return jobs caused by foo@XXX leading to pipelines bar and baz.
$ {{alias}} foo@XXX -p bar -p baz`,
		Run: cmdutil.Run(func(args []string) error {
			commits, err := cmdutil.ParseCommits(args)
			if err != nil {
				return err
			}

			c, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer c.Close()

			jobInfos, err := c.FlushJobAll(commits, pipelines)
			if err != nil {
				return err
			}

			if raw {
				e := encoder(output)
				for _, jobInfo := range jobInfos {
					if err := e.EncodeProto(jobInfo); err != nil {
						return err
					}
				}
				return nil
			} else if output != "" {
				cmdutil.ErrorAndExit("cannot set --output (-o) without --raw")
			}
			writer := tabwriter.NewWriter(os.Stdout, pretty.JobHeader)
			for _, jobInfo := range jobInfos {
				pretty.PrintJobInfo(writer, jobInfo, fullTimestamps)
			}

			return writer.Flush()
		}),
	}
	flushJob.Flags().VarP(&pipelines, "pipeline", "p", "Wait only for jobs leading to a specific set of pipelines")
	flushJob.MarkFlagCustom("pipeline", "__pachctl_get_pipeline")
	flushJob.Flags().AddFlagSet(outputFlags)
	flushJob.Flags().AddFlagSet(fullTimestampsFlags)
	shell.RegisterCompletionFunc(flushJob, func(flag, text string, maxCompletions int64) []prompt.Suggest {
		if flag == "--pipeline" || flag == "-p" {
			return shell.PipelineCompletion(flag, text, maxCompletions)
		}
		return shell.BranchCompletion(flag, text, maxCompletions)
	})
	commands = append(commands, cmdutil.CreateAlias(flushJob, "flush job"))

	deleteJob := &cobra.Command{
		Use:   "{{alias}} <job>",
		Short: "Delete a job.",
		Long:  "Delete a job.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer client.Close()
			if err := client.DeleteJob(args[0]); err != nil {
				cmdutil.ErrorAndExit("error from DeleteJob: %s", err.Error())
			}
			return nil
		}),
	}
	shell.RegisterCompletionFunc(deleteJob, shell.JobCompletion)
	commands = append(commands, cmdutil.CreateAlias(deleteJob, "delete job"))

	stopJob := &cobra.Command{
		Use:   "{{alias}} <job>",
		Short: "Stop a job.",
		Long:  "Stop a job.  The job will be stopped immediately.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer client.Close()
			if err := client.StopJob(args[0]); err != nil {
				cmdutil.ErrorAndExit("error from StopJob: %s", err.Error())
			}
			return nil
		}),
	}
	shell.RegisterCompletionFunc(stopJob, shell.JobCompletion)
	commands = append(commands, cmdutil.CreateAlias(stopJob, "stop job"))

	datumDocs := &cobra.Command{
		Short: "Docs for datums.",
		Long: `Datums are the small independent units of processing for Pachyderm jobs.

A datum is defined by applying a glob pattern (in the pipeline spec) to the file
paths in the input repo. A datum can include one or more files or directories.

Datums within a job will be processed independently, sometimes distributed
across separate workers.  A separate execution of user code will be run for
each datum.`,
	}
	commands = append(commands, cmdutil.CreateDocsAlias(datumDocs, "datum", " datum$"))

	restartDatum := &cobra.Command{
		Use:   "{{alias}} <job> <datum-path1>,<datum-path2>,...",
		Short: "Restart a datum.",
		Long:  "Restart a datum.",
		Run: cmdutil.RunFixedArgs(2, func(args []string) error {
			client, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer client.Close()
			datumFilter := strings.Split(args[1], ",")
			for i := 0; i < len(datumFilter); {
				if len(datumFilter[i]) == 0 {
					if i+1 < len(datumFilter) {
						copy(datumFilter[i:], datumFilter[i+1:])
					}
					datumFilter = datumFilter[:len(datumFilter)-1]
				} else {
					i++
				}
			}
			return client.RestartDatum(args[0], datumFilter)
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(restartDatum, "restart datum"))

	var pageSize int64
	var page int64
	listDatum := &cobra.Command{
		Use:   "{{alias}} <job>",
		Short: "Return the datums in a job.",
		Long:  "Return the datums in a job.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer client.Close()
			if pageSize < 0 {
				return fmt.Errorf("pageSize must be zero or positive")
			}
			if page < 0 {
				return fmt.Errorf("page must be zero or positive")
			}
			if raw {
				e := encoder(output)
				return client.ListDatumF(args[0], pageSize, page, func(di *ppsclient.DatumInfo) error {
					return e.EncodeProto(di)
				})
			} else if output != "" {
				cmdutil.ErrorAndExit("cannot set --output (-o) without --raw")
			}
			writer := tabwriter.NewWriter(os.Stdout, pretty.DatumHeader)
			if err := client.ListDatumF(args[0], pageSize, page, func(di *ppsclient.DatumInfo) error {
				pretty.PrintDatumInfo(writer, di)
				return nil
			}); err != nil {
				return err
			}
			return writer.Flush()
		}),
	}
	listDatum.Flags().Int64Var(&pageSize, "pageSize", 0, "Specify the number of results sent back in a single page")
	listDatum.Flags().Int64Var(&page, "page", 0, "Specify the page of results to send")
	listDatum.Flags().AddFlagSet(outputFlags)
	shell.RegisterCompletionFunc(listDatum, shell.JobCompletion)
	commands = append(commands, cmdutil.CreateAlias(listDatum, "list datum"))

	inspectDatum := &cobra.Command{
		Use:   "{{alias}} <job> <datum>",
		Short: "Display detailed info about a single datum.",
		Long:  "Display detailed info about a single datum. Requires the pipeline to have stats enabled.",
		Run: cmdutil.RunFixedArgs(2, func(args []string) error {
			client, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer client.Close()
			datumInfo, err := client.InspectDatum(args[0], args[1])
			if err != nil {
				return err
			}
			if raw {
				return encoder(output).EncodeProto(datumInfo)
			} else if output != "" {
				cmdutil.ErrorAndExit("cannot set --output (-o) without --raw")
			}
			pretty.PrintDetailedDatumInfo(os.Stdout, datumInfo)
			return nil
		}),
	}
	inspectDatum.Flags().AddFlagSet(outputFlags)
	commands = append(commands, cmdutil.CreateAlias(inspectDatum, "inspect datum"))

	var (
		jobID       string
		datumID     string
		commaInputs string // comma-separated list of input files of interest
		master      bool
		follow      bool
		tail        int64
	)
	getLogs := &cobra.Command{
		Use:   "{{alias}} [--pipeline=<pipeline>|--job=<job>] [--datum=<datum>]",
		Short: "Return logs from a job.",
		Long:  "Return logs from a job.",
		Example: `
# Return logs emitted by recent jobs in the "filter" pipeline
$ {{alias}} --pipeline=filter

# Return logs emitted by the job aedfa12aedf
$ {{alias}} --job=aedfa12aedf

# Return logs emitted by the pipeline \"filter\" while processing /apple.txt and a file with the hash 123aef
$ {{alias}} --pipeline=filter --inputs=/apple.txt,123aef`,
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			client, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return fmt.Errorf("error connecting to pachd: %v", err)
			}
			defer client.Close()

			// Break up comma-separated input paths, and filter out empty entries
			data := strings.Split(commaInputs, ",")
			for i := 0; i < len(data); {
				if len(data[i]) == 0 {
					if i+1 < len(data) {
						copy(data[i:], data[i+1:])
					}
					data = data[:len(data)-1]
				} else {
					i++
				}
			}

			// Issue RPC
			iter := client.GetLogs(pipelineName, jobID, data, datumID, master, follow, tail)
			var buf bytes.Buffer
			encoder := json.NewEncoder(&buf)
			for iter.Next() {
				if raw {
					buf.Reset()
					if err := encoder.Encode(iter.Message()); err != nil {
						fmt.Fprintf(os.Stderr, "error marshalling \"%v\": %s\n", iter.Message(), err)
					}
					fmt.Println(buf.String())
				} else if iter.Message().User {
					fmt.Println(iter.Message().Message)
				} else if iter.Message().Master && master {
					fmt.Println(iter.Message().Message)
				} else if pipelineName == "" && jobID == "" {
					fmt.Println(iter.Message().Message)
				}
			}
			return iter.Err()
		}),
	}
	getLogs.Flags().StringVarP(&pipelineName, "pipeline", "p", "", "Filter the log "+
		"for lines from this pipeline (accepts pipeline name)")
	getLogs.MarkFlagCustom("pipeline", "__pachctl_get_pipeline")
	getLogs.Flags().StringVarP(&jobID, "job", "j", "", "Filter for log lines from "+
		"this job (accepts job ID)")
	getLogs.MarkFlagCustom("job", "__pachctl_get_job")
	getLogs.Flags().StringVar(&datumID, "datum", "", "Filter for log lines for this datum (accepts datum ID)")
	getLogs.Flags().StringVar(&commaInputs, "inputs", "", "Filter for log lines "+
		"generated while processing these files (accepts PFS paths or file hashes)")
	getLogs.Flags().BoolVar(&master, "master", false, "Return log messages from the master process (pipeline must be set).")
	getLogs.Flags().BoolVar(&raw, "raw", false, "Return log messages verbatim from server.")
	getLogs.Flags().BoolVarP(&follow, "follow", "f", false, "Follow logs as more are created.")
	getLogs.Flags().Int64VarP(&tail, "tail", "t", 0, "Lines of recent logs to display.")
	shell.RegisterCompletionFunc(getLogs, func(flag, text string, maxCompletions int64) []prompt.Suggest {
		if flag == "--pipeline" || flag == "-p" {
			return shell.PipelineCompletion(flag, text, maxCompletions)
		}
		if flag == "--job" || flag == "-j" {
			return shell.JobCompletion(flag, text, maxCompletions)
		}
		return nil
	})
	commands = append(commands, cmdutil.CreateAlias(getLogs, "logs"))

	pipelineDocs := &cobra.Command{
		Short: "Docs for pipelines.",
		Long: `Pipelines are a powerful abstraction for automating jobs.

Pipelines take a set of repos and branches as inputs and will write to a single
output repo of the same name. Pipelines then subscribe to commits on those repos
and launch a job to process each incoming commit.

All jobs created by a pipeline will create commits in the pipeline's output repo.`,
	}
	commands = append(commands, cmdutil.CreateDocsAlias(pipelineDocs, "pipeline", " pipeline$"))

	var build bool
	var pushImages bool
	var registry string
	var username string
	var pipelinePath string
	createPipeline := &cobra.Command{
		Short: "Create a new pipeline.",
		Long:  "Create a new pipeline from a pipeline specification. For details on the format, see http://docs.pachyderm.io/en/latest/reference/pipeline_spec.html.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) (retErr error) {
			return pipelineHelper(false, build, pushImages, registry, username, pipelinePath, false)
		}),
	}
	createPipeline.Flags().StringVarP(&pipelinePath, "file", "f", "-", "The JSON file containing the pipeline, it can be a url or local file. - reads from stdin.")
	createPipeline.Flags().BoolVarP(&build, "build", "b", false, "If true, build and push local docker images into the docker registry.")
	createPipeline.Flags().BoolVarP(&pushImages, "push-images", "p", false, "If true, push local docker images into the docker registry.")
	createPipeline.Flags().StringVarP(&registry, "registry", "r", "index.docker.io", "The registry to push images to.")
	createPipeline.Flags().StringVarP(&username, "username", "u", "", "The username to push images as.")
	commands = append(commands, cmdutil.CreateAlias(createPipeline, "create pipeline"))

	var reprocess bool
	updatePipeline := &cobra.Command{
		Short: "Update an existing Pachyderm pipeline.",
		Long:  "Update a Pachyderm pipeline with a new pipeline specification. For details on the format, see http://docs.pachyderm.io/en/latest/reference/pipeline_spec.html.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) (retErr error) {
			return pipelineHelper(reprocess, build, pushImages, registry, username, pipelinePath, true)
		}),
	}
	updatePipeline.Flags().StringVarP(&pipelinePath, "file", "f", "-", "The JSON file containing the pipeline, it can be a url or local file. - reads from stdin.")
	updatePipeline.Flags().BoolVarP(&build, "build", "b", false, "If true, build and push local docker images into the docker registry.")
	updatePipeline.Flags().BoolVarP(&pushImages, "push-images", "p", false, "If true, push local docker images into the docker registry.")
	updatePipeline.Flags().StringVarP(&registry, "registry", "r", "index.docker.io", "The registry to push images to.")
	updatePipeline.Flags().StringVarP(&username, "username", "u", "", "The username to push images as.")
	updatePipeline.Flags().BoolVar(&reprocess, "reprocess", false, "If true, reprocess datums that were already processed by previous version of the pipeline.")
	commands = append(commands, cmdutil.CreateAlias(updatePipeline, "update pipeline"))

	runPipeline := &cobra.Command{
		Use:   "{{alias}} <pipeline> [<repo>@<branch>[=<commit>]...]",
		Short: "Run an existing Pachyderm pipeline on the specified commits-branch pairs.",
		Long:  "Run a Pachyderm pipeline on the datums from specific commit-branch pairs. If only the branch is given, the head commit of the branch is used to complete the pair. Note: pipelines run automatically when data is committed to them. This command is for the case where you want to run the pipeline on a specific set of data, or if you want to rerun the pipeline. If a commit or branch is not specified, it will default to using the HEAD of master.",
		Example: `
		# Rerun the latest job for the "filter" pipeline
		$ {{alias}} filter

		# Process the pipeline "filter" on the data from commit-branch pairs "repo1@A=a23e4" and "repo2@B=bf363"
		$ {{alias}} filter repo1@A=a23e4 repo2@B=bf363

		# Run the pipeline "filter" on the data from commit "167af5" on the "staging" branch on repo "repo1"
		$ {{alias}} filter repo1@staging=167af5`,
		Run: cmdutil.RunMinimumArgs(1, func(args []string) (retErr error) {
			client, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer client.Close()
			prov, err := cmdutil.ParseCommitProvenances(args[1:])
			if err != nil {
				return err
			}
			err = client.RunPipeline(args[0], prov, jobID)
			if err != nil {
				return err
			}
			return nil
		}),
	}
	runPipeline.Flags().StringVar(&jobID, "job", "", "rerun the given job")
	commands = append(commands, cmdutil.CreateAlias(runPipeline, "run pipeline"))

	runCron := &cobra.Command{
		Use:   "{{alias}} <pipeline>",
		Short: "Run an existing Pachyderm cron pipeline now",
		Long:  "Run an existing Pachyderm cron pipeline now",
		Example: `
		# Run a cron pipeline "clock" now
		$ {{alias}} clock`,
		Run: cmdutil.RunMinimumArgs(1, func(args []string) (retErr error) {
			client, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer client.Close()
			err = client.RunCron(args[0])
			if err != nil {
				return err
			}
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(runCron, "run cron"))

	inspectPipeline := &cobra.Command{
		Use:   "{{alias}} <pipeline>",
		Short: "Return info about a pipeline.",
		Long:  "Return info about a pipeline.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer client.Close()
			pipelineInfo, err := client.InspectPipeline(args[0])
			if err != nil {
				return err
			}
			if pipelineInfo == nil {
				return fmt.Errorf("pipeline %s not found", args[0])
			}
			if raw {
				return encoder(output).EncodeProto(pipelineInfo)
			} else if output != "" {
				cmdutil.ErrorAndExit("cannot set --output (-o) without --raw")
			}
			pi := &pretty.PrintablePipelineInfo{
				PipelineInfo:   pipelineInfo,
				FullTimestamps: fullTimestamps,
			}
			return pretty.PrintDetailedPipelineInfo(pi)
		}),
	}
	inspectPipeline.Flags().AddFlagSet(outputFlags)
	inspectPipeline.Flags().AddFlagSet(fullTimestampsFlags)
	commands = append(commands, cmdutil.CreateAlias(inspectPipeline, "inspect pipeline"))

	extractPipeline := &cobra.Command{
		Use:   "{{alias}} <pipeline>",
		Short: "Return the manifest used to create a pipeline.",
		Long:  "Return the manifest used to create a pipeline.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer client.Close()
			createPipelineRequest, err := client.ExtractPipeline(args[0])
			if err != nil {
				return err
			}
			return encoder(output).EncodeProto(createPipelineRequest)
		}),
	}
	extractPipeline.Flags().StringVarP(&output, "output", "o", "", "Output format: \"json\" or \"yaml\" (default \"json\")")
	commands = append(commands, cmdutil.CreateAlias(extractPipeline, "extract pipeline"))

	var editor string
	editPipeline := &cobra.Command{
		Use:   "{{alias}} <pipeline>",
		Short: "Edit the manifest for a pipeline in your text editor.",
		Long:  "Edit the manifest for a pipeline in your text editor.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) (retErr error) {
			client, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer client.Close()
			createPipelineRequest, err := client.ExtractPipeline(args[0])
			if err != nil {
				return err
			}
			f, err := ioutil.TempFile("", args[0])
			if err != nil {
				return err
			}
			if err := encoder(output, f).EncodeProto(createPipelineRequest); err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil && retErr == nil {
					retErr = err
				}
			}()
			if editor == "" {
				editor = os.Getenv("EDITOR")
			}
			if editor == "" {
				editor = "vim"
			}
			if err := cmdutil.RunIO(cmdutil.IO{
				Stdin:  os.Stdin,
				Stdout: os.Stdout,
				Stderr: os.Stderr,
			}, editor, f.Name()); err != nil {
				return err
			}
			pipelineReader, err := ppsutil.NewPipelineManifestReader(f.Name())
			if err != nil {
				return err
			}
			request, err := pipelineReader.NextCreatePipelineRequest()
			if err != nil {
				return err
			}
			if proto.Equal(createPipelineRequest, request) {
				fmt.Println("Pipeline unchanged, no update will be performed.")
				return nil
			}
			request.Update = true
			request.Reprocess = reprocess
			if _, err := client.PpsAPIClient.CreatePipeline(
				client.Ctx(),
				request,
			); err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			return nil
		}),
	}
	editPipeline.Flags().BoolVar(&reprocess, "reprocess", false, "If true, reprocess datums that were already processed by previous version of the pipeline.")
	editPipeline.Flags().StringVar(&editor, "editor", "", "Editor to use for modifying the manifest.")
	editPipeline.Flags().StringVarP(&output, "output", "o", "", "Output format: \"json\" or \"yaml\" (default \"json\")")
	commands = append(commands, cmdutil.CreateAlias(editPipeline, "edit pipeline"))

	var spec bool
	listPipeline := &cobra.Command{
		Use:   "{{alias}} [<pipeline>]",
		Short: "Return info about all pipelines.",
		Long:  "Return info about all pipelines.",
		Run: cmdutil.RunBoundedArgs(0, 1, func(args []string) error {
			// validate flags
			if raw && spec {
				return fmt.Errorf("cannot set both --raw and --spec")
			} else if !raw && !spec && output != "" {
				cmdutil.ErrorAndExit("cannot set --output (-o) without --raw or --spec")
			}
			history, err := cmdutil.ParseHistory(history)
			if err != nil {
				return fmt.Errorf("error parsing history flag: %v", err)
			}

			// init client & get pipeline info
			client, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return fmt.Errorf("error connecting to pachd: %v", err)
			}
			defer client.Close()
			var pipeline string
			if len(args) > 0 {
				pipeline = args[0]
			}
			pipelineInfos, err := client.ListPipelineHistory(pipeline, history)
			if err != nil {
				return err
			}
			if raw {
				e := encoder(output)
				for _, pipelineInfo := range pipelineInfos {
					if err := e.EncodeProto(pipelineInfo); err != nil {
						return err
					}
				}
				return nil
			} else if spec {
				e := encoder(output)
				for _, pipelineInfo := range pipelineInfos {
					if err := e.EncodeProto(ppsutil.PipelineReqFromInfo(pipelineInfo)); err != nil {
						return err
					}
				}
				return nil
			}
			writer := tabwriter.NewWriter(os.Stdout, pretty.PipelineHeader)
			for _, pipelineInfo := range pipelineInfos {
				pretty.PrintPipelineInfo(writer, pipelineInfo, fullTimestamps)
			}
			return writer.Flush()
		}),
	}
	listPipeline.Flags().BoolVarP(&spec, "spec", "s", false, "Output 'create pipeline' compatibility specs.")
	listPipeline.Flags().AddFlagSet(outputFlags)
	listPipeline.Flags().AddFlagSet(fullTimestampsFlags)
	listPipeline.Flags().StringVar(&history, "history", "none", "Return revision history for pipelines.")
	commands = append(commands, cmdutil.CreateAlias(listPipeline, "list pipeline"))

	var all bool
	var force bool
	deletePipeline := &cobra.Command{
		Use:   "{{alias}} (<pipeline>|--all)",
		Short: "Delete a pipeline.",
		Long:  "Delete a pipeline.",
		Run: cmdutil.RunBoundedArgs(0, 1, func(args []string) error {
			client, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer client.Close()
			if len(args) > 0 && all {
				return fmt.Errorf("cannot use the --all flag with an argument")
			}
			if len(args) == 0 && !all {
				return fmt.Errorf("either a pipeline name or the --all flag needs to be provided")
			}
			if all {
				_, err = client.PpsAPIClient.DeletePipeline(
					client.Ctx(),
					&ppsclient.DeletePipelineRequest{
						All:   all,
						Force: force,
					})
			} else {
				err = client.DeletePipeline(args[0], force)
			}
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			return nil
		}),
	}
	deletePipeline.Flags().BoolVar(&all, "all", false, "delete all pipelines")
	deletePipeline.Flags().BoolVarP(&force, "force", "f", false, "delete the pipeline regardless of errors; use with care")
	commands = append(commands, cmdutil.CreateAlias(deletePipeline, "delete pipeline"))

	startPipeline := &cobra.Command{
		Use:   "{{alias}} <pipeline>",
		Short: "Restart a stopped pipeline.",
		Long:  "Restart a stopped pipeline.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer client.Close()
			if err := client.StartPipeline(args[0]); err != nil {
				cmdutil.ErrorAndExit("error from StartPipeline: %s", err.Error())
			}
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(startPipeline, "start pipeline"))

	stopPipeline := &cobra.Command{
		Use:   "{{alias}} <pipeline>",
		Short: "Stop a running pipeline.",
		Long:  "Stop a running pipeline.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer client.Close()
			if err := client.StopPipeline(args[0]); err != nil {
				cmdutil.ErrorAndExit("error from StopPipeline: %s", err.Error())
			}
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(stopPipeline, "stop pipeline"))

	var memory string
	garbageCollect := &cobra.Command{
		Short: "Garbage collect unused data.",
		Long: `Garbage collect unused data.

When a file/commit/repo is deleted, the data is not immediately removed from
the underlying storage system (e.g. S3) for performance and architectural
reasons.  This is similar to how when you delete a file on your computer, the
file is not necessarily wiped from disk immediately.

To actually remove the data, you will need to manually invoke garbage
collection with "pachctl garbage-collect".

Currently "pachctl garbage-collect" can only be started when there are no
pipelines running.  You also need to ensure that there's no ongoing "put file".
Garbage collection puts the cluster into a readonly mode where no new jobs can
be created and no data can be added.

Pachyderm's garbage collection uses bloom filters to index live objects. This
means that some dead objects may erronously not be deleted during garbage
collection. The probability of this happening depends on how many objects you
have; at around 10M objects it starts to become likely with the default values.
To lower Pachyderm's error rate and make garbage-collection more comprehensive,
you can increase the amount of memory used for the bloom filters with the
--memory flag. The default value is 10MB.
`,
		Run: cmdutil.RunFixedArgs(0, func(args []string) (retErr error) {
			client, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer client.Close()
			memoryBytes, err := units.RAMInBytes(memory)
			if err != nil {
				return err
			}
			return client.GarbageCollect(memoryBytes)
		}),
	}
	garbageCollect.Flags().StringVarP(&memory, "memory", "m", "0", "The amount of memory to use during garbage collection. Default is 10MB.")
	commands = append(commands, cmdutil.CreateAlias(garbageCollect, "garbage-collect"))

	return commands
}

func pipelineHelper(reprocess bool, build bool, pushImages bool, registry string, username string, pipelinePath string, update bool) error {
	pipelineReader, err := ppsutil.NewPipelineManifestReader(pipelinePath)
	if err != nil {
		return err
	}
	client, err := pachdclient.NewOnUserMachine("user")
	if err != nil {
		return fmt.Errorf("error connecting to pachd: %v", err)
	}
	defer client.Close()
	for {
		request, err := pipelineReader.NextCreatePipelineRequest()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		// Add trace if env var is set
		if ctx, ok := extended.StartAnyExtendedTrace(client.Ctx(), "/pps.API/CreatePipeline", request.Pipeline.Name); ok {
			client = client.WithCtx(ctx)
		}

		if update {
			request.Update = true
			request.Reprocess = reprocess
		}
		if build || pushImages {
			if build && pushImages {
				fmt.Fprintln(os.Stderr, "WARNING: `--push-images` is redundant, as it's already enabled with `--build`")
			}
			dockerClient, err := docker.NewClientFromEnv()
			if err != nil {
				return fmt.Errorf("could not create a docker client from the environment: %s", err)
			}
			authConfig, err := dockerConfig(registry, username)
			if err != nil {
				return err
			}
			repo, sourceTag := docker.ParseRepositoryTag(request.Transform.Image)
			if sourceTag == "" {
				sourceTag = "latest"
			}
			destTag := uuid.NewWithoutDashes()

			if build {
				url, err := url.Parse(pipelinePath)
				if pipelinePath == "-" || (err == nil && url.Scheme != "") {
					return fmt.Errorf("`--build` can only be used when the pipeline path is local")
				}
				dockerfile := request.Transform.Dockerfile
				if dockerfile == "" {
					dockerfile = "./Dockerfile"
				}
				contextDir, dockerfile := filepath.Split(dockerfile)
				err = buildImage(dockerClient, repo, contextDir, dockerfile, destTag)
				if err != nil {
					return err
				}
				// Now that we've built into `destTag`, change the
				// `sourceTag` to be the same so that the push will work with
				// the right image
				sourceTag = destTag
			}
			image, err := pushImage(dockerClient, authConfig, repo, sourceTag, destTag)
			if err != nil {
				return err
			}
			request.Transform.Image = image
		}
		if _, err := client.PpsAPIClient.CreatePipeline(
			client.Ctx(),
			request,
		); err != nil {
			return grpcutil.ScrubGRPC(err)
		}
	}
	return nil
}

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

func dockerConfig(registry string, username string) (docker.AuthConfiguration, error) {
	// try to automatically determine the credentials
	authConfigs, err := docker.NewAuthConfigurationsFromDockerCfg()
	if err == nil {
		for _, ac := range authConfigs.Configs {
			u, err := url.Parse(ac.ServerAddress)
			if err == nil && u.Hostname() == registry && (username == "" || username == ac.Username) {
				return ac, nil
			}
		}
	}

	// if that failed, manually build credentials
	if username == "" {
		// request the username if it hasn't been specified yet
		fmt.Printf("Username for %s: ", registry)
		reader := bufio.NewReader(os.Stdin)
		username, err = reader.ReadString('\n')
		if err != nil {
			return docker.AuthConfiguration{}, fmt.Errorf("could not read username: %v", err)
		}
		username = strings.TrimRight(username, "\r\n")
	}
	fmt.Printf("Password for %s@%s: ", username, registry)
	passBytes, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return docker.AuthConfiguration{}, fmt.Errorf("could not read password: %v", err)
	}

	// print a newline, since `ReadPassword` gobbles the user-inputted one
	fmt.Println()

	return docker.AuthConfiguration{
		Username: username,
		Password: string(passBytes),
	}, nil
}

// buildImage builds a new docker image.
func buildImage(client *docker.Client, repo string, contextDir string, dockerfile string, destTag string) error {
	destImage := fmt.Sprintf("%s:%s", repo, destTag)

	fmt.Printf("Building %s, this may take a while.\n", destImage)

	err := client.BuildImage(docker.BuildImageOptions{
		Name:         destImage,
		ContextDir:   contextDir,
		Dockerfile:   dockerfile,
		OutputStream: os.Stdout,
	})

	if err != nil {
		return fmt.Errorf("could not build docker image: %s", err)
	}

	return nil
}

// pushImage pushes a docker image.
func pushImage(client *docker.Client, authConfig docker.AuthConfiguration, repo string, sourceTag string, destTag string) (string, error) {
	sourceImage := fmt.Sprintf("%s:%s", repo, sourceTag)
	destImage := fmt.Sprintf("%s:%s", repo, destTag)

	fmt.Printf("Tagging/pushing %s, this may take a while.\n", destImage)

	if err := client.TagImage(sourceImage, docker.TagImageOptions{
		Repo:    repo,
		Tag:     destTag,
		Context: context.Background(),
	}); err != nil {
		err = fmt.Errorf("could not tag docker image: %s", err)
		return "", err
	}

	if err := client.PushImage(
		docker.PushImageOptions{
			Name: repo,
			Tag:  destTag,
		},
		authConfig,
	); err != nil {
		err = fmt.Errorf("could not push docker image: %s", err)
		return "", err
	}

	return destImage, nil
}
