package cmds

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/pachyderm/pachyderm/v2/src/client"
	pachdclient "github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/clientsdk"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachtmpl"
	"github.com/pachyderm/pachyderm/v2/src/internal/pager"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/tabwriter"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing/extended"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	ppsclient "github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/cmd/pachctl/shell"
	"github.com/pachyderm/pachyderm/v2/src/server/pps/pretty"
	txncmds "github.com/pachyderm/pachyderm/v2/src/server/transaction/cmds"

	prompt "github.com/c-bata/go-prompt"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/itchyny/gojq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	// Plural variables are used below for user convenience.
	datums    = "datums"
	jobs      = "jobs"
	pipelines = "pipelines"
	secrets   = "secrets"
)

// Cmds returns a slice containing pps commands.
func Cmds() []*cobra.Command {
	var commands []*cobra.Command

	var raw bool
	var output string
	outputFlags := cmdutil.OutputFlags(&raw, &output)

	var fullTimestamps bool
	timestampFlags := cmdutil.TimestampFlags(&fullTimestamps)

	var noPager bool
	pagerFlags := cmdutil.PagerFlags(&noPager)

	jobDocs := &cobra.Command{
		Short: "Docs for jobs.",
		Long: `Jobs are the basic units of computation in Pachyderm.

Jobs run a containerized workload over a set of finished input commits. Jobs are
created by pipelines and will write output to a commit in the pipeline's output
repo. A job can have multiple datums, each processed independently and the
results will be merged together at the end.

If the job fails, the output commit will not be populated with data.`,
	}
	commands = append(commands, cmdutil.CreateDocsAliases(jobDocs, "job", " job$", jobs))

	inspectJob := &cobra.Command{
		Use:   "{{alias}} <pipeline>@<job>",
		Short: "Return info about a job.",
		Long:  "Return info about a job.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			job, err := cmdutil.ParseJob(args[0])
			if err != nil && uuid.IsUUIDWithoutDashes(args[0]) {
				return errors.New(`Use "list job <id>" to see jobs with a given ID across different pipelines`)
			} else if err != nil {
				return err
			}
			client, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer client.Close()
			jobInfo, err := client.InspectJob(job.Pipeline.Name, job.ID, true)
			if err != nil {
				return errors.Wrap(err, "error from InspectJob")
			}
			if raw {
				return errors.EnsureStack(cmdutil.Encoder(output, os.Stdout).EncodeProto(jobInfo))
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}
			pji := &pretty.PrintableJobInfo{
				JobInfo:        jobInfo,
				FullTimestamps: fullTimestamps,
			}
			return pretty.PrintDetailedJobInfo(os.Stdout, pji)
		}),
	}
	inspectJob.Flags().AddFlagSet(outputFlags)
	inspectJob.Flags().AddFlagSet(timestampFlags)
	shell.RegisterCompletionFunc(inspectJob, shell.JobCompletion)
	commands = append(commands, cmdutil.CreateAliases(inspectJob, "inspect job", jobs))

	writeJobInfos := func(out io.Writer, jobInfos []*pps.JobInfo) error {
		if raw {
			e := cmdutil.Encoder(output, out)
			for _, jobInfo := range jobInfos {
				if err := e.EncodeProto(jobInfo); err != nil {
					return errors.EnsureStack(err)
				}
			}
			return nil
		} else if output != "" {
			return errors.New("cannot set --output (-o) without --raw")
		}

		return pager.Page(noPager, out, func(w io.Writer) error {
			writer := tabwriter.NewWriter(w, pretty.JobHeader)
			for _, jobInfo := range jobInfos {
				pretty.PrintJobInfo(writer, jobInfo, fullTimestamps)
			}
			return writer.Flush()
		})
	}

	waitJob := &cobra.Command{
		Use:   "{{alias}} <job>|<pipeline>@<job>",
		Short: "Wait for a job to finish then return info about the job.",
		Long:  "Wait for a job to finish then return info about the job.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer client.Close()

			var jobInfos []*pps.JobInfo
			if uuid.IsUUIDWithoutDashes(args[0]) {
				jobInfos, err = client.WaitJobSetAll(args[0], true)
				if err != nil {
					return err
				}
			} else {
				job, err := cmdutil.ParseJob(args[0])
				if err != nil {
					return err
				}
				jobInfo, err := client.WaitJob(job.Pipeline.Name, job.ID, true)
				if err != nil {
					return errors.Wrap(err, "error from InspectJob")
				}
				jobInfos = []*pps.JobInfo{jobInfo}
			}

			return writeJobInfos(os.Stdout, jobInfos)
		}),
	}
	waitJob.Flags().AddFlagSet(outputFlags)
	waitJob.Flags().AddFlagSet(timestampFlags)
	shell.RegisterCompletionFunc(waitJob, shell.JobCompletion)
	commands = append(commands, cmdutil.CreateAliases(waitJob, "wait job", jobs))

	var pipelineName string
	var inputCommitStrs []string
	var history string
	var stateStrs []string
	var expand bool
	listJob := &cobra.Command{
		Use:   "{{alias}} [<job-id>]",
		Short: "Return info about jobs.",
		Long:  "Return info about jobs.",
		Example: `
# Return a summary list of all jobs
$ {{alias}}

# Return all sub-jobs in a job
$ {{alias}} <job-id>

# Return all sub-jobs split across all pipelines
$ {{alias}} --expand

# Return only the sub-jobs from the most recent version of pipeline "foo"
$ {{alias}} -p foo

# Return all sub-jobs from all versions of pipeline "foo"
$ {{alias}} -p foo --history all

# Return all sub-jobs whose input commits include foo@XXX and bar@YYY
$ {{alias}} -i foo@XXX -i bar@YYY

# Return all sub-jobs in pipeline foo and whose input commits include bar@YYY
$ {{alias}} -p foo -i bar@YYY`,
		Run: cmdutil.RunBoundedArgs(0, 1, func(args []string) error {
			commits, err := cmdutil.ParseCommits(inputCommitStrs)
			if err != nil {
				return err
			}
			historyCount, err := cmdutil.ParseHistory(history)
			if err != nil {
				return errors.Wrapf(err, "error parsing history flag")
			}
			var filter string
			if len(stateStrs) > 0 {
				filter, err = ParseJobStates(stateStrs)
				if err != nil {
					return errors.Wrap(err, "error parsing state")
				}
			}

			client, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer client.Close()

			if !raw && output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}

			if len(args) == 0 {
				if pipelineName == "" && !expand {
					// We are listing jobs
					if len(stateStrs) != 0 {
						return errors.Errorf("cannot specify '--state' when listing all jobs")
					} else if len(inputCommitStrs) != 0 {
						return errors.Errorf("cannot specify '--input' when listing all jobs")
					} else if history != "none" {
						return errors.Errorf("cannot specify '--history' when listing all jobs")
					}

					listJobSetClient, err := client.PpsAPIClient.ListJobSet(client.Ctx(), &pps.ListJobSetRequest{})
					if err != nil {
						return grpcutil.ScrubGRPC(err)
					}

					if raw {
						e := cmdutil.Encoder(output, os.Stdout)
						return clientsdk.ForEachJobSet(listJobSetClient, func(jobSetInfo *pps.JobSetInfo) error {
							return errors.EnsureStack(e.EncodeProto(jobSetInfo))
						})
					}

					return pager.Page(noPager, os.Stdout, func(w io.Writer) error {
						writer := tabwriter.NewWriter(w, pretty.JobSetHeader)
						if err := clientsdk.ForEachJobSet(listJobSetClient, func(jobSetInfo *pps.JobSetInfo) error {
							pretty.PrintJobSetInfo(writer, jobSetInfo, fullTimestamps)
							return nil
						}); err != nil {
							return err
						}
						return writer.Flush()
					})
				} else {
					// We are listing all sub-jobs, possibly restricted to a single pipeline
					if raw {
						e := cmdutil.Encoder(output, os.Stdout)
						return client.ListJobFilterF(pipelineName, commits, historyCount, true, filter, func(ji *ppsclient.JobInfo) error {
							return errors.EnsureStack(e.EncodeProto(ji))
						})
					}

					return pager.Page(noPager, os.Stdout, func(w io.Writer) error {
						writer := tabwriter.NewWriter(w, pretty.JobHeader)
						if err := client.ListJobFilterF(pipelineName, commits, historyCount, false, filter, func(ji *ppsclient.JobInfo) error {
							pretty.PrintJobInfo(writer, ji, fullTimestamps)
							return nil
						}); err != nil {
							return err
						}
						return writer.Flush()
					})
				}
			} else {
				// We are listing sub-jobs of a specific job
				if len(stateStrs) != 0 {
					return errors.Errorf("cannot specify '--state' when listing sub-jobs")
				} else if len(inputCommitStrs) != 0 {
					return errors.Errorf("cannot specify '--input' when listing sub-jobs")
				} else if history != "none" {
					return errors.Errorf("cannot specify '--history' when listing sub-jobs")
				} else if pipelineName != "" {
					return errors.Errorf("cannot specify '--pipeline' when listing sub-jobs")
				}

				var jobInfos []*pps.JobInfo
				jobInfos, err = client.InspectJobSet(args[0], false)
				if err != nil {
					return errors.Wrap(err, "error from InspectJobSet")
				}

				return writeJobInfos(os.Stdout, jobInfos)
			}
		}),
	}
	listJob.Flags().StringVarP(&pipelineName, "pipeline", "p", "", "Limit to jobs made by pipeline.")
	listJob.MarkFlagCustom("pipeline", "__pachctl_get_pipeline")
	listJob.Flags().StringSliceVarP(&inputCommitStrs, "input", "i", []string{}, "List jobs with a specific set of input commits. format: <repo>@<branch-or-commit>")
	listJob.MarkFlagCustom("input", "__pachctl_get_repo_commit")
	listJob.Flags().BoolVarP(&expand, "expand", "x", false, "Show one line for each sub-job and include more columns")
	listJob.Flags().AddFlagSet(outputFlags)
	listJob.Flags().AddFlagSet(timestampFlags)
	listJob.Flags().AddFlagSet(pagerFlags)
	listJob.Flags().StringVar(&history, "history", "none", "Return jobs from historical versions of pipelines.")
	listJob.Flags().StringArrayVar(&stateStrs, "state", []string{}, "Return only sub-jobs with the specified state. Can be repeated to include multiple states")
	shell.RegisterCompletionFunc(listJob,
		func(flag, text string, maxCompletions int64) ([]prompt.Suggest, shell.CacheFunc) {
			if flag == "-p" || flag == "--pipeline" {
				cs, cf := shell.PipelineCompletion(flag, text, maxCompletions)
				return cs, shell.AndCacheFunc(cf, shell.SameFlag(flag))
			}
			return shell.JobSetCompletion(flag, text, maxCompletions)
		})
	commands = append(commands, cmdutil.CreateAliases(listJob, "list job", jobs))

	deleteJob := &cobra.Command{
		Use:   "{{alias}} <pipeline>@<job>",
		Short: "Delete a job.",
		Long:  "Delete a job.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			job, err := cmdutil.ParseJob(args[0])
			if err != nil {
				return err
			}
			client, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer client.Close()
			if err := client.DeleteJob(job.Pipeline.Name, job.ID); err != nil {
				return errors.Wrap(err, "error from DeleteJob")
			}
			return nil
		}),
	}
	shell.RegisterCompletionFunc(deleteJob, shell.JobCompletion)
	commands = append(commands, cmdutil.CreateAliases(deleteJob, "delete job", jobs))

	stopJob := &cobra.Command{
		Use:   "{{alias}} <pipeline>@<job>",
		Short: "Stop a job.",
		Long:  "Stop a job.  The job will be stopped immediately.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer client.Close()

			if uuid.IsUUIDWithoutDashes(args[0]) {
				// Stop each subjob in a transaction
				jobInfos, err := client.InspectJobSet(args[0], false)
				if err != nil {
					return err
				}
				if _, err := client.RunBatchInTransaction(func(tb *pachdclient.TransactionBuilder) error {
					for _, jobInfo := range jobInfos {
						if err := tb.StopJob(jobInfo.Job.Pipeline.Name, jobInfo.Job.ID); err != nil {
							return err
						}
					}
					return nil
				}); err != nil {
					return err
				}
			} else {
				job, err := cmdutil.ParseJob(args[0])
				if err != nil {
					return err
				}
				if err := client.StopJob(job.Pipeline.Name, job.ID); err != nil {
					return errors.Wrap(err, "error from StopJob")
				}
			}
			return nil
		}),
	}
	shell.RegisterCompletionFunc(stopJob, shell.JobCompletion)
	commands = append(commands, cmdutil.CreateAliases(stopJob, "stop job", jobs))

	datumDocs := &cobra.Command{
		Short: "Docs for datums.",
		Long: `Datums are the small independent units of processing for Pachyderm jobs.

A datum is defined by applying a glob pattern (in the pipeline spec) to the file
paths in the input repo. A datum can include one or more files or directories.

Datums within a job will be processed independently, sometimes distributed
across separate workers.  A separate execution of user code will be run for
each datum.`,
	}
	commands = append(commands, cmdutil.CreateDocsAliases(datumDocs, "datum", " datum$", datums))

	restartDatum := &cobra.Command{
		Use:   "{{alias}} <pipeline>@<job> <datum-path1>,<datum-path2>,...",
		Short: "Restart a datum.",
		Long:  "Restart a datum.",
		Run: cmdutil.RunFixedArgs(2, func(args []string) error {
			job, err := cmdutil.ParseJob(args[0])
			if err != nil {
				return err
			}
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
			return client.RestartDatum(job.Pipeline.Name, job.ID, datumFilter)
		}),
	}
	commands = append(commands, cmdutil.CreateAliases(restartDatum, "restart datum", datums))

	var pipelineInputPath string
	listDatum := &cobra.Command{
		Use:   "{{alias}} <pipeline>@<job>",
		Short: "Return the datums in a job.",
		Long:  "Return the datums in a job.",
		Run: cmdutil.RunBoundedArgs(0, 1, func(args []string) (retErr error) {
			client, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer client.Close()
			var printF func(*ppsclient.DatumInfo) error
			if !raw {
				if output != "" {
					return errors.New("cannot set --output (-o) without --raw")
				}
				writer := tabwriter.NewWriter(os.Stdout, pretty.DatumHeader)
				printF = func(di *ppsclient.DatumInfo) error {
					pretty.PrintDatumInfo(writer, di)
					return nil
				}
				defer func() {
					if err := writer.Flush(); retErr == nil {
						retErr = err
					}
				}()
			} else {
				e := cmdutil.Encoder(output, os.Stdout)
				printF = func(di *ppsclient.DatumInfo) error {
					return errors.EnsureStack(e.EncodeProto(di))
				}
			}
			if pipelineInputPath != "" && len(args) == 1 {
				return errors.Errorf("can't specify both a job and a pipeline spec")
			} else if pipelineInputPath != "" {
				pipelineBytes, err := readPipelineBytes(pipelineInputPath)
				if err != nil {
					return err
				}
				pipelineReader, err := ppsutil.NewPipelineManifestReader(pipelineBytes)
				if err != nil {
					return err
				}
				request, err := pipelineReader.NextCreatePipelineRequest()
				if err != nil {
					return err
				}
				return client.ListDatumInput(request.Input, printF)
			} else if len(args) == 1 {
				job, err := cmdutil.ParseJob(args[0])
				if err != nil {
					return err
				}
				return client.ListDatum(job.Pipeline.Name, job.ID, printF)
			} else {
				return errors.Errorf("must specify either a job or a pipeline spec")
			}
		}),
	}
	listDatum.Flags().StringVarP(&pipelineInputPath, "file", "f", "", "The JSON file containing the pipeline to list datums from, the pipeline need not exist")
	listDatum.Flags().AddFlagSet(outputFlags)
	shell.RegisterCompletionFunc(listDatum, shell.JobCompletion)
	commands = append(commands, cmdutil.CreateAliases(listDatum, "list datum", datums))

	inspectDatum := &cobra.Command{
		Use:   "{{alias}} <pipeline>@<job> <datum>",
		Short: "Display detailed info about a single datum.",
		Long:  "Display detailed info about a single datum. Requires the pipeline to have stats enabled.",
		Run: cmdutil.RunFixedArgs(2, func(args []string) error {
			job, err := cmdutil.ParseJob(args[0])
			if err != nil {
				return err
			}
			client, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer client.Close()
			datumInfo, err := client.InspectDatum(job.Pipeline.Name, job.ID, args[1])
			if err != nil {
				return err
			}
			if raw {
				return errors.EnsureStack(cmdutil.Encoder(output, os.Stdout).EncodeProto(datumInfo))
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}
			pretty.PrintDetailedDatumInfo(os.Stdout, datumInfo)
			return nil
		}),
	}
	inspectDatum.Flags().AddFlagSet(outputFlags)
	commands = append(commands, cmdutil.CreateAliases(inspectDatum, "inspect datum", datums))

	var (
		jobStr      string
		datumID     string
		commaInputs string // comma-separated list of input files of interest
		master      bool
		worker      bool
		follow      bool
		tail        int64
		since       string
	)

	// prettyLogsPrinter helps to print the logs recieved in different colours
	prettyLogsPrinter := func(message string) {
		informationArray := strings.Split(message, " ")
		if len(informationArray) > 1 {
			debugString := informationArray[1]
			debugLevel := strings.ToLower(debugString)
			var debugLevelColoredString string
			if debugLevel == "info" {
				debugLevelColoredString = color.New(color.FgGreen).Sprint(debugString)
			} else if debugLevel == "warning" {
				debugLevelColoredString = color.New(color.FgYellow).Sprint(debugString)
			} else if debugLevel == "error" {
				debugLevelColoredString = color.New(color.FgRed).Sprint(debugString)
			} else {
				debugLevelColoredString = debugString
			}
			informationArray[1] = debugLevelColoredString
			coloredMessage := strings.Join(informationArray, " ")
			fmt.Println(coloredMessage)
		} else {
			fmt.Println(message)
		}

	}

	getLogs := &cobra.Command{
		Use:   "{{alias}} [--pipeline=<pipeline>|--job=<pipeline>@<job>] [--datum=<datum>]",
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
				return errors.Wrapf(err, "error connecting to pachd")
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
			since, err := time.ParseDuration(since)
			if err != nil {
				return errors.Wrapf(err, "error parsing since(%q)", since)
			}
			if tail != 0 {
				return errors.Errorf("tail has been deprecated and removed from Pachyderm, use --since instead")
			}

			if pipelineName != "" && jobStr != "" {
				return errors.Errorf("only one of pipeline or job should be specified")
			}

			var jobID string
			if jobStr != "" {
				job, err := cmdutil.ParseJob(jobStr)
				if err != nil {
					return err
				}
				pipelineName = job.Pipeline.Name
				jobID = job.ID
			}

			// Issue RPC
			iter := client.GetLogs(pipelineName, jobID, data, datumID, master, follow, since)
			var buf bytes.Buffer
			encoder := json.NewEncoder(&buf)
			for iter.Next() {
				if raw {
					buf.Reset()
					if err := encoder.Encode(iter.Message()); err != nil {
						fmt.Fprintf(os.Stderr, "error marshalling \"%v\": %s\n", iter.Message(), err)
					}
					fmt.Println(buf.String())
				} else if iter.Message().User && !master && !worker {
					prettyLogsPrinter(iter.Message().Message)
				} else if iter.Message().Master && master {
					prettyLogsPrinter(iter.Message().Message)
				} else if !iter.Message().User && !iter.Message().Master && worker {
					prettyLogsPrinter(iter.Message().Message)
				} else if pipelineName == "" && jobID == "" {
					prettyLogsPrinter(iter.Message().Message)
				}
			}
			return iter.Err()
		}),
	}
	getLogs.Flags().StringVarP(&pipelineName, "pipeline", "p", "", "Filter the log "+
		"for lines from this pipeline (accepts pipeline name)")
	getLogs.MarkFlagCustom("pipeline", "__pachctl_get_pipeline")
	getLogs.Flags().StringVarP(&jobStr, "job", "j", "", "Filter for log lines from "+
		"this job (accepts job ID)")
	getLogs.MarkFlagCustom("job", "__pachctl_get_job")
	getLogs.Flags().StringVar(&datumID, "datum", "", "Filter for log lines for this datum (accepts datum ID)")
	getLogs.Flags().StringVar(&commaInputs, "inputs", "", "Filter for log lines "+
		"generated while processing these files (accepts PFS paths or file hashes)")
	getLogs.Flags().BoolVar(&master, "master", false, "Return log messages from the master process (pipeline must be set).")
	getLogs.Flags().BoolVar(&worker, "worker", false, "Return log messages from the worker process.")
	getLogs.Flags().BoolVar(&raw, "raw", false, "Return log messages verbatim from server.")
	getLogs.Flags().BoolVarP(&follow, "follow", "f", false, "Follow logs as more are created.")
	getLogs.Flags().Int64VarP(&tail, "tail", "t", 0, "Lines of recent logs to display.")
	getLogs.Flags().StringVar(&since, "since", "24h", "Return log messages more recent than \"since\".")
	shell.RegisterCompletionFunc(getLogs,
		func(flag, text string, maxCompletions int64) ([]prompt.Suggest, shell.CacheFunc) {
			if flag == "--pipeline" || flag == "-p" {
				cs, cf := shell.PipelineCompletion(flag, text, maxCompletions)
				return cs, shell.AndCacheFunc(cf, shell.SameFlag(flag))
			}
			if flag == "--job" || flag == "-j" {
				cs, cf := shell.JobCompletion(flag, text, maxCompletions)
				return cs, shell.AndCacheFunc(cf, shell.SameFlag(flag))
			}
			return nil, shell.SameFlag(flag)
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
	commands = append(commands, cmdutil.CreateDocsAliases(pipelineDocs, "pipeline", " pipeline$", pipelines))

	var pushImages bool
	var registry string
	var username string
	var pipelinePath string
	var jsonnetPath string
	var jsonnetArgs []string
	createPipeline := &cobra.Command{
		Short: "Create a new pipeline.",
		Long:  "Create a new pipeline from a pipeline specification. For details on the format, see https://docs.pachyderm.com/latest/reference/pipeline_spec/.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) (retErr error) {
			return pipelineHelper(false, pushImages, registry, username, pipelinePath, jsonnetPath, jsonnetArgs, false)
		}),
	}
	createPipeline.Flags().StringVarP(&pipelinePath, "file", "f", "", "A JSON file (url or filepath) containing one or more pipelines. \"-\" reads from stdin (the default behavior). Exactly one of --file and --jsonnet must be set.")
	createPipeline.Flags().StringVar(&jsonnetPath, "jsonnet", "", "BETA: A Jsonnet template file (url or filepath) for one or more pipelines. \"-\" reads from stdin. Exactly one of --file and --jsonnet must be set. Jsonnet templates must contain a top-level function; strings can be passed to this function with --arg (below)")
	createPipeline.Flags().StringArrayVar(&jsonnetArgs, "arg", nil, "Top-level argument passed to the Jsonnet template in --jsonnet (which must be set if any --arg arguments are passed). Value must be of the form 'param=value'. For multiple args, --arg may be set more than once.")
	createPipeline.Flags().BoolVarP(&pushImages, "push-images", "p", false, "If true, push local docker images into the docker registry.")
	createPipeline.Flags().StringVarP(&registry, "registry", "r", "index.docker.io", "The registry to push images to.")
	createPipeline.Flags().StringVarP(&username, "username", "u", "", "The username to push images as.")
	commands = append(commands, cmdutil.CreateAliases(createPipeline, "create pipeline", pipelines))

	var reprocess bool
	updatePipeline := &cobra.Command{
		Short: "Update an existing Pachyderm pipeline.",
		Long:  "Update a Pachyderm pipeline with a new pipeline specification. For details on the format, see https://docs.pachyderm.com/latest/reference/pipeline-spec/.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) (retErr error) {
			return pipelineHelper(reprocess, pushImages, registry, username, pipelinePath, jsonnetPath, jsonnetArgs, true)
		}),
	}
	updatePipeline.Flags().StringVarP(&pipelinePath, "file", "f", "", "A JSON file (url or filepath) containing one or more pipelines. \"-\" reads from stdin (the default behavior). Exactly one of --file and --jsonnet must be set.")
	updatePipeline.Flags().StringVar(&jsonnetPath, "jsonnet", "", "BETA: A Jsonnet template file (url or filepath) for one or more pipelines. \"-\" reads from stdin. Exactly one of --file and --jsonnet must be set. Jsonnet templates must contain a top-level function; strings can be passed to this function with --arg (below)")
	updatePipeline.Flags().StringArrayVar(&jsonnetArgs, "arg", nil, "Top-level argument passed to the Jsonnet template in --jsonnet (which must be set if any --arg arguments are passed). Value must be of the form 'param=value'. For multiple args, --arg may be set more than once.")
	updatePipeline.Flags().BoolVarP(&pushImages, "push-images", "p", false, "If true, push local docker images into the docker registry.")
	updatePipeline.Flags().StringVarP(&registry, "registry", "r", "index.docker.io", "The registry to push images to.")
	updatePipeline.Flags().StringVarP(&username, "username", "u", "", "The username to push images as.")
	updatePipeline.Flags().BoolVar(&reprocess, "reprocess", false, "If true, reprocess datums that were already processed by previous version of the pipeline.")
	commands = append(commands, cmdutil.CreateAliases(updatePipeline, "update pipeline", pipelines))

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
			pipelineInfo, err := client.InspectPipeline(args[0], true)
			if err != nil {
				return err
			}
			if raw {
				return errors.EnsureStack(cmdutil.Encoder(output, os.Stdout).EncodeProto(pipelineInfo))
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}
			pi := &pretty.PrintablePipelineInfo{
				PipelineInfo:   pipelineInfo,
				FullTimestamps: fullTimestamps,
			}
			return pretty.PrintDetailedPipelineInfo(os.Stdout, pi)
		}),
	}
	inspectPipeline.Flags().AddFlagSet(outputFlags)
	inspectPipeline.Flags().AddFlagSet(timestampFlags)
	commands = append(commands, cmdutil.CreateAliases(inspectPipeline, "inspect pipeline", pipelines))

	var editor string
	var editorArgs []string
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

			pipelineInfo, err := client.InspectPipeline(args[0], true)
			if err != nil {
				return err
			}

			createPipelineRequest := ppsutil.PipelineReqFromInfo(pipelineInfo)
			f, err := os.CreateTemp("", args[0])
			if err != nil {
				return errors.EnsureStack(err)
			}
			if err := cmdutil.Encoder(output, f).EncodeProto(createPipelineRequest); err != nil {
				return errors.EnsureStack(err)
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
			editorArgs = strings.Split(editor, " ")
			editorArgs = append(editorArgs, f.Name())
			if err := cmdutil.RunIO(cmdutil.IO{
				Stdin:  os.Stdin,
				Stdout: os.Stdout,
				Stderr: os.Stderr,
			}, editorArgs...); err != nil {
				return err
			}
			pipelineBytes, err := readPipelineBytes(f.Name())
			if err != nil {
				return err
			}
			pipelineReader, err := ppsutil.NewPipelineManifestReader(pipelineBytes)
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
			return txncmds.WithActiveTransaction(client, func(txClient *pachdclient.APIClient) error {
				_, err := txClient.PpsAPIClient.CreatePipeline(
					txClient.Ctx(),
					request,
				)
				return grpcutil.ScrubGRPC(err)
			})
		}),
	}
	editPipeline.Flags().BoolVar(&reprocess, "reprocess", false, "If true, reprocess datums that were already processed by previous version of the pipeline.")
	editPipeline.Flags().StringVar(&editor, "editor", "", "Editor to use for modifying the manifest.")
	editPipeline.Flags().StringVarP(&output, "output", "o", "", "Output format: \"json\" or \"yaml\" (default \"json\")")
	commands = append(commands, cmdutil.CreateAliases(editPipeline, "edit pipeline", pipelines))

	var spec bool
	var commit string
	listPipeline := &cobra.Command{
		Use:   "{{alias}} [<pipeline>]",
		Short: "Return info about all pipelines.",
		Long:  "Return info about all pipelines.",
		Run: cmdutil.RunBoundedArgs(0, 1, func(args []string) error {
			// validate flags
			if raw && spec {
				return errors.Errorf("cannot set both --raw and --spec")
			} else if !raw && !spec && output != "" {
				return errors.New("cannot set --output (-o) without --raw or --spec")
			}
			history, err := cmdutil.ParseHistory(history)
			if err != nil {
				return errors.Wrapf(err, "error parsing history flag")
			}
			var filter string
			if len(stateStrs) > 0 {
				filter, err = ParsePipelineStates(stateStrs)
				if err != nil {
					return errors.Wrap(err, "error parsing state")
				}
			}
			// init client & get pipeline info
			client, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "error connecting to pachd")
			}
			defer client.Close()
			var pipeline string
			if len(args) > 0 {
				pipeline = args[0]
			}
			request := &ppsclient.ListPipelineRequest{
				History:   history,
				CommitSet: &pfs.CommitSet{ID: commit},
				JqFilter:  filter,
				Details:   true,
			}
			if pipeline != "" {
				request.Pipeline = pachdclient.NewPipeline(pipeline)
			}
			lpClient, err := client.PpsAPIClient.ListPipeline(client.Ctx(), request)
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			pipelineInfos, err := clientsdk.ListPipelineInfo(lpClient)
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			if raw {
				e := cmdutil.Encoder(output, os.Stdout)
				for _, pipelineInfo := range pipelineInfos {
					if err := e.EncodeProto(pipelineInfo); err != nil {
						return errors.EnsureStack(err)
					}
				}
				return nil
			} else if spec {
				e := cmdutil.Encoder(output, os.Stdout)
				for _, pipelineInfo := range pipelineInfos {
					if err := e.EncodeProto(ppsutil.PipelineReqFromInfo(pipelineInfo)); err != nil {
						return errors.EnsureStack(err)
					}
				}
				return nil
			}
			for _, pi := range pipelineInfos {
				if ppsutil.ErrorState(pi.State) {
					fmt.Fprintln(os.Stderr, "One or more pipelines have encountered errors, use inspect pipeline to get more info.")
					break
				}
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
	listPipeline.Flags().AddFlagSet(timestampFlags)
	listPipeline.Flags().StringVar(&history, "history", "none", "Return revision history for pipelines.")
	listPipeline.Flags().StringVarP(&commit, "commit", "c", "", "List the pipelines as they existed at this commit.")
	listPipeline.Flags().StringArrayVar(&stateStrs, "state", []string{}, "Return only pipelines with the specified state. Can be repeated to include multiple states")
	commands = append(commands, cmdutil.CreateAliases(listPipeline, "list pipeline", pipelines))

	var (
		all      bool
		force    bool
		keepRepo bool
	)
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
				return errors.Errorf("cannot use the --all flag with an argument")
			}
			if len(args) == 0 && !all {
				return errors.Errorf("either a pipeline name or the --all flag needs to be provided")
			}
			req := &ppsclient.DeletePipelineRequest{
				All:      all,
				Force:    force,
				KeepRepo: keepRepo,
			}
			if len(args) > 0 {
				req.Pipeline = pachdclient.NewPipeline(args[0])
			}
			if _, err = client.PpsAPIClient.DeletePipeline(client.Ctx(), req); err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			return nil
		}),
	}
	deletePipeline.Flags().BoolVar(&all, "all", false, "delete all pipelines")
	deletePipeline.Flags().BoolVarP(&force, "force", "f", false, "delete the pipeline regardless of errors; use with care")
	deletePipeline.Flags().BoolVar(&keepRepo, "keep-repo", false, "delete the pipeline, but keep the output repo data around (the pipeline cannot be recreated later with the same name unless the repo is deleted)")
	commands = append(commands, cmdutil.CreateAliases(deletePipeline, "delete pipeline", pipelines))

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
				return errors.Wrap(err, "error from StartPipeline")
			}
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAliases(startPipeline, "start pipeline", pipelines))

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
				return errors.Wrap(err, "error from StopPipeline")
			}
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAliases(stopPipeline, "stop pipeline", pipelines))

	var file string
	createSecret := &cobra.Command{
		Short: "Create a secret on the cluster.",
		Long:  "Create a secret on the cluster.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) (retErr error) {
			client, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer client.Close()
			fileBytes, err := os.ReadFile(file)
			if err != nil {
				return errors.EnsureStack(err)
			}

			_, err = client.PpsAPIClient.CreateSecret(
				client.Ctx(),
				&ppsclient.CreateSecretRequest{
					File: fileBytes,
				})

			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			return nil
		}),
	}
	createSecret.Flags().StringVarP(&file, "file", "f", "", "File containing Kubernetes secret.")
	commands = append(commands, cmdutil.CreateAliases(createSecret, "create secret", secrets))

	deleteSecret := &cobra.Command{
		Short: "Delete a secret from the cluster.",
		Long:  "Delete a secret from the cluster.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) (retErr error) {
			client, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer client.Close()

			_, err = client.PpsAPIClient.DeleteSecret(
				client.Ctx(),
				&ppsclient.DeleteSecretRequest{
					Secret: &ppsclient.Secret{
						Name: args[0],
					},
				})

			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAliases(deleteSecret, "delete secret", secrets))

	inspectSecret := &cobra.Command{
		Short: "Inspect a secret from the cluster.",
		Long:  "Inspect a secret from the cluster.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) (retErr error) {
			client, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer client.Close()

			secretInfo, err := client.PpsAPIClient.InspectSecret(
				client.Ctx(),
				&ppsclient.InspectSecretRequest{
					Secret: &ppsclient.Secret{
						Name: args[0],
					},
				})

			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			writer := tabwriter.NewWriter(os.Stdout, pretty.SecretHeader)
			pretty.PrintSecretInfo(writer, secretInfo)
			return writer.Flush()
		}),
	}
	commands = append(commands, cmdutil.CreateAliases(inspectSecret, "inspect secret", secrets))

	listSecret := &cobra.Command{
		Short: "List all secrets from a namespace in the cluster.",
		Long:  "List all secrets from a namespace in the cluster.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) (retErr error) {
			client, err := pachdclient.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer client.Close()

			secretInfos, err := client.PpsAPIClient.ListSecret(
				client.Ctx(),
				&types.Empty{},
			)

			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			writer := tabwriter.NewWriter(os.Stdout, pretty.SecretHeader)
			for _, si := range secretInfos.GetSecretInfo() {
				pretty.PrintSecretInfo(writer, si)
			}
			return writer.Flush()
		}),
	}
	commands = append(commands, cmdutil.CreateAliases(listSecret, "list secret", secrets))

	var dagSpecFile string
	var seed int64
	var parallelism int64
	var podPatchFile string
	runLoadTest := &cobra.Command{
		Use:   "{{alias}} <spec-file> ",
		Short: "Run a PPS load test.",
		Long:  "Run a PPS load test.",
		Run: cmdutil.RunBoundedArgs(0, 1, func(args []string) (retErr error) {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer func() {
				if err := c.Close(); retErr == nil {
					retErr = err
				}
			}()
			if len(args) == 0 {
				resp, err := c.PpsAPIClient.RunLoadTestDefault(c.Ctx(), &types.Empty{})
				if err != nil {
					return errors.EnsureStack(err)
				}
				if err := cmdutil.Encoder(output, os.Stdout).EncodeProto(resp); err != nil {
					return errors.EnsureStack(err)
				}
				fmt.Println()
				return nil
			}
			var dagSpec []byte
			if dagSpecFile != "" {
				var err error
				dagSpec, err = os.ReadFile(dagSpecFile)
				if err != nil {
					return errors.EnsureStack(err)
				}
			}
			var podPatch []byte
			if podPatchFile != "" {
				podPatch, err = os.ReadFile(podPatchFile)
				if err != nil {
					return errors.EnsureStack(err)
				}
			}
			err = filepath.Walk(args[0], func(file string, fi os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if fi.IsDir() {
					return nil
				}
				loadSpec, err := os.ReadFile(file)
				if err != nil {
					return errors.EnsureStack(err)
				}
				resp, err := c.PpsAPIClient.RunLoadTest(c.Ctx(), &pps.RunLoadTestRequest{
					DagSpec:     string(dagSpec),
					LoadSpec:    string(loadSpec),
					Seed:        seed,
					Parallelism: parallelism,
					PodPatch:    string(podPatch),
				})
				if err != nil {
					return errors.EnsureStack(err)
				}
				if err := cmdutil.Encoder(output, os.Stdout).EncodeProto(resp); err != nil {
					return errors.EnsureStack(err)
				}
				fmt.Println()
				return nil
			})
			return errors.EnsureStack(err)
		}),
	}
	runLoadTest.Flags().StringVarP(&dagSpecFile, "dag", "d", "", "The DAG specification file to use for the load test")
	runLoadTest.Flags().Int64VarP(&seed, "seed", "s", 0, "The seed to use for generating the load.")
	runLoadTest.Flags().Int64VarP(&parallelism, "parallelism", "p", 0, "The parallelism to use for the pipelines.")
	runLoadTest.Flags().StringVarP(&podPatchFile, "pod-patch", "", "", "The pod patch file to use for the pipelines.")
	commands = append(commands, cmdutil.CreateAlias(runLoadTest, "run pps-load-test"))

	return commands
}

