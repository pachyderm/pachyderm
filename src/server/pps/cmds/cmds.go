package cmds

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"strings"

	units "github.com/docker/go-units"
	"github.com/fsouza/go-dockerclient"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	pachdclient "github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/tabwriter"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/pps/pretty"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

const (
	codestart  = "```sh"
	codeend    = "```"
	termHeight = 24
)

// Cmds returns a slice containing pps commands.
func Cmds(noMetrics *bool) ([]*cobra.Command, error) {
	metrics := !*noMetrics
	raw := false
	rawFlag := func(cmd *cobra.Command) {
		cmd.Flags().BoolVar(&raw, "raw", false, "disable pretty printing, print raw json")
	}
	marshaller := &jsonpb.Marshaler{
		Indent:   "  ",
		OrigName: true,
	}

	job := &cobra.Command{
		Use:   "job",
		Short: "Docs for jobs.",
		Long: `Jobs are the basic unit of computation in Pachyderm.

Jobs run a containerized workload over a set of finished input commits.
Creating a job will also create a new repo and a commit in that repo which
contains the output of the job. Unless the job is created with another job as a
parent. If the job is created with a parent it will use the same repo as its
parent job and the commit it creates will use the parent job's commit as a
parent.
If the job fails the commit it creates will not be finished.
The increase the throughput of a job increase the Shard paremeter.
`,
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			return nil
		}),
	}

	pipelineSpec := "[Pipeline Specification](../reference/pipeline_spec.html)"

	var block bool
	inspectJob := &cobra.Command{
		Use:   "inspect-job job-id",
		Short: "Return info about a job.",
		Long:  "Return info about a job.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := pachdclient.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
			jobInfo, err := client.InspectJob(args[0], block)
			if err != nil {
				cmdutil.ErrorAndExit("error from InspectJob: %s", err.Error())
			}
			if jobInfo == nil {
				cmdutil.ErrorAndExit("job %s not found.", args[0])
			}
			if raw {
				return marshaller.Marshal(os.Stdout, jobInfo)
			}
			return pretty.PrintDetailedJobInfo(jobInfo)
		}),
	}
	inspectJob.Flags().BoolVarP(&block, "block", "b", false, "block until the job has either succeeded or failed")
	rawFlag(inspectJob)

	var pipelineName string
	var outputCommitStr string
	var inputCommitStrs []string
	listJob := &cobra.Command{
		Use:   "list-job [commits]",
		Short: "Return info about jobs.",
		Long: `Return info about jobs.

Examples:

` + codestart + `# return all jobs
$ pachctl list-job

# return all jobs in pipeline foo
$ pachctl list-job -p foo

# return all jobs whose input commits include foo/XXX and bar/YYY
$ pachctl list-job foo/XXX bar/YYY

# return all jobs in pipeline foo and whose input commits include bar/YYY
$ pachctl list-job -p foo bar/YYY
` + codeend,
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			client, err := pachdclient.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}

			commits, err := cmdutil.ParseCommits(inputCommitStrs)
			if err != nil {
				return err
			}

			var outputCommit *pfs.Commit
			if outputCommitStr != "" {
				outputCommits, err := cmdutil.ParseCommits([]string{outputCommitStr})
				if err != nil {
					return err
				}
				if len(outputCommits) == 1 {
					outputCommit = outputCommits[0]
				}
			}

			if raw {
				if err := client.ListJobF(pipelineName, commits, outputCommit, func(ji *ppsclient.JobInfo) error {
					if err := marshaller.Marshal(os.Stdout, ji); err != nil {
						return err
					}
					return nil
					return nil
				}); err != nil {
					return err
				}
				return nil
			}
			writer := tabwriter.NewWriter(os.Stdout, pretty.JobHeader)
			if err := client.ListJobF(pipelineName, commits, outputCommit, func(ji *ppsclient.JobInfo) error {
				pretty.PrintJobInfo(writer, ji)
				return nil
			}); err != nil {
				return err
			}
			return writer.Flush()
		}),
	}
	listJob.Flags().StringVarP(&pipelineName, "pipeline", "p", "", "Limit to jobs made by pipeline.")
	listJob.Flags().StringVarP(&outputCommitStr, "output", "o", "", "List jobs with a specific output commit.")
	listJob.Flags().StringSliceVarP(&inputCommitStrs, "input", "i", []string{}, "List jobs with a specific set of input commits.")
	rawFlag(listJob)

	var pipelines cmdutil.RepeatedStringArg
	flushJob := &cobra.Command{
		Use:   "flush-job commit [commit ...]",
		Short: "Wait for all jobs caused by the specified commits to finish and return them.",
		Long: `Wait for all jobs caused by the specified commits to finish and return them.

Examples:

` + codestart + `# return jobs caused by foo/XXX and bar/YYY
$ pachctl flush-job foo/XXX bar/YYY

# return jobs caused by foo/XXX leading to pipelines bar and baz
$ pachctl flush-job foo/XXX -p bar -p baz
` + codeend,
		Run: cmdutil.Run(func(args []string) error {
			commits, err := cmdutil.ParseCommits(args)
			if err != nil {
				return err
			}

			c, err := pachdclient.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}

			jobInfos, err := c.FlushJobAll(commits, pipelines)
			if err != nil {
				return err
			}

			if raw {
				for _, jobInfo := range jobInfos {
					if err := marshaller.Marshal(os.Stdout, jobInfo); err != nil {
						return err
					}
				}
				return nil
			}
			writer := tabwriter.NewWriter(os.Stdout, pretty.JobHeader)
			for _, jobInfo := range jobInfos {
				pretty.PrintJobInfo(writer, jobInfo)
			}

			return writer.Flush()
		}),
	}
	flushJob.Flags().VarP(&pipelines, "pipeline", "p", "Wait only for jobs leading to a specific set of pipelines")
	rawFlag(flushJob)

	deleteJob := &cobra.Command{
		Use:   "delete-job job-id",
		Short: "Delete a job.",
		Long:  "Delete a job.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := pachdclient.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
			if err := client.DeleteJob(args[0]); err != nil {
				cmdutil.ErrorAndExit("error from DeleteJob: %s", err.Error())
			}
			return nil
		}),
	}

	stopJob := &cobra.Command{
		Use:   "stop-job job-id",
		Short: "Stop a job.",
		Long:  "Stop a job.  The job will be stopped immediately.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := pachdclient.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
			if err := client.StopJob(args[0]); err != nil {
				cmdutil.ErrorAndExit("error from StopJob: %s", err.Error())
			}
			return nil
		}),
	}

	restartDatum := &cobra.Command{
		Use:   "restart-datum job-id datum-path1,datum-path2",
		Short: "Restart a datum.",
		Long:  "Restart a datum.",
		Run: cmdutil.RunFixedArgs(2, func(args []string) error {
			client, err := pachdclient.NewOnUserMachine(metrics, "user")
			if err != nil {
				return fmt.Errorf("error connecting to pachd: %v", err)
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
			return client.RestartDatum(args[0], datumFilter)
		}),
	}
	var pageSize int64
	var page int64
	listDatum := &cobra.Command{
		Use:   "list-datum job-id",
		Short: "Return the datums in a job.",
		Long:  "Return the datums in a job.",
		Run: cmdutil.RunBoundedArgs(1, 1, func(args []string) error {
			client, err := pachdclient.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
			if pageSize < 0 {
				return fmt.Errorf("pageSize must be zero or positive")
			}
			if page < 0 {
				return fmt.Errorf("page must be zero or positive")
			}
			if raw {
				if err := client.ListDatumF(args[0], pageSize, page, func(di *ppsclient.DatumInfo) error {
					return marshaller.Marshal(os.Stdout, di)
				}); err != nil {
					return err
				}
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
	rawFlag(listDatum)
	listDatum.Flags().Int64Var(&pageSize, "pageSize", 0, "Specify the number of results sent back in a single page")
	listDatum.Flags().Int64Var(&page, "page", 0, "Specify the page of results to send")

	inspectDatum := &cobra.Command{
		Use:   "inspect-datum job-id datum-id",
		Short: "Display detailed info about a single datum.",
		Long:  "Display detailed info about a single datum.",
		Run: cmdutil.RunBoundedArgs(2, 2, func(args []string) error {
			client, err := pachdclient.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
			datumInfo, err := client.InspectDatum(args[0], args[1])
			if err != nil {
				return err
			}
			if raw {
				return marshaller.Marshal(os.Stdout, datumInfo)
			}
			pretty.PrintDetailedDatumInfo(os.Stdout, datumInfo)
			return nil
		}),
	}
	rawFlag(inspectDatum)

	var (
		jobID       string
		datumID     string
		commaInputs string // comma-separated list of input files of interest
		master      bool
		follow      bool
		tail        int64
	)
	getLogs := &cobra.Command{
		Use:   "get-logs [--pipeline=<pipeline>|--job=<job id>] [--datum=<datum id>]",
		Short: "Return logs from a job.",
		Long: `Return logs from a job.

Examples:

` + codestart + `# return logs emitted by recent jobs in the "filter" pipeline
$ pachctl get-logs --pipeline=filter

# return logs emitted by the job aedfa12aedf
$ pachctl get-logs --job=aedfa12aedf

# return logs emitted by the pipeline \"filter\" while processing /apple.txt and a file with the hash 123aef
$ pachctl get-logs --pipeline=filter --inputs=/apple.txt,123aef
` + codeend,
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			client, err := pachdclient.NewOnUserMachine(metrics, "user")
			if err != nil {
				return fmt.Errorf("error connecting to pachd: %v", err)
			}

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
			marshaler := &jsonpb.Marshaler{}
			iter := client.GetLogs(pipelineName, jobID, data, datumID, master, follow, tail)
			for iter.Next() {
				var messageStr string
				if raw {
					var err error
					messageStr, err = marshaler.MarshalToString(iter.Message())
					if err != nil {
						fmt.Fprintf(os.Stderr, "error marshalling \"%v\": %s\n", iter.Message(), err)
					}
					fmt.Println(messageStr)
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
	getLogs.Flags().StringVar(&jobID, "job", "", "Filter for log lines from "+
		"this job (accepts job ID)")
	getLogs.Flags().StringVar(&datumID, "datum", "", "Filter for log lines for this datum (accepts datum ID)")
	getLogs.Flags().StringVar(&commaInputs, "inputs", "", "Filter for log lines "+
		"generated while processing these files (accepts PFS paths or file hashes)")
	getLogs.Flags().BoolVar(&master, "master", false, "Return log messages from the master process (pipeline must be set).")
	getLogs.Flags().BoolVar(&raw, "raw", false, "Return log messages verbatim from server.")
	getLogs.Flags().BoolVarP(&follow, "follow", "f", false, "Follow logs as more are created.")
	getLogs.Flags().Int64VarP(&tail, "tail", "t", 0, "Lines of recent logs to display.")

	pipeline := &cobra.Command{
		Use:   "pipeline",
		Short: "Docs for pipelines.",
		Long: `Pipelines are a powerful abstraction for automating jobs.

Pipelines take a set of repos as inputs, rather than the set of commits that
jobs take. Pipelines then subscribe to commits on those repos and launches a job
to process each incoming commit.
Creating a pipeline will also create a repo of the same name.
All jobs created by a pipeline will create commits in the pipeline's repo.
`,
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			return nil
		}),
	}

	var pushImages bool
	var registry string
	var username string
	var password string
	var pipelinePath string
	createPipeline := &cobra.Command{
		Use:   "create-pipeline -f pipeline.json",
		Short: "Create a new pipeline.",
		Long:  fmt.Sprintf("Create a new pipeline from a %s", pipelineSpec),
		Run: cmdutil.RunFixedArgs(0, func(args []string) (retErr error) {
			cfgReader, err := ppsutil.NewPipelineManifestReader(pipelinePath)
			if err != nil {
				return err
			}
			client, err := pachdclient.NewOnUserMachine(metrics, "user")
			if err != nil {
				return fmt.Errorf("error connecting to pachd: %v", err)
			}
			for {
				request, err := cfgReader.NextCreatePipelineRequest()
				if err == io.EOF {
					break
				} else if err != nil {
					return err
				}
				if request.Input.Atom != nil {
					fmt.Println("WARNING: The `atom` input type has been deprecated and will be removed in a future version. Please replace `atom` with `pfs`.")
				}
				if pushImages {
					pushedImage, err := pushImage(registry, username, password, request.Transform.Image)
					if err != nil {
						return err
					}
					request.Transform.Image = pushedImage
				}
				if _, err := client.PpsAPIClient.CreatePipeline(
					client.Ctx(),
					request,
				); err != nil {
					return grpcutil.ScrubGRPC(err)
				}
			}
			return nil
		}),
	}
	createPipeline.Flags().StringVarP(&pipelinePath, "file", "f", "-", "The file containing the pipeline, it can be a url or local file. - reads from stdin.")
	createPipeline.Flags().BoolVarP(&pushImages, "push-images", "p", false, "If true, push local docker images into the cluster registry.")
	createPipeline.Flags().StringVarP(&registry, "registry", "r", "docker.io", "The registry to push images to.")
	createPipeline.Flags().StringVarP(&username, "username", "u", "", "The username to push images as, defaults to your OS username.")
	createPipeline.Flags().StringVarP(&password, "password", "", "", "Your password for the registry being pushed to.")

	var reprocess bool
	updatePipeline := &cobra.Command{
		Use:   "update-pipeline -f pipeline.json",
		Short: "Update an existing Pachyderm pipeline.",
		Long:  fmt.Sprintf("Update a Pachyderm pipeline with a new %s", pipelineSpec),
		Run: cmdutil.RunFixedArgs(0, func(args []string) (retErr error) {
			cfgReader, err := ppsutil.NewPipelineManifestReader(pipelinePath)
			if err != nil {
				return err
			}
			client, err := pachdclient.NewOnUserMachine(metrics, "user")
			if err != nil {
				return fmt.Errorf("error connecting to pachd: %v", err)
			}
			for {
				request, err := cfgReader.NextCreatePipelineRequest()
				if err == io.EOF {
					break
				} else if err != nil {
					return err
				}
				request.Update = true
				request.Reprocess = reprocess
				if request.Input.Atom != nil {
					fmt.Println("WARNING: The `atom` input type has been deprecated and will be removed in a future version. Please replace `atom` with `pfs`.")
				}
				if pushImages {
					pushedImage, err := pushImage(registry, username, password, request.Transform.Image)
					if err != nil {
						return err
					}
					request.Transform.Image = pushedImage
				}
				if _, err := client.PpsAPIClient.CreatePipeline(
					client.Ctx(),
					request,
				); err != nil {
					return grpcutil.ScrubGRPC(err)
				}
			}
			return nil
		}),
	}
	updatePipeline.Flags().StringVarP(&pipelinePath, "file", "f", "-", "The file containing the pipeline, it can be a url or local file. - reads from stdin.")
	updatePipeline.Flags().BoolVarP(&pushImages, "push-images", "p", false, "If true, push local docker images into the cluster registry.")
	updatePipeline.Flags().StringVarP(&registry, "registry", "r", "docker.io", "The registry to push images to.")
	updatePipeline.Flags().StringVarP(&username, "username", "u", "", "The username to push images as, defaults to your OS username.")
	updatePipeline.Flags().StringVarP(&password, "password", "", "", "Your password for the registry being pushed to.")
	updatePipeline.Flags().BoolVar(&reprocess, "reprocess", false, "If true, reprocess datums that were already processed by previous version of the pipeline.")

	inspectPipeline := &cobra.Command{
		Use:   "inspect-pipeline pipeline-name",
		Short: "Return info about a pipeline.",
		Long:  "Return info about a pipeline.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := pachdclient.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
			pipelineInfo, err := client.InspectPipeline(args[0])
			if err != nil {
				return err
			}
			if pipelineInfo == nil {
				return fmt.Errorf("pipeline %s not found", args[0])
			}
			if raw {
				return marshaller.Marshal(os.Stdout, pipelineInfo)
			}
			return pretty.PrintDetailedPipelineInfo(pipelineInfo)
		}),
	}
	rawFlag(inspectPipeline)

	extractPipeline := &cobra.Command{
		Use:   "extract-pipeline pipeline-name",
		Short: "Return the manifest used to create a pipeline.",
		Long:  "Return the manifest used to create a pipeline.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := pachdclient.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
			createPipelineRequest, err := client.ExtractPipeline(args[0])
			if err != nil {
				return err
			}
			return marshaller.Marshal(os.Stdout, createPipelineRequest)
		}),
	}

	var editor string
	editPipeline := &cobra.Command{
		Use:   "edit-pipeline pipeline-name",
		Short: "Edit the manifest for a pipeline in your text editor.",
		Long:  "Edit the manifest for a pipeline in your text editor.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) (retErr error) {
			client, err := pachdclient.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
			createPipelineRequest, err := client.ExtractPipeline(args[0])
			if err != nil {
				return err
			}
			f, err := ioutil.TempFile("", args[0])
			if err != nil {
				return err
			}
			if err := marshaller.Marshal(f, createPipelineRequest); err != nil {
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
			cfgReader, err := ppsutil.NewPipelineManifestReader(f.Name())
			if err != nil {
				return err
			}
			request, err := cfgReader.NextCreatePipelineRequest()
			if err != nil {
				return err
			}
			if proto.Equal(createPipelineRequest, request) {
				fmt.Println("Pipeline unchanged, no update will be performed.")
				return nil
			}
			request.Update = true
			request.Reprocess = reprocess
			if request.Input.Atom != nil {
				fmt.Println("WARNING: The `atom` input type has been deprecated and will be removed in a future version. Please replace `atom` with `pfs`.")
			}
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

	var spec bool
	listPipeline := &cobra.Command{
		Use:   "list-pipeline",
		Short: "Return info about all pipelines.",
		Long:  "Return info about all pipelines.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			client, err := pachdclient.NewOnUserMachine(metrics, "user")
			if err != nil {
				return fmt.Errorf("error connecting to pachd: %v", err)
			}
			pipelineInfos, err := client.ListPipeline()
			if err != nil {
				return err
			}
			if raw {
				for _, pipelineInfo := range pipelineInfos {
					if err := marshaller.Marshal(os.Stdout, pipelineInfo); err != nil {
						return err
					}
				}
				return nil
			}
			if spec {
				for _, pipelineInfo := range pipelineInfos {
					if err := marshaller.Marshal(os.Stdout, ppsutil.PipelineReqFromInfo(pipelineInfo)); err != nil {
						return err
					}
				}
				return nil
			}
			writer := tabwriter.NewWriter(os.Stdout, pretty.PipelineHeader)
			for _, pipelineInfo := range pipelineInfos {
				pretty.PrintPipelineInfo(writer, pipelineInfo)
			}
			return writer.Flush()
		}),
	}
	rawFlag(listPipeline)
	listPipeline.Flags().BoolVarP(&spec, "spec", "s", false, "Output create-pipeline compatibility specs.")

	var all bool
	var force bool
	deletePipeline := &cobra.Command{
		Use:   "delete-pipeline pipeline-name",
		Short: "Delete a pipeline.",
		Long:  "Delete a pipeline.",
		Run: cmdutil.RunBoundedArgs(0, 1, func(args []string) error {
			client, err := pachdclient.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
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

	startPipeline := &cobra.Command{
		Use:   "start-pipeline pipeline-name",
		Short: "Restart a stopped pipeline.",
		Long:  "Restart a stopped pipeline.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := pachdclient.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
			if err := client.StartPipeline(args[0]); err != nil {
				cmdutil.ErrorAndExit("error from StartPipeline: %s", err.Error())
			}
			return nil
		}),
	}

	stopPipeline := &cobra.Command{
		Use:   "stop-pipeline pipeline-name",
		Short: "Stop a running pipeline.",
		Long:  "Stop a running pipeline.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := pachdclient.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
			if err := client.StopPipeline(args[0]); err != nil {
				cmdutil.ErrorAndExit("error from StopPipeline: %s", err.Error())
			}
			return nil
		}),
	}

	var memory string
	garbageCollect := &cobra.Command{
		Use:   "garbage-collect",
		Short: "Garbage collect unused data.",
		Long: `Garbage collect unused data.

When a file/commit/repo is deleted, the data is not immediately removed from
the underlying storage system (e.g. S3) for performance and architectural
reasons.  This is similar to how when you delete a file on your computer, the
file is not necessarily wiped from disk immediately.

To actually remove the data, you will need to manually invoke garbage
collection with "pachctl garbage-collect".

Currently "pachctl garbage-collect" can only be started when there are no
pipelines running.  You also need to ensure that there's no ongoing "put-file".
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
			client, err := pachdclient.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
			memoryBytes, err := units.RAMInBytes(memory)
			if err != nil {
				return err
			}
			return client.GarbageCollect(memoryBytes)
		}),
	}
	garbageCollect.Flags().StringVarP(&memory, "memory", "m", "0", "The amount of memory to use during garbage collection. Default is 10MB.")

	var result []*cobra.Command
	result = append(result, job)
	result = append(result, inspectJob)
	result = append(result, listJob)
	result = append(result, flushJob)
	result = append(result, deleteJob)
	result = append(result, stopJob)
	result = append(result, restartDatum)
	result = append(result, listDatum)
	result = append(result, inspectDatum)
	result = append(result, getLogs)
	result = append(result, pipeline)
	result = append(result, createPipeline)
	result = append(result, updatePipeline)
	result = append(result, inspectPipeline)
	result = append(result, extractPipeline)
	result = append(result, editPipeline)
	result = append(result, listPipeline)
	result = append(result, deletePipeline)
	result = append(result, startPipeline)
	result = append(result, stopPipeline)
	result = append(result, garbageCollect)
	return result, nil
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

// pushImage pushes an image as registry/user/image. Registry and user can be
// left empty.
func pushImage(registry string, username string, password string, image string) (string, error) {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return "", err
	}
	repo, _ := docker.ParseRepositoryTag(image)
	components := strings.Split(repo, "/")
	name := components[len(components)-1]
	if username == "" {
		user, err := user.Current()
		if err != nil {
			return "", err
		}
		username = user.Username
	}
	pushRepo := fmt.Sprintf("%s/%s/%s", registry, username, name)
	pushTag := uuid.NewWithoutDashes()
	if err := client.TagImage(image, docker.TagImageOptions{
		Repo:    pushRepo,
		Tag:     pushTag,
		Context: context.Background(),
	}); err != nil {
		return "", err
	}
	var authConfig docker.AuthConfiguration
	if password != "" {
		authConfig = docker.AuthConfiguration{ServerAddress: registry}
		authConfig.Username = username
		authConfig.Password = password
	} else {
		authConfigs, err := docker.NewAuthConfigurationsFromDockerCfg()
		if err != nil {
			return "", fmt.Errorf("error parsing auth: %s, try running `docker login`", err.Error())
		}
		for _, _authConfig := range authConfigs.Configs {
			serverAddress := _authConfig.ServerAddress
			if strings.Contains(serverAddress, registry) {
				authConfig = _authConfig
				break
			}
		}
	}
	fmt.Printf("Pushing %s:%s, this may take a while.\n", pushRepo, pushTag)
	if err := client.PushImage(
		docker.PushImageOptions{
			Name: pushRepo,
			Tag:  pushTag,
		},
		authConfig,
	); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s", pushRepo, pushTag), nil
}
