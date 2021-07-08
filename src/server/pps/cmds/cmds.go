package cmds

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	pachdclient "github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/clientsdk"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pager"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/tabwriter"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing/extended"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
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
	"golang.org/x/net/context"
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
	commands = append(commands, cmdutil.CreateDocsAlias(jobDocs, "job", " job$"))

	var wait bool
	inspectJob := &cobra.Command{
		Use:   "{{alias}} <pipeline>@<job>",
		Short: "Return info about a job.",
		Long:  "Return info about a job.",
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) error {
			job, err := cmdutil.ParseJob(args[0])
			if err != nil {
				return err
			}
			jobInfo, err := env.Client("user").InspectJob(job.Pipeline.Name, job.ID, true)
			if err != nil {
				return errors.Wrap(err, "error from InspectJob")
			}
			if raw {
				return cmdutil.Encoder(output, env.Out()).EncodeProto(jobInfo)
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}
			pji := &pretty.PrintableJobInfo{
				JobInfo:        jobInfo,
				FullTimestamps: fullTimestamps,
			}
			return pretty.PrintDetailedJobInfo(env.Out(), pji)
		}),
	}
	inspectJob.Flags().AddFlagSet(outputFlags)
	inspectJob.Flags().AddFlagSet(timestampFlags)
	shell.RegisterCompletionFunc(inspectJob, shell.JobCompletion)
	commands = append(commands, cmdutil.CreateAlias(inspectJob, "inspect job"))

	waitJob := &cobra.Command{
		Use:   "{{alias}} <pipeline>@<job>",
		Short: "Wait for a job to finish then return info about the job.",
		Long:  "Wait for a job to finish then return info about the job.",
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) error {
			job, err := cmdutil.ParseJob(args[0])
			if err != nil {
				return err
			}
			jobInfo, err := env.Client("user").WaitJob(job.Pipeline.Name, job.ID, true)
			if err != nil {
				errors.Wrap(err, "error from InspectJob")
			}
			if raw {
				return cmdutil.Encoder(output, env.Out()).EncodeProto(jobInfo)
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}
			pji := &pretty.PrintableJobInfo{
				JobInfo:        jobInfo,
				FullTimestamps: fullTimestamps,
			}
			return pretty.PrintDetailedJobInfo(env.Out(), pji)
		}),
	}
	waitJob.Flags().AddFlagSet(outputFlags)
	waitJob.Flags().AddFlagSet(timestampFlags)
	shell.RegisterCompletionFunc(waitJob, shell.JobCompletion)
	commands = append(commands, cmdutil.CreateAlias(waitJob, "wait job"))

	inspectJobSet := &cobra.Command{
		Use:   "{{alias}} <job-id>",
		Short: "Return info about all the jobs in a jobset.",
		Long:  "Return info about all the jobs in a jobset.",
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) error {
			var err error
			var jobInfos []*pps.JobInfo
			if wait {
				jobInfos, err = env.Client("user").WaitJobSetAll(args[0], false)
			} else {
				jobInfos, err = env.Client("user").InspectJobSet(args[0], false)
			}
			if err != nil {
				return errors.Wrap(err, "error from InspectJobSet")
			}

			if raw {
				encoder := cmdutil.Encoder(output, env.Out())
				for _, jobInfo := range jobInfos {
					if err := encoder.EncodeProto(jobInfo); err != nil {
						return err
					}
				}
				return nil
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}

			return pager.Page(noPager, env, func(w io.Writer) error {
				writer := tabwriter.NewWriter(w, pretty.JobHeader)
				for _, jobInfo := range jobInfos {
					pretty.PrintJobInfo(writer, jobInfo, fullTimestamps)
				}
				return writer.Flush()
			})
		}),
	}
	inspectJobSet.Flags().BoolVarP(&wait, "wait", "w", false, "wait until each job has either succeeded or failed")
	inspectJobSet.Flags().AddFlagSet(outputFlags)
	inspectJobSet.Flags().AddFlagSet(timestampFlags)
	inspectJobSet.Flags().AddFlagSet(pagerFlags)
	shell.RegisterCompletionFunc(inspectJobSet, shell.JobCompletion)
	commands = append(commands, cmdutil.CreateAlias(inspectJobSet, "inspect jobset"))

	listJobSet := &cobra.Command{
		Short: "Return info about jobsets.",
		Long:  "Return info about jobsets.",
		Example: `
# Return all jobsets
$ {{alias}}`,
		RunE: cmdutil.RunFixedArgs(0, func(args []string, env cmdutil.Env) error {
			client := env.Client("user")
			listJobSetClient, err := client.PpsAPIClient.ListJobSet(client.Ctx(), &pps.ListJobSetRequest{})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}

			if raw {
				e := cmdutil.Encoder(output, env.Out())
				return clientsdk.ForEachJobSet(listJobSetClient, func(jobSetInfo *pps.JobSetInfo) error {
					return e.EncodeProto(jobSetInfo)
				})
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}

			return pager.Page(noPager, env, func(w io.Writer) error {
				writer := tabwriter.NewWriter(w, pretty.JobSetHeader)
				if err := clientsdk.ForEachJobSet(listJobSetClient, func(jobSetInfo *pps.JobSetInfo) error {
					pretty.PrintJobSetInfo(writer, jobSetInfo, fullTimestamps)
					return nil
				}); err != nil {
					return err
				}
				return writer.Flush()
			})
		}),
	}
	listJobSet.Flags().AddFlagSet(outputFlags)
	listJobSet.Flags().AddFlagSet(timestampFlags)
	listJobSet.Flags().AddFlagSet(pagerFlags)
	commands = append(commands, cmdutil.CreateAlias(listJobSet, "list jobset"))

	var pipelineName string
	var inputCommitStrs []string
	var history string
	var stateStrs []string
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
		RunE: cmdutil.RunFixedArgs(0, func(args []string, env cmdutil.Env) error {
			commits, err := cmdutil.ParseCommits(inputCommitStrs)
			if err != nil {
				return err
			}
			history, err := cmdutil.ParseHistory(history)
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

			if raw {
				encoder := cmdutil.Encoder(output, env.Out())
				return env.Client("user").ListJobFilterF(pipelineName, commits, history, true, filter, func(ji *ppsclient.JobInfo) error {
					return encoder.EncodeProto(ji)
				})
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}

			return pager.Page(noPager, env, func(w io.Writer) error {
				writer := tabwriter.NewWriter(w, pretty.JobHeader)
				if err := env.Client("user").ListJobFilterF(pipelineName, commits, history, false, filter, func(ji *ppsclient.JobInfo) error {
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
	listJob.Flags().StringSliceVarP(&inputCommitStrs, "input", "i", []string{}, "List jobs with a specific set of input commits. format: <repo>@<branch-or-commit>")
	listJob.MarkFlagCustom("input", "__pachctl_get_repo_commit")
	listJob.Flags().AddFlagSet(outputFlags)
	listJob.Flags().AddFlagSet(timestampFlags)
	listJob.Flags().AddFlagSet(pagerFlags)
	listJob.Flags().StringVar(&history, "history", "none", "Return jobs from historical versions of pipelines.")
	listJob.Flags().StringArrayVar(&stateStrs, "state", []string{}, "Return only jobs with the specified state. Can be repeated to include multiple states")
	shell.RegisterCompletionFunc(listJob,
		func(flag, text string, maxCompletions int64) ([]prompt.Suggest, shell.CacheFunc) {
			if flag == "-p" || flag == "--pipeline" {
				cs, cf := shell.PipelineCompletion(flag, text, maxCompletions)
				return cs, shell.AndCacheFunc(cf, shell.SameFlag(flag))
			}
			return nil, shell.SameFlag(flag)
		})
	commands = append(commands, cmdutil.CreateAlias(listJob, "list job"))

	deleteJob := &cobra.Command{
		Use:   "{{alias}} <pipeline>@<job>",
		Short: "Delete a job.",
		Long:  "Delete a job.",
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) error {
			job, err := cmdutil.ParseJob(args[0])
			if err != nil {
				return err
			}
			if err := env.Client("user").DeleteJob(job.Pipeline.Name, job.ID); err != nil {
				return errors.Wrap(err, "error from DeleteJob")
			}
			return nil
		}),
	}
	shell.RegisterCompletionFunc(deleteJob, shell.JobCompletion)
	commands = append(commands, cmdutil.CreateAlias(deleteJob, "delete job"))

	stopJob := &cobra.Command{
		Use:   "{{alias}} <pipeline>@<job>",
		Short: "Stop a job.",
		Long:  "Stop a job.  The job will be stopped immediately.",
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) error {
			job, err := cmdutil.ParseJob(args[0])
			if err != nil {
				return err
			}
			if err := env.Client("user").StopJob(job.Pipeline.Name, job.ID); err != nil {
				return errors.Wrap(err, "error from StopJob")
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
		Use:   "{{alias}} <pipeline>@<job> <datum-path1>,<datum-path2>,...",
		Short: "Restart a datum.",
		Long:  "Restart a datum.",
		RunE: cmdutil.RunFixedArgs(2, func(args []string, env cmdutil.Env) error {
			job, err := cmdutil.ParseJob(args[0])
			if err != nil {
				return err
			}
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
			return env.Client("user").RestartDatum(job.Pipeline.Name, job.ID, datumFilter)
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(restartDatum, "restart datum"))

	var pipelineInputPath string
	listDatum := &cobra.Command{
		Use:   "{{alias}} <pipeline>@<job>",
		Short: "Return the datums in a job.",
		Long:  "Return the datums in a job.",
		RunE: cmdutil.RunBoundedArgs(0, 1, func(args []string, env cmdutil.Env) (retErr error) {
			var printF func(*ppsclient.DatumInfo) error
			if !raw {
				if output != "" {
					return errors.New("cannot set --output (-o) without --raw")
				}
				writer := tabwriter.NewWriter(env.Out(), pretty.DatumHeader)
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
				encoder := cmdutil.Encoder(output, env.Out())
				printF = func(di *ppsclient.DatumInfo) error {
					return encoder.EncodeProto(di)
				}
			}
			if pipelineInputPath != "" && len(args) == 1 {
				return errors.Errorf("can't specify both a job and a pipeline spec")
			} else if pipelineInputPath != "" {
				pipelineReader, err := ppsutil.NewPipelineManifestReader(pipelineInputPath)
				if err != nil {
					return err
				}
				request, err := pipelineReader.NextCreatePipelineRequest()
				if err != nil {
					return err
				}
				return env.Client("user").ListDatumInput(request.Input, printF)
			} else if len(args) == 1 {
				job, err := cmdutil.ParseJob(args[0])
				if err != nil {
					return err
				}
				return env.Client("user").ListDatum(job.Pipeline.Name, job.ID, printF)
			} else {
				return errors.Errorf("must specify either a job or a pipeline spec")
			}
		}),
	}
	listDatum.Flags().StringVarP(&pipelineInputPath, "file", "f", "", "The JSON file containing the pipeline to list datums from, the pipeline need not exist")
	listDatum.Flags().AddFlagSet(outputFlags)
	shell.RegisterCompletionFunc(listDatum, shell.JobCompletion)
	commands = append(commands, cmdutil.CreateAlias(listDatum, "list datum"))

	inspectDatum := &cobra.Command{
		Use:   "{{alias}} <pipeline>@<job> <datum>",
		Short: "Display detailed info about a single datum.",
		Long:  "Display detailed info about a single datum. Requires the pipeline to have stats enabled.",
		RunE: cmdutil.RunFixedArgs(2, func(args []string, env cmdutil.Env) error {
			job, err := cmdutil.ParseJob(args[0])
			if err != nil {
				return err
			}
			datumInfo, err := env.Client("user").InspectDatum(job.Pipeline.Name, job.ID, args[1])
			if err != nil {
				return err
			}
			if raw {
				return cmdutil.Encoder(output, env.Out()).EncodeProto(datumInfo)
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}
			pretty.PrintDetailedDatumInfo(env.Out(), datumInfo)
			return nil
		}),
	}
	inspectDatum.Flags().AddFlagSet(outputFlags)
	commands = append(commands, cmdutil.CreateAlias(inspectDatum, "inspect datum"))

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
		RunE: cmdutil.RunFixedArgs(0, func(args []string, env cmdutil.Env) error {
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
			iter := env.Client("user").GetLogs(pipelineName, jobID, data, datumID, master, follow, since)
			var buf bytes.Buffer
			encoder := json.NewEncoder(&buf)
			for iter.Next() {
				if raw {
					buf.Reset()
					if err := encoder.Encode(iter.Message()); err != nil {
						fmt.Fprintf(env.Err(), "error marshalling \"%v\": %s\n", iter.Message(), err)
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
	commands = append(commands, cmdutil.CreateDocsAlias(pipelineDocs, "pipeline", " pipeline$"))

	var pushImages bool
	var registry string
	var username string
	var pipelinePath string
	createPipeline := &cobra.Command{
		Short: "Create a new pipeline.",
		Long:  "Create a new pipeline from a pipeline specification. For details on the format, see http://docs.pachyderm.io/en/latest/reference/pipeline_spec.html.",
		RunE: cmdutil.RunFixedArgs(0, func(args []string, env cmdutil.Env) error {
			return pipelineHelper(env, false, pushImages, registry, username, pipelinePath, false)
		}),
	}
	createPipeline.Flags().StringVarP(&pipelinePath, "file", "f", "-", "The JSON file containing the pipeline, it can be a url or local file. - reads from stdin.")
	createPipeline.Flags().BoolVarP(&pushImages, "push-images", "p", false, "If true, push local docker images into the docker registry.")
	createPipeline.Flags().StringVarP(&registry, "registry", "r", "index.docker.io", "The registry to push images to.")
	createPipeline.Flags().StringVarP(&username, "username", "u", "", "The username to push images as.")
	commands = append(commands, cmdutil.CreateAlias(createPipeline, "create pipeline"))

	var reprocess bool
	updatePipeline := &cobra.Command{
		Short: "Update an existing Pachyderm pipeline.",
		Long:  "Update a Pachyderm pipeline with a new pipeline specification. For details on the format, see http://docs.pachyderm.io/en/latest/reference/pipeline_spec.html.",
		RunE: cmdutil.RunFixedArgs(0, func(args []string, env cmdutil.Env) error {
			return pipelineHelper(env, reprocess, pushImages, registry, username, pipelinePath, true)
		}),
	}
	updatePipeline.Flags().StringVarP(&pipelinePath, "file", "f", "-", "The JSON file containing the pipeline, it can be a url or local file. - reads from stdin.")
	updatePipeline.Flags().BoolVarP(&pushImages, "push-images", "p", false, "If true, push local docker images into the docker registry.")
	updatePipeline.Flags().StringVarP(&registry, "registry", "r", "index.docker.io", "The registry to push images to.")
	updatePipeline.Flags().StringVarP(&username, "username", "u", "", "The username to push images as.")
	updatePipeline.Flags().BoolVar(&reprocess, "reprocess", false, "If true, reprocess datums that were already processed by previous version of the pipeline.")
	commands = append(commands, cmdutil.CreateAlias(updatePipeline, "update pipeline"))

	runPipeline := &cobra.Command{
		Use:   "{{alias}} <pipeline> [--job=<job>] [<repo>@[<branch>|<commit>|<branch>=<commit>]...]",
		Short: "Run an existing Pachyderm pipeline on the specified commits-branch pairs.",
		Long:  "Run a Pachyderm pipeline on the datums from specific commit-branch pairs. If you only specify a branch, Pachyderm uses the HEAD commit to complete the pair. Similarly, if you only specify a commit, Pachyderm will try to use the branch the commit originated on. Note: Pipelines run automatically when data is committed to them. This command is for the case where you want to run the pipeline on a specific set of data.",
		Example: `
		# Rerun the latest job for the "filter" pipeline
		$ {{alias}} filter

		# Process the pipeline "filter" on the data from commit-branch pairs "repo1@A=a23e4" and "repo2@B=bf363"
		$ {{alias}} filter repo1@A=a23e4 repo2@B=bf363

		# Run the pipeline "filter" on the data from commit "167af5" on the "staging" branch on repo "repo1"
		$ {{alias}} filter repo1@staging=167af5

		# Run the pipeline "filter" on the HEAD commit of the "testing" branch on repo "repo1"
		$ {{alias}} filter repo1@testing

		# Run the pipeline "filter" on the commit "af159e which originated on the "master" branch on repo "repo1"
		$ {{alias}} filter repo1@af159`,

		RunE: cmdutil.RunMinimumArgs(1, func(args []string, env cmdutil.Env) error {
			prov, err := cmdutil.ParseCommits(args[1:])
			if err != nil {
				return err
			}
			err = env.Client("user").RunPipeline(args[0], prov, jobStr)
			if err != nil {
				return err
			}
			return nil
		}),
	}
	runPipeline.Flags().StringVar(&jobStr, "job", "", "rerun the given job")
	commands = append(commands, cmdutil.CreateAlias(runPipeline, "run pipeline"))

	runCron := &cobra.Command{
		Use:   "{{alias}} <pipeline>",
		Short: "Run an existing Pachyderm cron pipeline now",
		Long:  "Run an existing Pachyderm cron pipeline now",
		Example: `
		# Run a cron pipeline "clock" now
		$ {{alias}} clock`,
		RunE: cmdutil.RunMinimumArgs(1, func(args []string, env cmdutil.Env) error {
			return env.Client("user").RunCron(args[0])
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(runCron, "run cron"))

	inspectPipeline := &cobra.Command{
		Use:   "{{alias}} <pipeline>",
		Short: "Return info about a pipeline.",
		Long:  "Return info about a pipeline.",
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) error {
			pipelineInfo, err := env.Client("user").InspectPipeline(args[0], true)
			if err != nil {
				return err
			}
			if raw {
				return cmdutil.Encoder(output, env.Out()).EncodeProto(pipelineInfo)
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}
			pi := &pretty.PrintablePipelineInfo{
				PipelineInfo:   pipelineInfo,
				FullTimestamps: fullTimestamps,
			}
			return pretty.PrintDetailedPipelineInfo(env.Out(), pi)
		}),
	}
	inspectPipeline.Flags().AddFlagSet(outputFlags)
	inspectPipeline.Flags().AddFlagSet(timestampFlags)
	commands = append(commands, cmdutil.CreateAlias(inspectPipeline, "inspect pipeline"))

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
			f, err := ioutil.TempFile("", args[0])
			if err != nil {
				return err
			}
			if err := cmdutil.Encoder(output, f).EncodeProto(createPipelineRequest); err != nil {
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
			editorArgs = strings.Split(editor, " ")
			editorArgs = append(editorArgs, f.Name())
			if err := cmdutil.RunIO(cmdutil.IO{
				Stdin:  os.Stdin,
				Stdout: os.Stdout,
				Stderr: os.Stderr,
			}, editorArgs...); err != nil {
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
	commands = append(commands, cmdutil.CreateAlias(editPipeline, "edit pipeline"))

	var spec bool
	listPipeline := &cobra.Command{
		Use:   "{{alias}} [<pipeline>]",
		Short: "Return info about all pipelines.",
		Long:  "Return info about all pipelines.",
		RunE: cmdutil.RunBoundedArgs(0, 1, func(args []string, env cmdutil.Env) error {
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
			var pipeline string
			if len(args) > 0 {
				pipeline = args[0]
			}
			request := &ppsclient.ListPipelineRequest{History: history, JqFilter: filter, Details: true}
			if pipeline != "" {
				request.Pipeline = pachdclient.NewPipeline(pipeline)
			}
			c := env.Client("user")
			lpClient, err := c.PpsAPIClient.ListPipeline(c.Ctx(), request)
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			pipelineInfos, err := clientsdk.ListPipelineInfo(lpClient)
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			if raw {
				encoder := cmdutil.Encoder(output, env.Out())
				for _, pipelineInfo := range pipelineInfos {
					if err := encoder.EncodeProto(pipelineInfo); err != nil {
						return err
					}
				}
				return nil
			} else if spec {
				encoder := cmdutil.Encoder(output, env.Out())
				for _, pipelineInfo := range pipelineInfos {
					if err := encoder.EncodeProto(ppsutil.PipelineReqFromInfo(pipelineInfo)); err != nil {
						return err
					}
				}
				return nil
			}
			for _, pi := range pipelineInfos {
				if ppsutil.ErrorState(pi.State) {
					fmt.Fprintln(env.Err(), "One or more pipelines have encountered errors, use inspect pipeline to get more info.")
					break
				}
			}
			writer := tabwriter.NewWriter(env.Out(), pretty.PipelineHeader)
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
	listPipeline.Flags().StringArrayVar(&stateStrs, "state", []string{}, "Return only pipelines with the specified state. Can be repeated to include multiple states")
	commands = append(commands, cmdutil.CreateAlias(listPipeline, "list pipeline"))

	var (
		all      bool
		force    bool
		keepRepo bool
	)
	deletePipeline := &cobra.Command{
		Use:   "{{alias}} (<pipeline>|--all)",
		Short: "Delete a pipeline.",
		Long:  "Delete a pipeline.",
		RunE: cmdutil.RunBoundedArgs(0, 1, func(args []string, env cmdutil.Env) error {
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
			c := env.Client("user")
			if _, err := c.PpsAPIClient.DeletePipeline(c.Ctx(), req); err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			return nil
		}),
	}
	deletePipeline.Flags().BoolVar(&all, "all", false, "delete all pipelines")
	deletePipeline.Flags().BoolVarP(&force, "force", "f", false, "delete the pipeline regardless of errors; use with care")
	deletePipeline.Flags().BoolVar(&keepRepo, "keep-repo", false, "delete the pipeline, but keep the output repo around (the pipeline can be recreated later and use the same repo)")
	commands = append(commands, cmdutil.CreateAlias(deletePipeline, "delete pipeline"))

	startPipeline := &cobra.Command{
		Use:   "{{alias}} <pipeline>",
		Short: "Restart a stopped pipeline.",
		Long:  "Restart a stopped pipeline.",
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) error {
			if err := env.Client("user").StartPipeline(args[0]); err != nil {
				return errors.Wrap(err, "error from StartPipeline")
			}
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(startPipeline, "start pipeline"))

	stopPipeline := &cobra.Command{
		Use:   "{{alias}} <pipeline>",
		Short: "Stop a running pipeline.",
		Long:  "Stop a running pipeline.",
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) error {
			if err := env.Client("user").StopPipeline(args[0]); err != nil {
				return errors.Wrap(err, "error from StopPipeline")
			}
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(stopPipeline, "stop pipeline"))

	var file string
	createSecret := &cobra.Command{
		Short: "Create a secret on the cluster.",
		Long:  "Create a secret on the cluster.",
		RunE: cmdutil.RunFixedArgs(0, func(args []string, env cmdutil.Env) error {
			fileBytes, err := ioutil.ReadFile(file)
			if err != nil {
				return err
			}

			c := env.Client("user")
			_, err = c.PpsAPIClient.CreateSecret(
				c.Ctx(),
				&ppsclient.CreateSecretRequest{
					File: fileBytes,
				})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	createSecret.Flags().StringVarP(&file, "file", "f", "", "File containing Kubernetes secret.")
	commands = append(commands, cmdutil.CreateAlias(createSecret, "create secret"))

	deleteSecret := &cobra.Command{
		Short: "Delete a secret from the cluster.",
		Long:  "Delete a secret from the cluster.",
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) error {
			c := env.Client("user")
			_, err := c.PpsAPIClient.DeleteSecret(
				c.Ctx(),
				&ppsclient.DeleteSecretRequest{
					Secret: &ppsclient.Secret{
						Name: args[0],
					},
				})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(deleteSecret, "delete secret"))

	inspectSecret := &cobra.Command{
		Short: "Inspect a secret from the cluster.",
		Long:  "Inspect a secret from the cluster.",
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) error {
			c := env.Client("user")
			secretInfo, err := c.PpsAPIClient.InspectSecret(
				c.Ctx(),
				&ppsclient.InspectSecretRequest{
					Secret: &ppsclient.Secret{
						Name: args[0],
					},
				})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}

			writer := tabwriter.NewWriter(env.Out(), pretty.SecretHeader)
			pretty.PrintSecretInfo(writer, secretInfo)
			return writer.Flush()
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(inspectSecret, "inspect secret"))

	listSecret := &cobra.Command{
		Short: "List all secrets from a namespace in the cluster.",
		Long:  "List all secrets from a namespace in the cluster.",
		RunE: cmdutil.RunFixedArgs(0, func(args []string, env cmdutil.Env) error {
			c := env.Client("user")
			secretInfos, err := c.PpsAPIClient.ListSecret(c.Ctx(), &types.Empty{})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}

			writer := tabwriter.NewWriter(env.Out(), pretty.SecretHeader)
			for _, si := range secretInfos.GetSecretInfo() {
				pretty.PrintSecretInfo(writer, si)
			}
			return writer.Flush()
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(listSecret, "list secret"))

	return commands
}

func pipelineHelper(env cmdutil.Env, reprocess bool, pushImages bool, registry, username, pipelinePath string, update bool) error {
	pipelineReader, err := ppsutil.NewPipelineManifestReader(pipelinePath)
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

		if update {
			request.Update = true
			request.Reprocess = reprocess
		}

		if pushImages {
			if request.Transform == nil {
				return errors.New("must specify a pipeline `transform`")
			}
			pipelineParentPath, _ := filepath.Split(pipelinePath)
			if err := dockerBuildHelper(env, request, registry, username, pipelineParentPath); err != nil {
				return err
			}
		}

		if request.Transform != nil && request.Transform.Image != "" {
			if !strings.Contains(request.Transform.Image, ":") {
				fmt.Fprintf(env.Err(),
					"WARNING: please specify a tag for the docker image in your transform.image spec.\n"+
						"For example, change 'python' to 'python:3' or 'bash' to 'bash:5'. This improves\n"+
						"reproducibility of your pipelines.\n\n")
			} else if strings.HasSuffix(request.Transform.Image, ":latest") {
				fmt.Fprintf(env.Err(),
					"WARNING: please do not specify the ':latest' tag for the docker image in your\n"+
						"transform.image spec. For example, change 'python:latest' to 'python:3' or\n"+
						"'bash:latest' to 'bash:5'. This improves reproducibility of your pipelines.\n\n")
			}
		}
		if err = txncmds.WithActiveTransaction(env, func(txClient *pachdclient.APIClient) error {
			// Add trace if env var is set
			ctx, err := extended.EmbedAnyDuration(txClient.Ctx())
			if err != nil {
				logrus.Warning(err)
			}

			_, err = txClient.WithCtx(ctx).PpsAPIClient.CreatePipeline(
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

func dockerBuildHelper(env cmdutil.Env, request *ppsclient.CreatePipelineRequest, registry, username, pipelineParentPath string) error {
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
			reader := bufio.NewReader(env.In())
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
		return "", err
	}
	_, err = gojq.Compile(q)
	if err != nil {
		return "", err
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