// readPipelineBytes reads the pipeline spec at 'pipelinePath' (which may be
// '-' for stdin, a local path, or a remote URL) and returns the bytes stored
// there.
//
// TODO(msteffen) This is very similar to readConfigBytes in
// s/s/identity/cmds/cmds.go (which differs only in not supporting URLs),
// so the two could perhaps be refactored.
func readPipelineBytes(pipelinePath string) (pipelineBytes []byte, retErr error) {
	if pipelinePath == "-" {
		cmdutil.PrintStdinReminder()
		var err error
		pipelineBytes, err = io.ReadAll(os.Stdin)
		if err != nil {
			return nil, errors.EnsureStack(err)
		}
	} else if url, err := url.Parse(pipelinePath); err == nil && url.Scheme != "" {
		resp, err := http.Get(url.String())
		if err != nil {
			return nil, errors.EnsureStack(err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil && retErr == nil {
				retErr = err
			}
		}()
		pipelineBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.EnsureStack(err)
		}
	} else {
		var err error
		pipelineBytes, err = os.ReadFile(pipelinePath)
		if err != nil {
			return nil, errors.EnsureStack(err)
		}
	}
	return pipelineBytes, nil
}

func evaluateJsonnetTemplate(client *client.APIClient, jsonnetPath string, jsonnetArgs []string) ([]byte, error) {
	templateBytes, err := readPipelineBytes(jsonnetPath)
	if err != nil {
		return nil, err
	}
	args, err := pachtmpl.ParseArgs(jsonnetArgs)
	if err != nil {
		return nil, err
	}
	res, err := client.RenderTemplate(client.Ctx(), &ppsclient.RenderTemplateRequest{
		Template: string(templateBytes),
		Args:     args,
	})
	if err != nil {
		return nil, err
	}
	return []byte(res.Json), nil
}

