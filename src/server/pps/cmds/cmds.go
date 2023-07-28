package cmds

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	prompt "github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/itchyny/gojq"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"

	pachdclient "github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachctl"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachtmpl"
	"github.com/pachyderm/pachyderm/v2/src/internal/pager"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/tabwriter"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing/extended"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/cmd/pachctl/shell"
	"github.com/pachyderm/pachyderm/v2/src/server/pps/pretty"
	txncmds "github.com/pachyderm/pachyderm/v2/src/server/transaction/cmds"
	workerserver "github.com/pachyderm/pachyderm/v2/src/server/worker/server"
	workerapi "github.com/pachyderm/pachyderm/v2/src/worker"
)

const (
	// Plural variables are used below for user convenience.
	datums    = "datums"
	jobs      = "jobs"
	pipelines = "pipelines"
	secrets   = "secrets"
)

// Cmds returns a slice containing pps commands.
func Cmds(mainCtx context.Context, pachCtx *config.Context, pachctlCfg *pachctl.Config) []*cobra.Command {
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
		Long: "Jobs are the basic units of computation in Pachyderm and are created by pipelines. \n \n" +
			"When created, a job runs a containerized workload over a set of finished input commits; once completed, they  write the output to a commit in the pipeline's output repo. " +
			"Jobs can have multiple datums, each processed independently, with the results merged together at the end. \n \n" +
			"If a job fails, the output commit will not be populated with data.",
	}
	commands = append(commands, cmdutil.CreateDocsAliases(jobDocs, "job", " job$", jobs))

	var project = pachCtx.Project
	inspectJob := &cobra.Command{
		Use:   "{{alias}} <pipeline>@<job>",
		Short: "Return info about a job.",
		Long: "This command returns detailed info about a job, including processing stats, inputs, and transformation configuration (the image and commands used). \n \n" +
			"If you pass in a job set ID (without the `pipeline@`), it will defer you to using the `pachctl list job <id>` command. See examples for proper use. \n \n" +
			"\t- To specify the project where the parent pipeline lives, use the `--project` flag \n" +
			"\t- To specify the output should be raw JSON or YAML, use the `--raw` flag along with `--output`",
		Example: "\t- {{alias}} foo@e0f68a2fcda7458880c9e2e2dae9e678 \n" +
			"\t- {{alias}} foo@e0f68a2fcda7458880c9e2e2dae9e678 --project bar \n" +
			"\t- {{alias}} foo@e0f68a2fcda7458880c9e2e2dae9e678 --project bar --raw --output yaml \n",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			job, err := cmdutil.ParseJob(project, args[0])
			if err != nil && uuid.IsUUIDWithoutDashes(args[0]) {
				return errors.New(`Use "list job <id>" to see jobs with a given ID across different pipelines`)
			} else if err != nil {
				return err
			}
			client, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer client.Close()
			jobInfo, err := client.InspectJob(job.Pipeline.Project.GetName(), job.Pipeline.Name, job.Id, true)
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
	inspectJob.Flags().StringVar(&project, "project", project, "Specify the project (by name) containing the parent pipeline for this job.")
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
		Long:  "This command waits for a job to finish then return info about the job.",
		Example: "\t- {{alias}} e0f68a2fcda7458880c9e2e2dae9e678 \n" +
			"\t- {{alias}} foo@e0f68a2fcda7458880c9e2e2dae9e678 \n" +
			"\t- {{alias}} foo@e0f68a2fcda7458880c9e2e2dae9e678 --project bar \n" +
			"\t- {{alias}} foo@e0f68a2fcda7458880c9e2e2dae9e678 --project bar --raw --output yaml \n",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
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
				job, err := cmdutil.ParseJob(project, args[0])
				if err != nil {
					return err
				}
				jobInfo, err := client.WaitJob(job.Pipeline.Project.GetName(), job.Pipeline.Name, job.Id, true)
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
	waitJob.Flags().StringVar(&project, "project", project, "Specify the project (by name) containing the parent pipeline for this job.")
	shell.RegisterCompletionFunc(waitJob, shell.JobCompletion)
	commands = append(commands, cmdutil.CreateAliases(waitJob, "wait job", jobs))

	var pipelineName string
	var allProjects bool
	var inputCommitStrs []string
	var history string
	var stateStrs []string
	var expand bool
	listJob := &cobra.Command{
		Use:   "{{alias}} [<job-id>]",
		Short: "Return info about jobs.",
		Long: "This command returns info about a list of jobs. You can pass in the command with or without a job ID. \n \n" +
			"Without an ID, this command returns a global list of top-level job sets which contain their own sub-jobs; " +
			"With an ID, it returns a list of sub-jobs within the specified job set. \n \n" +
			"\t- To return a list of sub-jobs across all job sets, use the `--expand` flag without passing an ID \n" +
			"\t- To return only the sub-jobs from the most recent version of a pipeline, use the `--pipeline` flag \n" +
			"\t- To return all sub-jobs from all versions of a pipeline, use the `--history` flag \n" +
			"\t- To return all sub-jobs whose input commits include data from a particular repo branch/commit, use the `--input` flag \n" +
			"\t- To turn only sub-jobs with a particular state, use the `--state` flag; options: CREATED, STARTING, UNRUNNABLE, RUNNING, EGRESS, FINISHING, FAILURE, KILLED, SUCCESS",
		Example: "\t- {{alias}} \n" +
			"\t- {{alias}} --state starting \n" +
			"\t- {{alias}} --pipeline foo \n" +
			"\t- {{alias}} --expand \n" +
			"\t- {{alias}} --expand --pipeline foo \n" +
			"\t- {{alias}} --expand --pipeline foo  --state failure --state unrunnable \n" +
			"\t- {{alias}} 5f93d03b65fa421996185e53f7f8b1e4 \n" +
			"\t- {{alias}} 5f93d03b65fa421996185e53f7f8b1e4 --state running\n" +
			"\t- {{alias}} --input foo-repo@staging \n" +
			"\t- {{alias}} --input foo-repo@5f93d03b65fa421996185e53f7f8b1e4 \n" +
			"\t- {{alias}} --pipeline foo --input bar-repo@staging \n" +
			"\t- {{alias}} --pipeline foo --input bar-repo@5f93d03b65fa421996185e53f7f8b1e4 \n",
		Run: cmdutil.RunBoundedArgs(0, 1, func(args []string) error {
			commits, err := cmdutil.ParseCommits(project, inputCommitStrs)
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

			client, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer client.Close()

			if !raw && output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}

			// To list jobs for all projects, user must be explicit about it.
			// The --project filter takes precedence over everything else.
			// By default use pfs.DefaultProjectName
			projectsFilter := []*pfs.Project{{Name: project}}
			if allProjects {
				projectsFilter = nil
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

					req := &pps.ListJobSetRequest{Projects: projectsFilter}
					listJobSetClient, err := client.PpsAPIClient.ListJobSet(client.Ctx(), req)
					if err != nil {
						return grpcutil.ScrubGRPC(err)
					}

					if raw {
						e := cmdutil.Encoder(output, os.Stdout)
						return grpcutil.ForEach[*pps.JobSetInfo](listJobSetClient, func(jobSetInfo *pps.JobSetInfo) error {
							return errors.EnsureStack(e.EncodeProto(jobSetInfo))
						})
					}

					return pager.Page(noPager, os.Stdout, func(w io.Writer) error {
						writer := tabwriter.NewWriter(w, pretty.JobSetHeader)
						if err := grpcutil.ForEach[*pps.JobSetInfo](listJobSetClient, func(jobSetInfo *pps.JobSetInfo) error {
							pretty.PrintJobSetInfo(writer, jobSetInfo, fullTimestamps)
							return nil
						}); err != nil {
							return err
						}
						return writer.Flush()
					})
				} else {
					// We are listing all sub-jobs, possibly restricted to a single pipeline
					var pipeline *pps.Pipeline
					if pipelineName != "" {
						pipeline = &pps.Pipeline{Name: pipelineName, Project: &pfs.Project{Name: project}}
					}
					req := &pps.ListJobRequest{
						Projects:    projectsFilter,
						Pipeline:    pipeline,
						InputCommit: commits,
						History:     historyCount,
						Details:     true,
						JqFilter:    filter,
					}

					ctx, cf := pctx.WithCancel(client.Ctx())
					defer cf()
					ljClient, err := client.PpsAPIClient.ListJob(ctx, req)
					if err != nil {
						return grpcutil.ScrubGRPC(err)
					}
					if raw {
						e := cmdutil.Encoder(output, os.Stdout)
						return listJobFilterF(ctx, ljClient, func(ji *pps.JobInfo) error {
							return errors.EnsureStack(e.EncodeProto(ji))
						})
					}

					return pager.Page(noPager, os.Stdout, func(w io.Writer) error {
						writer := tabwriter.NewWriter(w, pretty.JobHeader)
						if err := listJobFilterF(ctx, ljClient, func(ji *pps.JobInfo) error {
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
	listJob.Flags().StringVarP(&pipelineName, "pipeline", "p", "", "Specify results should only return jobs created by a given pipeline.")
	listJob.Flags().BoolVarP(&allProjects, "all-projects", "A", false, "Specify results should return jobs from all projects.")
	listJob.Flags().StringVar(&project, "project", project, "Specify the project (by name) containing the parent pipeline for returned jobs.")
	listJob.MarkFlagCustom("pipeline", "__pachctl_get_pipeline")
	listJob.Flags().StringSliceVarP(&inputCommitStrs, "input", "i", []string{}, "Specify results should only return jobs with a specific set of input commits; format: <repo>@<branch-or-commit>")
	listJob.MarkFlagCustom("input", "__pachctl_get_repo_commit")
	listJob.Flags().BoolVarP(&expand, "expand", "x", false, "Specify results return as one line for each sub-job and include more columns; not needed if ID is passed.")
	listJob.Flags().AddFlagSet(outputFlags)
	listJob.Flags().AddFlagSet(timestampFlags)
	listJob.Flags().AddFlagSet(pagerFlags)
	listJob.Flags().StringVar(&history, "history", "none", "Specify results returned include jobs from historical versions of pipelines.")
	listJob.Flags().StringArrayVar(&stateStrs, "state", []string{}, "Specify results return only sub-jobs with the specified state; can be repeated to include multiple states.")
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
		Long:  "This command deletes a job.",
		Example: "\t- {{alias}} 5f93d03b65fa421996185e53f7f8b1e4 \n" +
			"\t- {{alias}} 5f93d03b65fa421996185e53f7f8b1e4 --project foo",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			job, err := cmdutil.ParseJob(project, args[0])
			if err != nil {
				return err
			}
			client, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer client.Close()
			if err := client.DeleteJob(job.Pipeline.Project.GetName(), job.Pipeline.Name, job.Id); err != nil {
				return errors.Wrap(err, "error from DeleteJob")
			}
			return nil
		}),
	}
	deleteJob.Flags().StringVar(&project, "project", project, "Specify the project (by name) containing the parent pipeline for this job.")
	shell.RegisterCompletionFunc(deleteJob, shell.JobCompletion)
	commands = append(commands, cmdutil.CreateAliases(deleteJob, "delete job", jobs))

	stopJob := &cobra.Command{
		Use:   "{{alias}} <pipeline>@<job>",
		Short: "Stop a job.",
		Long: "This command stops a job immediately." +
			"\t- To specify the project where the parent pipeline lives, use the `--project` flag \n",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
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
						if err := tb.StopJob(jobInfo.Job.Pipeline.Project.Name, jobInfo.Job.Pipeline.Name, jobInfo.Job.Id); err != nil {
							return err
						}
					}
					return nil
				}); err != nil {
					return err
				}
			} else {
				job, err := cmdutil.ParseJob(project, args[0])
				if err != nil {
					return err
				}
				if err := client.StopJob(job.Pipeline.Project.Name, job.Pipeline.Name, job.Id); err != nil {
					return errors.Wrap(err, "error from StopProjectJob")
				}
			}
			return nil
		}),
	}
	stopJob.Flags().StringVar(&project, "project", project, "Specify the project (by name) containing the parent pipeline for the job.")
	shell.RegisterCompletionFunc(stopJob, shell.JobCompletion)
	commands = append(commands, cmdutil.CreateAliases(stopJob, "stop job", jobs))

	datumDocs := &cobra.Command{
		Short: "Docs for datums.",
		Long: "Datums are the smallest independent unit of processing for a Job. " +
			"Datums are defined by applying a glob pattern in the pipeline spec to the file paths in an input repo, and they can include any number of files and directories. \n \n" +
			"Datums within a job are processed independently -- and sometimes distributed across separate workers (see `datum_set_spec` and `parallelism_spec` options for pipeline specification).\n \n" +
			"A separate execution of user code will be run for each datum unless datum batching is utilized.",
	}
	commands = append(commands, cmdutil.CreateDocsAliases(datumDocs, "datum", " datum$", datums))

	restartDatum := &cobra.Command{
		Use:   "{{alias}} <pipeline>@<job> <datum-path1>,<datum-path2>,...",
		Short: "Restart a stuck datum during a currently running job.",
		Long: "This command restarts a stuck datum during a currently running job; it does not solve failed datums. \n \n" +
			"You can configure a job to skip failed datums via the transform.err_cmd setting of your pipeline spec. \n \n" +
			"\t- To specify the project where the parent pipeline lives, use the `--project` flag \n",
		Example: "\t- {{alias}} foo@5f93d03b65fa421996185e53f7f8b1e4 /logs/logs.txt \n" +
			"\t- {{alias}} foo@5f93d03b65fa421996185e53f7f8b1e4 /logs/logs-a.txt, /logs/logs-b.txt \n" +
			"\t- {{alias}} foo@5f93d03b65fa421996185e53f7f8b1e4 /logs/logs-a.txt, /logs/logs-b.txt --project bar ",
		Run: cmdutil.RunFixedArgs(2, func(args []string) error {
			job, err := cmdutil.ParseJob(project, args[0])
			if err != nil {
				return err
			}
			client, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
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
			return client.RestartDatum(job.Pipeline.Project.GetName(), job.Pipeline.Name, job.Id, datumFilter)
		}),
	}
	restartDatum.Flags().StringVar(&project, "project", project, "Specify the project (by name) containing parent pipeline for the datum's job")
	commands = append(commands, cmdutil.CreateAliases(restartDatum, "restart datum", datums))

	var pipelineInputPath string
	listDatum := &cobra.Command{
		Use:   "{{alias}} <pipeline>@<job>",
		Short: "Return the datums in a job.",
		Long: "This command returns the datums in a job. \n \n" +
			"\t- To pass in a JSON pipeline spec instead of `pipeline@job`, use the `--file` flag \n " +
			"\t- To specify the project where the parent pipeline lives, use the `--project` flag \n",
		Example: "\t- {{alias}} foo@5f93d03b65fa421996185e53f7f8b1e4 \n" +
			"\t- {{alias}} foo@5f93d03b65fa421996185e53f7f8b1e4 --project bar \n" +
			"\t- {{alias}} --file pipeline.json",
		Run: cmdutil.RunBoundedArgs(0, 1, func(args []string) (retErr error) {
			client, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer client.Close()
			var printF func(*pps.DatumInfo) error
			if !raw {
				if output != "" {
					return errors.New("cannot set --output (-o) without --raw")
				}
				writer := tabwriter.NewWriter(os.Stdout, pretty.DatumHeader)
				printF = func(di *pps.DatumInfo) error {
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
				printF = func(di *pps.DatumInfo) error {
					return errors.EnsureStack(e.EncodeProto(di))
				}
			}
			if pipelineInputPath != "" && len(args) == 1 {
				return errors.Errorf("can't specify both a job and a pipeline spec")
			} else if pipelineInputPath != "" {
				r, err := fileIndicatorToReadCloser(pipelineInputPath)
				if err != nil {
					return err
				}
				defer r.Close()
				pipelineReader, err := ppsutil.NewPipelineManifestReader(r)
				if err != nil {
					return err
				}
				request, err := pipelineReader.NextCreatePipelineRequest()
				if err != nil {
					return err
				}
				if err := pps.VisitInput(request.Input, func(i *pps.Input) error {
					if i.Pfs != nil && i.Pfs.Project == "" {
						i.Pfs.Project = project
					}
					return nil
				}); err != nil {
					return err
				}
				return client.ListDatumInput(request.Input, printF)
			} else if len(args) == 1 {
				job, err := cmdutil.ParseJob(project, args[0])
				if err != nil {
					return err
				}
				return client.ListDatum(job.Pipeline.Project.GetName(), job.Pipeline.Name, job.Id, printF)
			} else {
				return errors.Errorf("must specify either a job or a pipeline spec")
			}
		}),
	}
	listDatum.Flags().StringVarP(&pipelineInputPath, "file", "f", "", "Set the JSON file containing the pipeline to list datums from; the pipeline need not exist.")
	listDatum.Flags().StringVar(&project, "project", project, "Specify the project (by name) containing parent pipeline for the job.")
	listDatum.Flags().AddFlagSet(outputFlags)
	shell.RegisterCompletionFunc(listDatum, shell.JobCompletion)
	commands = append(commands, cmdutil.CreateAliases(listDatum, "list datum", datums))

	var since string
	kubeEvents := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Return the kubernetes events.",
		Long: "This command returns the kubernetes events. \n" +
			"\t- To return results starting from a certain amount of time before now, use the `--since` flag \n" +
			"\t- To return the raw events, use the `--raw` flag \n",
		Example: "\t- {{alias}} --raw \n" +
			"\t- {{alias}} --since 100s \n" +
			"\t- {{alias}} --raw --since 1h \n",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			client, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer client.Close()
			since, err := time.ParseDuration(since)
			if err != nil {
				return errors.Wrapf(err, "parse since(%q)", since)
			}
			request := pps.LokiRequest{
				Since: durationpb.New(since),
			}
			kubeEventsClient, err := client.PpsAPIClient.GetKubeEvents(client.Ctx(), &request)
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			writer := tabwriter.NewWriter(os.Stdout, pretty.KubeEventsHeader)
			if err := grpcutil.ForEach[*pps.LokiLogMessage](kubeEventsClient, func(msg *pps.LokiLogMessage) error {
				if raw {
					fmt.Println(msg.Message)
				} else {
					pretty.PrintKubeEvent(writer, msg.Message)
				}
				return nil
			}); err != nil {
				return err
			}
			return writer.Flush()
		}),
	}
	kubeEvents.Flags().BoolVar(&raw, "raw", false, "Specify results should return log messages verbatim from server.")
	kubeEvents.Flags().StringVar(&since, "since", "0", "Specify results should return log messages more recent than \"since\".")
	commands = append(commands, cmdutil.CreateAlias(kubeEvents, "kube-events"))

	queryLoki := &cobra.Command{
		Use:   "{{alias}} <query>",
		Short: "Query the loki logs.",
		Long:  "This command queries the loki logs.",
		Example: "\t- {{alias}} <query> --since 100s \n" +
			"\t- {{alias}} <query> --since 1h",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			query := args[0]
			client, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer client.Close()
			since, err := time.ParseDuration(since)
			if err != nil {
				return errors.Wrapf(err, "parse since(%q)", since)
			}
			request := pps.LokiRequest{
				Query: query,
				Since: durationpb.New(since),
			}
			lokiClient, err := client.PpsAPIClient.QueryLoki(client.Ctx(), &request)
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			if err := grpcutil.ForEach[*pps.LokiLogMessage](lokiClient, func(log *pps.LokiLogMessage) error {
				fmt.Println(log.Message)
				return nil
			}); err != nil {
				return err
			}
			return nil
		}),
	}
	queryLoki.Flags().StringVar(&since, "since", "0", "Specify results should return log messages more recent than \"since\".")
	commands = append(commands, cmdutil.CreateAlias(queryLoki, "loki"))

	inspectDatum := &cobra.Command{
		Use:   "{{alias}} <pipeline>@<job> <datum>",
		Short: "Display detailed info about a single datum.",
		Long:  "This command displays detailed info about a single datum; requires the pipeline to have stats enabled.",
		Example: "\t- {{alias}} foo@5f93d03b65fa421996185e53f7f8b1e4 7f3cd988429894000bdad549dfe2d09b5ca7bfc5083b79fec0e6bda3db8cc705 \n" +
			"\t- {{alias}} foo@5f93d03b65fa421996185e53f7f8b1e4 7f3cd988429894000bdad549dfe2d09b5ca7bfc5083b79fec0e6bda3db8cc705 --project foo",
		Run: cmdutil.RunFixedArgs(2, func(args []string) error {
			job, err := cmdutil.ParseJob(project, args[0])
			if err != nil {
				return err
			}
			client, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer client.Close()
			datumInfo, err := client.InspectDatum(job.Pipeline.Project.GetName(), job.Pipeline.Name, job.Id, args[1])
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
	inspectDatum.Flags().StringVar(&project, "project", project, "Project containing the job")
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
		Long: "This command returns logs from a job. \n" +
			"\t- To filter your logs by pipeline, use the `--pipeline` flag \n" +
			"\t- To filter your logs by job, use the `--job` flag \n" +
			"\t- To filter your logs by datum, use the `--datum` flag \n" +
			"\t- To filter your logs by the master process, use the `--master` flag  with the `--pipeline` flag \n" +
			"\t- To filter your logs by the worker process, use the `--worker` flag \n" +
			"\t- To follow the logs as more are created, use the `--follow` flag \n" +
			"\t- To set the number of lines to return, use the `--tail` flag \n" +
			"\t- To return results starting from a certain amount of time before now, use the `--since` flag \n",
		Example: "\t- {{alias}} --pipeline foo \n" +
			"\t- {{alias}} --job foo@5f93d03b65fa421996185e53f7f8b1e4 \n" +
			"\t- {{alias}} --job foo@5f93d03b65fa421996185e53f7f8b1e4 --tail 10 \n" +
			"\t- {{alias}} --job foo@5f93d03b65fa421996185e53f7f8b1e4 --follow \n" +
			"\t- {{alias}} --job foo@5f93d03b65fa421996185e53f7f8b1e4 --datum 7f3c[...] \n" +
			"\t- {{alias}} --pipeline foo --datum 7f3c[...] --master \n" +
			"\t- {{alias}} --pipeline foo --datum 7f3c[...] --worker  \n" +
			"\t- {{alias}} --pipeline foo --datum 7f3c[...] --master --tail 10  \n" +
			"\t- {{alias}} --pipeline foo --datum 7f3c[...] --worker --follow \n",
		Run: cmdutil.RunFixedArgsCmd(0, func(cmd *cobra.Command, args []string) error {
			client, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
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
				return errors.Wrapf(err, "error parsing since (%q)", since)
			}
			if tail != 0 {
				return errors.Errorf("tail has been deprecated and removed from Pachyderm, use --since instead")
			}

			if pipelineName != "" && jobStr != "" {
				return errors.Errorf("only one of pipeline or job should be specified")
			}

			var jobID string
			if jobStr != "" {
				job, err := cmdutil.ParseJob(project, jobStr)
				if err != nil {
					return err
				}
				pipelineName = job.Pipeline.Name
				jobID = job.Id
			}

			// Issue RPC
			if !cmd.Flags().Changed("since") {
				since = 0
			}
			iter := client.GetLogs(project, pipelineName, jobID, data, datumID, master, follow, since)
			for iter.Next() {
				if raw {
					fmt.Println(protojson.Format(iter.Message()))
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
	getLogs.Flags().StringVarP(&pipelineName, "pipeline", "p", "", "Specify results should only return logs for a given pipeline.")
	getLogs.MarkFlagCustom("pipeline", "__pachctl_get_pipeline")
	getLogs.Flags().StringVarP(&jobStr, "job", "j", "", "Specify results should only return logs for a given job ID.")
	getLogs.MarkFlagCustom("job", "__pachctl_get_job")
	getLogs.Flags().StringVar(&datumID, "datum", "", "Specify results should only return logs for a given datum ID.")
	getLogs.Flags().StringVar(&commaInputs, "inputs", "", "Filter for log lines generated while processing these files (accepts PFS paths or file hashes)")
	getLogs.Flags().BoolVar(&master, "master", false, "Specify results should only return logs from the master process; --pipeline must be set.")
	getLogs.Flags().BoolVar(&worker, "worker", false, "Specify results should only return logs from the worker process.")
	getLogs.Flags().BoolVar(&raw, "raw", false, "Specify results should only return log messages verbatim from server.")
	getLogs.Flags().BoolVarP(&follow, "follow", "f", false, "Follow logs as more are created.")
	getLogs.Flags().Int64VarP(&tail, "tail", "t", 0, "Set the number of lines to return of the most recent logs.")
	getLogs.Flags().StringVar(&since, "since", "24h", "Specify results should return log messages more recent than \"since\".")
	getLogs.Flags().StringVar(&project, "project", project, "Specify the project (by name) containing parent pipeline for the job.")
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
		Long: "Pipelines are a powerful abstraction for automating jobs. They take a set of repos and branches as inputs and write to a single output repo of the same name. \n" +
			"Pipelines then subscribe to commits on those repos and launch a job to process each incoming commit. All jobs created by a pipeline will create commits in the pipeline's output repo.			",
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
		Long: "This command creates a new pipeline from a pipeline specification. \n \n" +
			"You can create a pipeline using a JSON/YAML file or a jsonnet template file -- via either a local filepath or URL. Multiple pipelines can be created from one file." +
			"For details on the format, see https://docs.pachyderm.com/latest/reference/pipeline_spec/. \n \n" +
			"\t- To create a pipeline from a JSON/YAML file, use the `--file` flag \n" +
			"\t- To create a pipeline from a jsonnet template file, use the `--jsonnet` flag; you can optionally pay multiple arguments separately using `--arg` \n" +
			"\t- To push your local images to docker registry, use the `--push-images` and `--username` flags \n" +
			"\t- To push your local images to custom registry, use the `--push-images`, `--registry`, and `--username` flags \n",
		Example: "\t {{alias}} -file regression.json \n" +
			"\t {{alias}} -file foo.json --project bar \n" +
			"\t {{alias}} -file foo.json --push-images --username lbliii \n" +
			"\t {{alias}} --jsonnet /templates/foo.jsonnet --arg myimage=bar --arg src=image \n",
		Run: cmdutil.RunFixedArgs(0, func(args []string) (retErr error) {
			return pipelineHelper(mainCtx, pachctlCfg, false, pushImages, registry, username, project, pipelinePath, jsonnetPath, jsonnetArgs, false)
		}),
	}
	createPipeline.Flags().StringVarP(&pipelinePath, "file", "f", "", "Provide a JSON/YAML file (url or filepath) for one or more pipelines. \"-\" reads from stdin (the default behavior). Exactly one of --file and --jsonnet must be set.")
	createPipeline.Flags().StringVar(&jsonnetPath, "jsonnet", "", "Provide a Jsonnet template file (url or filepath) for one or more pipelines. \"-\" reads from stdin. Exactly one of --file and --jsonnet must be set. Jsonnet templates must contain a top-level function; strings can be passed to this function with --arg (below)")
	createPipeline.Flags().StringArrayVar(&jsonnetArgs, "arg", nil, "Provide a top-level argument in the form of 'param=value' passed to the Jsonnet template; requires --jsonnet. For multiple args, --arg may be set more than once.")
	createPipeline.Flags().BoolVarP(&pushImages, "push-images", "p", false, "Specify that the local docker images should be pushed into the registry (docker by default).")
	createPipeline.Flags().StringVarP(&registry, "registry", "r", "index.docker.io", "Specify an alternative registry to push images to.")
	createPipeline.Flags().StringVarP(&username, "username", "u", "", "Specify the username to push images as.")
	createPipeline.Flags().StringVar(&project, "project", project, "Specify the project (by name) in which to create the pipeline.")
	commands = append(commands, cmdutil.CreateAliases(createPipeline, "create pipeline", pipelines))

	var reprocess bool
	updatePipeline := &cobra.Command{
		Short: "Update an existing Pachyderm pipeline.",
		Long: "This command updates a Pachyderm pipeline with a new pipeline specification. For details on the format, see https://docs.pachyderm.com/latest/reference/pipeline-spec/ \n \n" +
			"\t- To update a pipeline from a JSON/YAML file, use the `--file` flag \n" +
			"\t- To update a pipeline from a jsonnet template file, use the `--jsonnet` flag. You can optionally pay multiple arguments separately using `--arg` \n" +
			"\t- To reprocess all data in the pipeline, use the `--reprocess` flag \n" +
			"\t- To push your local images to docker registry, use the `--push-images` and `--username` flags \n" +
			"\t- To push your local images to custom registry, use the `--push-images`, `--registry`, and `--username` flags \n",
		Example: "\t {{alias}} -file regression.json \n" +
			"\t {{alias}} -file foo.json --project bar \n" +
			"\t {{alias}} -file foo.json --push-images --username lbliii \n" +
			"\t {{alias}} --jsonnet /templates/foo.jsonnet --arg myimage=bar --arg src=image \n",
		Run: cmdutil.RunFixedArgs(0, func(args []string) (retErr error) {
			return pipelineHelper(mainCtx, pachctlCfg, reprocess, pushImages, registry, username, project, pipelinePath, jsonnetPath, jsonnetArgs, true)
		}),
	}
	updatePipeline.Flags().StringVarP(&pipelinePath, "file", "f", "", "Provide a JSON/YAML file (url or filepath) for one or more pipelines. \"-\" reads from stdin (the default behavior). Exactly one of --file and --jsonnet must be set.")
	updatePipeline.Flags().StringVar(&jsonnetPath, "jsonnet", "", "Provide a Jsonnet template file (url or filepath) for one or more pipelines. \"-\" reads from stdin. Exactly one of --file and --jsonnet must be set. Jsonnet templates must contain a top-level function; strings can be passed to this function with --arg (below)")
	updatePipeline.Flags().StringArrayVar(&jsonnetArgs, "arg", nil, "Provide a top-level argument in the form of 'param=value' passed to the Jsonnet template; requires --jsonnet. For multiple args, --arg may be set more than once.")
	updatePipeline.Flags().BoolVarP(&pushImages, "push-images", "p", false, "Specify that the local docker images should be pushed into the registry (docker by default).")
	updatePipeline.Flags().StringVarP(&registry, "registry", "r", "index.docker.io", "Specify an alternative registry to push images to.")
	updatePipeline.Flags().StringVarP(&username, "username", "u", "", "Specify the username to push images as.")
	updatePipeline.Flags().BoolVar(&reprocess, "reprocess", false, "Reprocess all datums that were already processed by previous version of the pipeline.")
	updatePipeline.Flags().StringVar(&project, "project", project, "Specify the project (by name) in which to create the pipeline.")
	commands = append(commands, cmdutil.CreateAliases(updatePipeline, "update pipeline", pipelines))

	runCron := &cobra.Command{
		Use:   "{{alias}} <pipeline>",
		Short: "Run an existing Pachyderm cron pipeline now",
		Long:  "This command runs an existing Pachyderm cron pipeline immediately.",
		Example: "\t- {{alias}} foo \n" +
			"\t- {{alias}} foo  --project bar \n",
		Run: cmdutil.RunMinimumArgs(1, func(args []string) (retErr error) {
			client, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer client.Close()
			err = client.RunCron(project, args[0])
			if err != nil {
				return err
			}
			return nil
		}),
	}
	runCron.Flags().StringVar(&project, "project", project, "Specify the project (by name) containing the cron pipeline.")
	commands = append(commands, cmdutil.CreateAlias(runCron, "run cron"))

	inspectPipeline := &cobra.Command{
		Use:   "{{alias}} <pipeline>",
		Short: "Return info about a pipeline.",
		Long:  "This command returns info about a pipeline.",
		Example: "\t- {{alias}} foo \n" +
			"\t- {{alias}} foo --project bar \n" +
			"\t- {{alias}} foo --project bar --raw -o yaml \n",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer client.Close()
			pipelineInfo, err := client.InspectPipeline(project, args[0], true)
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
	inspectPipeline.Flags().StringVar(&project, "project", project, "Specify the project (by name) containing the inspected pipeline.")
	commands = append(commands, cmdutil.CreateAliases(inspectPipeline, "inspect pipeline", pipelines))

	var editor string
	var editorArgs []string
	editPipeline := &cobra.Command{
		Use:   "{{alias}} <pipeline>",
		Short: "Edit the manifest for a pipeline in your text editor.",
		Long:  "This command edits the manifest for a pipeline in your text editor.",
		Example: "\t- {{alias}} foo \n" +
			"\t- {{alias}} foo --project bar \n" +
			"\t- {{alias}} foo --project bar --editor vim \n" +
			"\t- {{alias}} foo --project bar --editor vim --output yaml \n" +
			"\t- {{alias}} foo --project bar --editor vim --reprocess \n",

		Run: cmdutil.RunFixedArgs(1, func(args []string) (retErr error) {
			client, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer client.Close()

			pipelineInfo, err := client.InspectPipeline(project, args[0], true)
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
			r, err := fileIndicatorToReadCloser(f.Name())
			if err != nil {
				return err
			}
			defer r.Close()
			pipelineReader, err := ppsutil.NewPipelineManifestReader(r)
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
			// May not change the project name, but if it is omitted
			// then it is considered unchanged.
			project := request.Pipeline.GetProject()
			projectName := project.GetName()
			if projectName != "" && projectName != pipelineInfo.Pipeline.GetProject().GetName() {
				return errors.New("may not change project name")
			}
			request.Pipeline.Project = pipelineInfo.Pipeline.GetProject() // in case of empty project
			// Likewise, may not change the pipeline name, but if it
			// is omitted then it is considered unchanged.
			pipelineName := request.Pipeline.GetName()
			if pipelineName != "" && pipelineName != pipelineInfo.Pipeline.GetName() {
				return errors.New("may not change pipeline name")
			}
			request.Pipeline.Name = pipelineInfo.Pipeline.GetName() // in case of empty pipeline name
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
	editPipeline.Flags().StringVar(&editor, "editor", "", "Specify the editor to use for modifying the manifest.")
	editPipeline.Flags().StringVarP(&output, "output", "o", "", "Specify the output format: \"json\" or \"yaml\" (default \"json\")")
	editPipeline.Flags().StringVar(&project, "project", project, "Specify the project (by name) containing pipeline to edit.")
	commands = append(commands, cmdutil.CreateAliases(editPipeline, "edit pipeline", pipelines))

	var spec bool
	var commit string
	listPipeline := &cobra.Command{
		Use:   "{{alias}} [<pipeline>]",
		Short: "Return info about all pipelines.",
		Long: "This command returns information about all pipelines. \n \n" +
			"\t- To return pipelines with a specific state, use the `--state` flag \n" +
			"\t- To return pipelines as they existed at a specific commit, use the `--commit` flag \n" +
			"\t- To return a history of pipeline revisions, use the `--history` flag \n",
		Example: "\t- {{alias}} \n" +
			"\t- {{alias}} --spec --output yaml \n" +
			"\t- {{alias}} --commit 5f93d03b65fa421996185e53f7f8b1e4 \n" +
			"\t- {{alias}} --state crashing \n" +
			"\t- {{alias}} --project foo \n" +
			"\t- {{alias}} --project foo --state restarting \n",
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
			client, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return errors.Wrapf(err, "error connecting to pachd")
			}
			defer client.Close()
			var pipeline string
			if len(args) > 0 {
				pipeline = args[0]
			}
			projectsFilter := []*pfs.Project{{Name: project}}
			if allProjects {
				projectsFilter = nil
			}
			request := &pps.ListPipelineRequest{
				History:   history,
				CommitSet: &pfs.CommitSet{Id: commit},
				JqFilter:  filter,
				Details:   true,
				Projects:  projectsFilter,
			}
			if pipeline != "" {
				request.Pipeline = pachdclient.NewPipeline(project, pipeline)
			}
			lpClient, err := client.PpsAPIClient.ListPipeline(client.Ctx(), request)
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			pipelineInfos, err := grpcutil.Collect[*pps.PipelineInfo](lpClient, 1000)
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
	listPipeline.Flags().StringVar(&history, "history", "none", "Specify results should include revision history for pipelines.")
	listPipeline.Flags().StringVarP(&commit, "commit", "c", "", "List the pipelines as they existed at this commit.")
	listPipeline.Flags().StringArrayVar(&stateStrs, "state", []string{}, "Specify results should include only pipelines with the specified state (starting, running, restarting, failure, paused, standby, crashing); can be repeated for multiple states.")
	listPipeline.Flags().StringVar(&project, "project", project, "Specify the project (by name) containing the pipelines.")
	listPipeline.Flags().BoolVarP(&allProjects, "all-projects", "A", false, "Show pipelines form all projects.")
	commands = append(commands, cmdutil.CreateAliases(listPipeline, "list pipeline", pipelines))

	var commitSet string
	var boxWidth int
	var edgeHeight int
	draw := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Draw a DAG",
		Long:  "This command draws a DAG",
		Example: "\t- {{alias}} --commit 5f93d03b65fa421996185e53f7f8b1e4" +
			"\t- {{alias}} --box-width 20" +
			"\t- {{alias}} --edge-height 8" +
			"\t- {{alias}} --project foo" +
			"\t- {{alias}} --box-width 20 --edge-height 8 --commit 5f93d03b65fa421996185e53f7f8b1e4",
		Run: cmdutil.RunBoundedArgs(0, 1, func(args []string) error {
			client, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return errors.Wrapf(err, "error connecting to pachd")
			}
			defer client.Close()
			request := &pps.ListPipelineRequest{
				History:   0,
				JqFilter:  "",
				Details:   true,
				CommitSet: &pfs.CommitSet{Id: commitSet},
				Projects:  []*pfs.Project{{Name: project}},
			}
			lpClient, err := client.PpsAPIClient.ListPipeline(client.Ctx(), request)
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			pipelineInfos, err := grpcutil.Collect[*pps.PipelineInfo](lpClient, 1000)
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			if picture, err := pretty.Draw(pipelineInfos, pretty.BoxWidthOption(boxWidth), pretty.EdgeHeightOption(edgeHeight)); err != nil {
				return err
			} else {
				fmt.Print(picture)
			}
			return nil
		}),
	}
	draw.Flags().StringVarP(&commitSet, "commit", "c", "", "Commit at which you would to draw the DAG")
	draw.Flags().IntVar(&boxWidth, "box-width", 11, "Character width of each box in the DAG")
	draw.Flags().IntVar(&edgeHeight, "edge-height", 5, "Number of vertical lines spanned by each edge")
	draw.Flags().StringVar(&project, "project", project, "Project containing pipelines.")
	commands = append(commands, cmdutil.CreateAlias(draw, "draw pipeline"))

	var (
		all      bool
		force    bool
		keepRepo bool
	)
	deletePipeline := &cobra.Command{
		Use:   "{{alias}} (<pipeline>|--all)",
		Short: "Delete a pipeline.",
		Long:  "This command deletes a pipeline.",
		Example: "\t- {{alias}} foo" +
			"\t- {{alias}} --all" +
			"\t- {{alias}} foo --force" +
			"\t- {{alias}} foo --keep-repo" +
			"\t- {{alias}} foo --project bar --keep-repo",
		Run: cmdutil.RunBoundedArgs(0, 1, func(args []string) error {
			client, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
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
			if all {
				if _, err := client.PpsAPIClient.DeletePipelines(client.Ctx(), &pps.DeletePipelinesRequest{
					Projects: []*pfs.Project{{Name: project}},
					All:      allProjects,
				}); err != nil {
					return grpcutil.ScrubGRPC(err)
				}
				return nil
			}
			if allProjects {
				return errors.Errorf("--allProjects only valid with --all")
			}
			req := &pps.DeletePipelineRequest{
				Force:    force,
				KeepRepo: keepRepo,
			}
			if len(args) > 0 {
				req.Pipeline = pachdclient.NewPipeline(project, args[0])
			}
			if _, err = client.PpsAPIClient.DeletePipeline(client.Ctx(), req); err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			return nil
		}),
	}
	deletePipeline.Flags().BoolVar(&all, "all", false, "Delete all pipelines")
	deletePipeline.Flags().BoolVarP(&force, "force", "f", false, "Delete the pipeline regardless of errors; use with care")
	deletePipeline.Flags().BoolVar(&keepRepo, "keep-repo", false, "Specify that the pipeline's output repo should be saved after pipeline deletion; to reuse this pipeline's name, you'll also need to delete this output repo.")
	deletePipeline.Flags().StringVar(&project, "project", project, "Specify the project (by name) containing project")
	deletePipeline.Flags().BoolVarP(&allProjects, "all-projects", "A", false, "Delete pipelines from all projects; only valid with --all")
	commands = append(commands, cmdutil.CreateAliases(deletePipeline, "delete pipeline", pipelines))

	startPipeline := &cobra.Command{
		Use:   "{{alias}} <pipeline>",
		Short: "Restart a stopped pipeline.",
		Long:  "This command restarts a stopped pipeline.",
		Example: "\t- {{alias}} foo \n" +
			"\t- {{alias}} foo --project bar \n",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer client.Close()
			if err := client.StartPipeline(project, args[0]); err != nil {
				return errors.Wrap(err, "error from StartProjectPipeline")
			}
			return nil
		}),
	}
	startPipeline.Flags().StringVar(&project, "project", project, "Project containing pipeline.")
	commands = append(commands, cmdutil.CreateAliases(startPipeline, "start pipeline", pipelines))

	stopPipeline := &cobra.Command{
		Use:   "{{alias}} <pipeline>",
		Short: "Stop a running pipeline.",
		Long:  "This command stops a running pipeline.",
		Example: "\t- {{alias}} foo \n" +
			"\t- {{alias}} foo --project bar \n",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer client.Close()
			if err := client.StopPipeline(project, args[0]); err != nil {
				return errors.Wrap(err, "error from StopProjectPipeline")
			}
			return nil
		}),
	}
	stopPipeline.Flags().StringVar(&project, "project", project, "Project containing pipeline.")
	commands = append(commands, cmdutil.CreateAliases(stopPipeline, "stop pipeline", pipelines))

	var file string
	createSecret := &cobra.Command{
		Short:   "Create a secret on the cluster.",
		Long:    "This command creates a secret on the cluster.",
		Example: "\t- {{alias}} --file my-secret.json",
		Run: cmdutil.RunFixedArgs(0, func(args []string) (retErr error) {
			client, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
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
				&pps.CreateSecretRequest{
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
		Short:   "Delete a secret from the cluster.",
		Long:    "This command deletes a secret from the cluster.",
		Example: "\t- {{alias}} my-secret \n",
		Run: cmdutil.RunFixedArgs(1, func(args []string) (retErr error) {
			client, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer client.Close()

			_, err = client.PpsAPIClient.DeleteSecret(
				client.Ctx(),
				&pps.DeleteSecretRequest{
					Secret: &pps.Secret{
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
		Short:   "Inspect a secret from the cluster.",
		Long:    "This command inspects a secret from the cluster.",
		Example: "\t- {{alias}} my-secret \n",
		Run: cmdutil.RunFixedArgs(1, func(args []string) (retErr error) {
			client, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer client.Close()

			secretInfo, err := client.PpsAPIClient.InspectSecret(
				client.Ctx(),
				&pps.InspectSecretRequest{
					Secret: &pps.Secret{
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
		Short:   "List all secrets from a namespace in the cluster.",
		Long:    "This command lists all secrets from a namespace in the cluster.",
		Example: "\t- {{alias}} \n",
		Run: cmdutil.RunFixedArgs(0, func(args []string) (retErr error) {
			client, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer client.Close()

			secretInfos, err := client.PpsAPIClient.ListSecret(
				client.Ctx(),
				&emptypb.Empty{},
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
	var stateID string
	runLoadTest := &cobra.Command{
		Use:   "{{alias}} <spec-file> ",
		Short: "Run a PPS load test.",
		Long: "This command runs a PPS load test for a specified pipeline specification file. \n" +
			"\t- To run a load test with a specific seed, use the `--seed` flag \n" +
			"\t- To run a load test with a specific parallelism count, use the `--parallelism` flag \n" +
			"\t- To run a load test with a specific pod patch, use the `--pod-patch` flag",
		Example: "\t- {{alias}} --dag myspec.json \n" +
			"\t- {{alias}} --dag myspec.json --seed 1 \n" +
			"\t- {{alias}} --dag myspec.json  --parallelism 3 \n" +
			"\t- {{alias}} --dag myspec.json  --pod-patch patch.json \n" +
			"\t- {{alias}} --dag myspec.json --state-id xyz\n",
		Run: cmdutil.RunBoundedArgs(0, 1, func(args []string) (retErr error) {
			c, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer func() {
				if err := c.Close(); retErr == nil {
					retErr = err
				}
			}()
			if len(args) == 0 {
				resp, err := c.PpsAPIClient.RunLoadTestDefault(c.Ctx(), &emptypb.Empty{})
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
					StateId:     stateID,
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
	runLoadTest.Flags().StringVarP(&dagSpecFile, "dag", "d", "", "Provide DAG specification file to use for the load test")
	runLoadTest.Flags().Int64VarP(&seed, "seed", "s", 0, "Specify the seed to use for generating the load.")
	runLoadTest.Flags().Int64VarP(&parallelism, "parallelism", "p", 0, "Set the parallelism count to use for the pipelines.")
	runLoadTest.Flags().StringVarP(&podPatchFile, "pod-patch", "", "", "Provide pod patch file to use for the pipelines.")
	runLoadTest.Flags().StringVar(&stateID, "state-id", "", "Provide the ID of the base state to use for the load.")
	commands = append(commands, cmdutil.CreateAlias(runLoadTest, "run pps-load-test"))

	var errStr string
	nextDatum := &cobra.Command{
		Use:     "{{alias}}",
		Short:   "Used internally for datum batching",
		Long:    "This command is used internally for datum batching",
		Example: "\t- {{alias}}",
		Run: cmdutil.Run(func(_ []string) error {
			c, err := workerserver.NewClient("127.0.0.1")
			if err != nil {
				return err
			}
			defer c.Close()
			// TODO: Decide how to handle the environment variables in the response.
			_, err = c.NextDatum(context.Background(), &workerapi.NextDatumRequest{Error: errStr})
			return err
		}),
		Hidden: true,
	}
	nextDatum.Flags().StringVar(&errStr, "error", "", "A string representation of an error that occurred while processing the current datum.")
	commands = append(commands, cmdutil.CreateAlias(nextDatum, "next datum"))

	validatePipeline := &cobra.Command{
		Use:     "{{alias}}",
		Short:   "Validate pipeline spec.",
		Long:    "This command validates a pipeline spec.  Client-side only; does not check that repos, images, etc exist on the server.",
		Example: "\t- {{alias}} --file spec.json",
		Run: cmdutil.RunFixedArgs(0, func(_ []string) error {
			r, err := fileIndicatorToReadCloser(pipelinePath)
			if err != nil {
				return err
			}
			defer r.Close()
			pr, err := ppsutil.NewPipelineManifestReader(r)
			if err != nil {
				return err
			}
			for {
				_, err := pr.NextCreatePipelineRequest()
				if errors.Is(err, io.EOF) {
					return nil
				} else if err != nil {
					return err
				}
			}
		}),
	}
	validatePipeline.Flags().StringVarP(&pipelinePath, "file", "f", "", "A JSON file (url or filepath) containing one or more pipelines. \"-\" reads from stdin (the default behavior). Exactly one of --file and --jsonnet must be set.")
	commands = append(commands, cmdutil.CreateAliases(validatePipeline, "validate pipeline", pipelines))

	var cluster bool
	inspectDefaults := &cobra.Command{
		Use:   "{{alias}} [--cluster | --project PROJECT]",
		Short: "Return defaults.",
		Long:  "Return cluster or project defaults.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			client, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer client.Close()
			if cluster {
				resp, err := client.PpsAPIClient.GetClusterDefaults(mainCtx, &pps.GetClusterDefaultsRequest{})
				if err != nil {
					return errors.Wrap(err, "could not get cluster defaults")
				}
				fmt.Println(resp.ClusterDefaultsJson)
				return nil
			}
			return errors.New("--cluster must be specified")
		}),
		Hidden: true,
	}
	inspectDefaults.Flags().BoolVar(&cluster, "cluster", false, "Inspect cluster defaults.")
	//inspectDefaults.Flags().StringVar(&project, "project", project, "Inspect project defaults.")
	commands = append(commands, cmdutil.CreateAliases(inspectDefaults, "inspect defaults"))

	var pathname string
	createDefaults := &cobra.Command{
		Use:   "{{alias}} [--cluster]",
		Short: "Set cluster defaults.",
		Long:  "Set cluster defaults.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			if cluster {
				r, err := fileIndicatorToReadCloser(pathname)
				if err != nil {
					return errors.Wrapf(err, "could not open path %q for reading", pathname)
				}
				b, err := io.ReadAll(r)
				if err != nil {
					return errors.Wrapf(err, "could not read from %q", pathname)
				}
				b = bytes.TrimSpace(b) // remove leading & trailing whitespace
				// validate that the provided defaults parse
				var cd pps.ClusterDefaults
				if err := protojson.Unmarshal(b, &cd); err != nil {
					return errors.Wrapf(err, "invalid cluster defaults")
				}
				var req pps.SetClusterDefaultsRequest
				req.ClusterDefaultsJson = string(b)

				client, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
				if err != nil {
					return err
				}
				defer client.Close()

				if _, err := client.PpsAPIClient.SetClusterDefaults(mainCtx, &pps.SetClusterDefaultsRequest{
					ClusterDefaultsJson: string(b),
				}); err != nil {
					return errors.Wrap(err, "could not set cluster defaults")
				}
				return nil
			}
			return errors.New("--cluster must be specified")
		}),
		Hidden: true,
	}
	createDefaults.Flags().BoolVar(&cluster, "cluster", false, "Create cluster defaults.")
	createDefaults.Flags().StringVarP(&pathname, "file", "f", "-", "A JSON file containing cluster defaults.  \"-\" reads from stdin (the default behavior.)")
	commands = append(commands, cmdutil.CreateAliases(createDefaults, "create defaults"))

	deleteDefaults := &cobra.Command{
		Use:   "{{alias}} [--cluster]",
		Short: "Delete defaults.",
		Long:  "Delete defaults.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			if cluster {
				client, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
				if err != nil {
					return err
				}
				defer client.Close()

				if _, err := client.PpsAPIClient.SetClusterDefaults(mainCtx, &pps.SetClusterDefaultsRequest{
					ClusterDefaultsJson: "{}",
				}); err != nil {
					return errors.Wrap(err, "could not set cluster defaults")
				}
				return nil
			}
			return errors.New("--cluster must be specified")
		}),
		Hidden: true,
	}
	deleteDefaults.Flags().BoolVar(&cluster, "cluster", false, "Delete cluster defaults.")
	commands = append(commands, cmdutil.CreateAliases(deleteDefaults, "delete defaults"))

	updateDefaults := &cobra.Command{
		Use:   "{{alias}} [--cluster]",
		Short: "Update defaults.",
		Long:  "Update defaults.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			if cluster {
				r, err := fileIndicatorToReadCloser(pathname)
				if err != nil {
					return errors.Wrapf(err, "could not open path %q for reading", pathname)
				}
				b, err := io.ReadAll(r)
				if err != nil {
					return errors.Wrapf(err, "could not read from %q", pathname)
				}
				b = bytes.TrimSpace(b) // remove leading & trailing whitespace
				// validate that the provided defaults parse
				var cd pps.ClusterDefaults
				if err := protojson.Unmarshal(b, &cd); err != nil {
					return errors.Wrapf(err, "invalid cluster defaults")
				}
				var req pps.SetClusterDefaultsRequest
				req.ClusterDefaultsJson = string(b)
				client, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
				if err != nil {
					return err
				}
				defer client.Close()
				if _, err := client.PpsAPIClient.SetClusterDefaults(mainCtx, &pps.SetClusterDefaultsRequest{
					ClusterDefaultsJson: string(b),
				}); err != nil {
					return errors.Wrap(err, "could not set cluster defaults")
				}
				return nil
			}
			return errors.New("--cluster must be specified")
		}),
		Hidden: true,
	}
	updateDefaults.Flags().BoolVar(&cluster, "cluster", false, "Update cluster defaults.")
	updateDefaults.Flags().StringVarP(&pathname, "file", "f", "-", "A JSON file containing cluster defaults.  \"-\" reads from stdin (the default behavior.)")
	commands = append(commands, cmdutil.CreateAliases(updateDefaults, "update defaults"))

	return commands
}

// fileIndicatorToReadCloser returns an IO reader for a file based on an indicator
// (which may be a local path, a remote URL or "-" for stdin).
//
// TODO(msteffen) This is very similar to readConfigBytes in
// s/s/identity/cmds/cmds.go (which differs only in not supporting URLs),
// so the two could perhaps be refactored.
func fileIndicatorToReadCloser(indicator string) (io.ReadCloser, error) {
	if indicator == "-" {
		cmdutil.PrintStdinReminder()
		return os.Stdin, nil
	} else if u, err := url.Parse(indicator); err == nil && u.Scheme != "" {
		resp, err := http.Get(u.String())
		if err != nil {
			return nil, errors.Wrapf(err, "could not retrieve URL %v", u)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, errors.Wrapf(err, "cannot handle HTTP status code %s (%d)", resp.Status, resp.StatusCode)
		}
		return resp.Body, nil
	}
	f, err := os.Open(indicator)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open path %q", indicator)
	}
	return f, nil
}

func evaluateJsonnetTemplate(client *pachdclient.APIClient, jsonnetPath string, jsonnetArgs []string) ([]byte, error) {
	r, err := fileIndicatorToReadCloser(jsonnetPath)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	templateBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read Jsonnet file %q", jsonnetPath)
	}
	args, err := pachtmpl.ParseArgs(jsonnetArgs)
	if err != nil {
		return nil, err
	}
	res, err := client.RenderTemplate(client.Ctx(), &pps.RenderTemplateRequest{
		Template: string(templateBytes),
		Args:     args,
	})
	if err != nil {
		return nil, err
	}
	return []byte(res.Json), nil
}

func pipelineHelper(ctx context.Context, pachctlCfg *pachctl.Config, reprocess bool, pushImages bool, registry, username, project, pipelinePath, jsonnetPath string, jsonnetArgs []string, update bool) error {
	// validate arguments
	if pipelinePath != "" && jsonnetPath != "" {
		return errors.New("cannot set both --file and --jsonnet; exactly one must be set")
	}
	if pipelinePath == "" && jsonnetPath == "" {
		pipelinePath = "-" // default input
	}
	pc, err := pachctlCfg.NewOnUserMachine(ctx, false)
	if err != nil {
		return errors.Wrapf(err, "error connecting to pachd")
	}
	defer pc.Close()
	// read/compute pipeline spec(s) (file, stdin, url, or via template)
	var pipelineReader *ppsutil.PipelineManifestReader
	if pipelinePath != "" {
		r, err := fileIndicatorToReadCloser(pipelinePath)
		if err != nil {
			return err
		}
		defer r.Close()
		pipelineReader, err = ppsutil.NewPipelineManifestReader(r)
		if err != nil {
			return err
		}

	} else if jsonnetPath != "" {
		pipelineBytes, err := evaluateJsonnetTemplate(pc, jsonnetPath, jsonnetArgs)
		if err != nil {
			return err
		}
		pipelineReader, err = ppsutil.NewPipelineManifestReader(bytes.NewReader(pipelineBytes))
		if err != nil {
			return err
		}
	}
	for {
		request, err := pipelineReader.NextCreatePipelineRequest()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}

		// Add trace if env var is set
		ctx, err := extended.EmbedAnyDuration(pc.Ctx())
		pc = pc.WithCtx(ctx)
		if err != nil {
			log.Error(ctx, "problem adding trace data", zap.Error(err))
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

		if request.Pipeline.Project.GetName() == "" {
			request.Pipeline.Project = &pfs.Project{Name: project}
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

func dockerPushHelper(request *pps.CreatePipelineRequest, registry, username string) error {
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
type ByCreationTime []*pps.JobInfo

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
		if state, err := pps.JobStateFromName(stateStr); err == nil {
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
		if state, err := pps.PipelineStateFromName(stateStr); err == nil {
			conditions = append(conditions, fmt.Sprintf(".state == \"%s\"", state))
		} else {
			return "", err
		}
	}
	return validateJQConditionString(strings.Join(conditions, " or "))
}

// Copied from src/client/pps.go's APIClient, because we need to add the new projects filter to the request,
// rather then adding yet another param, we are just going to pass in the request.
func listJobFilterF(ctx context.Context, client pps.API_ListJobClient, f func(*pps.JobInfo) error) error {
	for {
		ji, err := client.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		} else if err != nil {
			return grpcutil.ScrubGRPC(err)
		}
		if err := f(ji); err != nil {
			if errors.Is(err, errutil.ErrBreak) {
				return nil
			}
			return err
		}
	}
}