func pipelineHelper(reprocess bool, pushImages bool, registry, username, pipelinePath, jsonnetPath string, jsonnetArgs []string, update bool) error {
	// validate arguments
	if pipelinePath != "" && jsonnetPath != "" {
		return errors.New("cannot set both --file and --jsonnet; exactly one must be set")
	}
	if pipelinePath == "" && jsonnetPath == "" {
		pipelinePath = "-" // default input
	}
	pc, err := pachdclient.NewOnUserMachine("user")
	if err != nil {
		return errors.Wrapf(err, "error connecting to pachd")
	}
	defer pc.Close()
	// read/compute pipeline spec(s) (file, stdin, url, or via template)
	var pipelineBytes []byte
	if pipelinePath != "" {
		pipelineBytes, err = readPipelineBytes(pipelinePath)
	} else if jsonnetPath != "" {
		pipelineBytes, err = evaluateJsonnetTemplate(pc, jsonnetPath, jsonnetArgs)
	}
	if err != nil {
		return err
	}
	pipelineReader, err := ppsutil.NewPipelineManifestReader(pipelineBytes)
	if err != nil {
		return err
	}
	for {
		request, err := pipelineReader.NextCreatePipelineRequest()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}

		if request.Pipeline == nil {
			return errors.New("no `pipeline` specified")
		}
		if request.Pipeline.Name == "" {
			return errors.New("no pipeline `name` specified")
		}

		// Add trace if env var is set
		ctx, err := extended.EmbedAnyDuration(pc.Ctx())
		pc = pc.WithCtx(ctx)
		if err != nil {
			logrus.Warning(err)
		}

		if update {
			request.Update = true
			request.Reprocess = reprocess
		}

		if pushImages {
			if request.Transform == nil {
				return errors.New("must specify a pipeline `transform`")
			}
			if err := dockerPushHelper(request, registry, username); err != nil {
				return err
			}
		}

		if request.Transform != nil && request.Transform.Image != "" {
			if !strings.Contains(request.Transform.Image, ":") {
				fmt.Fprintf(os.Stderr,
					"WARNING: please specify a tag for the docker image in your transform.image spec.\n"+
						"For example, change 'python' to 'python:3' or 'bash' to 'bash:5'. This improves\n"+
						"reproducibility of your pipelines.\n\n")
			} else if strings.HasSuffix(request.Transform.Image, ":latest") {
				fmt.Fprintf(os.Stderr,
					"WARNING: please do not specify the ':latest' tag for the docker image in your\n"+
						"transform.image spec. For example, change 'python:latest' to 'python:3' or\n"+
						"'bash:latest' to 'bash:5'. This improves reproducibility of your pipelines.\n\n")
			}
		}
		if err = txncmds.WithActiveTransaction(pc, func(txClient *pachdclient.APIClient) error {
			_, err := txClient.PpsAPIClient.CreatePipeline(
				txClient.Ctx(),
				request,
			)
			return grpcutil.ScrubGRPC(err)
		}); err != nil {
			return err
		}
	}

	return nil
}

func dockerPushHelper(request *ppsclient.CreatePipelineRequest, registry, username string) error {
	// create docker client
	dockerClient, err := docker.NewClientFromEnv()
	if err != nil {
		return errors.Wrapf(err, "could not create a docker client from the environment")
	}

	var authConfig docker.AuthConfiguration
	detectedAuthConfig := false

	// try to automatically determine the credentials
	authConfigs, err := docker.NewAuthConfigurationsFromDockerCfg()
	if err == nil {
		for _, ac := range authConfigs.Configs {
			u, err := url.Parse(ac.ServerAddress)
			if err == nil && u.Hostname() == registry && (username == "" || username == ac.Username) {
				authConfig = ac
				detectedAuthConfig = true
				break
			}
		}
	}
	// if that failed, manually build credentials
	if !detectedAuthConfig {
		if username == "" {
			// request the username if it hasn't been specified yet
			fmt.Printf("Username for %s: ", registry)
			reader := bufio.NewReader(os.Stdin)
			username, err = reader.ReadString('\n')
			if err != nil {
				return errors.Wrapf(err, "could not read username")
			}
			username = strings.TrimRight(username, "\r\n")
		}

		// request the password
		password, err := cmdutil.ReadPassword(fmt.Sprintf("Password for %s@%s: ", username, registry))
		if err != nil {
			return errors.Wrapf(err, "could not read password")
		}

		authConfig = docker.AuthConfiguration{
			Username: username,
			Password: password,
		}
	}

	repo, sourceTag := docker.ParseRepositoryTag(request.Transform.Image)
	if sourceTag == "" {
		sourceTag = "latest"
	}
	destTag := uuid.NewWithoutDashes()

	sourceImage := fmt.Sprintf("%s:%s", repo, sourceTag)
	destImage := fmt.Sprintf("%s:%s", repo, destTag)

	fmt.Printf("Tagging/pushing %q, this may take a while.\n", destImage)

	if err := dockerClient.TagImage(sourceImage, docker.TagImageOptions{
		Repo:    repo,
		Tag:     destTag,
		Context: context.Background(),
	}); err != nil {
		return errors.Wrapf(err, "could not tag docker image")
	}

	if err := dockerClient.PushImage(
		docker.PushImageOptions{
			Name: repo,
			Tag:  destTag,
		},
		authConfig,
	); err != nil {
		return errors.Wrapf(err, "could not push docker image")
	}

	request.Transform.Image = destImage
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

func validateJQConditionString(filter string) (string, error) {
	q, err := gojq.Parse(filter)
	if err != nil {
		return "", errors.EnsureStack(err)
	}
	_, err = gojq.Compile(q)
	if err != nil {
		return "", errors.EnsureStack(err)
	}
	return filter, nil
}

// ParseJobStates parses a slice of state names into a jq filter suitable for ListJob
func ParseJobStates(stateStrs []string) (string, error) {
	var conditions []string
	for _, stateStr := range stateStrs {
		if state, err := ppsclient.JobStateFromName(stateStr); err == nil {
			conditions = append(conditions, fmt.Sprintf(".state == \"%s\"", state))
		} else {
			return "", err
		}
	}
	return validateJQConditionString(strings.Join(conditions, " or "))
}

// ParsePipelineStates parses a slice of state names into a jq filter suitable for ListPipeline
func ParsePipelineStates(stateStrs []string) (string, error) {
	var conditions []string
	for _, stateStr := range stateStrs {
		if state, err := ppsclient.PipelineStateFromName(stateStr); err == nil {
			conditions = append(conditions, fmt.Sprintf(".state == \"%s\"", state))
		} else {
			return "", err
		}
	}
	return validateJQConditionString(strings.Join(conditions, " or "))
}
