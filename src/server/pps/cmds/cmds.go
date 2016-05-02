package cmds

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/golang/protobuf/jsonpb"
	"github.com/pachyderm/pachyderm/src/client"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pps/example"
	"github.com/pachyderm/pachyderm/src/server/pps/pretty"
	"github.com/spf13/cobra"
	"go.pedge.io/pkg/cobra"
	"golang.org/x/net/context"
)

func Cmds(address string) ([]*cobra.Command, error) {
	marshaller := &jsonpb.Marshaler{Indent: "  "}

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
		Run: pkgcobra.RunFixedArgs(0, func(args []string) error {
			return nil
		}),
	}

	exampleCreateJobRequest, err := marshaller.MarshalToString(example.CreateJobRequest())
	if err != nil {
		return nil, err
	}

	exampleRunPipelineSpec, err := marshaller.MarshalToString(example.RunPipelineSpec())
	if err != nil {
		return nil, err
	}

	var jobPath string
	createJob := &cobra.Command{
		Use:   "create-job -f job.json",
		Short: "Create a new job. Returns the id of the created job.",
		Long:  fmt.Sprintf("Create a new job from a spec, the spec looks like this\n%s", exampleCreateJobRequest),
		Run: func(cmd *cobra.Command, args []string) {
			client, err := client.NewFromAddress(address)
			if err != nil {
				errorAndExit("Error connecting to pps: %s", err.Error())
			}
			var jobReader io.Reader
			if jobPath == "-" {
				jobReader = os.Stdin
				fmt.Print("Reading from stdin.\n")
			} else {
				jobFile, err := os.Open(jobPath)
				if err != nil {
					errorAndExit("Error opening %s: %s", jobPath, err.Error())
				}
				defer func() {
					if err := jobFile.Close(); err != nil {
						errorAndExit("Error closing%s: %s", jobPath, err.Error())
					}
				}()
				jobReader = jobFile
			}
			var request ppsclient.CreateJobRequest
			if err := jsonpb.Unmarshal(jobReader, &request); err != nil {
				errorAndExit("Error reading from stdin: %s", err.Error())
			}
			job, err := client.PpsAPIClient.CreateJob(
				context.Background(),
				&request,
			)
			if err != nil {
				errorAndExit("Error from CreateJob: %s", err.Error())
			}
			fmt.Println(job.ID)
		},
	}
	createJob.Flags().StringVarP(&jobPath, "file", "f", "-", "The file containing the job, - reads from stdin.")

	inspectJob := &cobra.Command{
		Use:   "inspect-job job-id",
		Short: "Return info about a job.",
		Long:  "Return info about a job.",
		Run: pkgcobra.RunFixedArgs(1, func(args []string) error {
			client, err := client.NewFromAddress(address)
			if err != nil {
				return err
			}
			jobInfo, err := client.InspectJob(args[0], false, false)
			if err != nil {
				errorAndExit("Error from InspectJob: %s", err.Error())
			}
			if jobInfo == nil {
				errorAndExit("Job %s not found.", args[0])
			}
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			pretty.PrintJobHeader(writer)
			pretty.PrintJobInfo(writer, jobInfo)
			return writer.Flush()
		}),
	}

	var pipelineName string
	listJob := &cobra.Command{
		Use:   "list-job -p pipeline-name",
		Short: "Return info about all jobs.",
		Long:  "Return info about all jobs.",
		Run: pkgcobra.RunFixedArgs(0, func(args []string) error {
			client, err := client.NewFromAddress(address)
			if err != nil {
				return err
			}
			jobInfos, err := client.ListJob(pipelineName, nil)
			if err != nil {
				errorAndExit("Error from InspectJob: %s", err.Error())
			}
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			pretty.PrintJobHeader(writer)
			for _, jobInfo := range jobInfos {
				pretty.PrintJobInfo(writer, jobInfo)
			}
			return writer.Flush()
		}),
	}
	listJob.Flags().StringVarP(&pipelineName, "pipeline", "p", "", "Limit to jobs made by pipeline.")

	getLogs := &cobra.Command{
		Use:   "get-logs job-id",
		Short: "Return logs from a job.",
		Long:  "Return logs from a job.",
		Run: pkgcobra.RunFixedArgs(1, func(args []string) error {
			client, err := client.NewFromAddress(address)
			if err != nil {
				return err
			}
			return client.GetLogs(args[0], os.Stdout)
		}),
	}

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
		Run: pkgcobra.RunFixedArgs(0, func(args []string) error {
			return nil
		}),
	}

	var pipelinePath string
	exampleCreatePipelineRequest, err := marshaller.MarshalToString(example.CreatePipelineRequest())
	if err != nil {
		return nil, err
	}
	createPipeline := &cobra.Command{
		Use:   "create-pipeline -f pipeline.json",
		Short: "Create a new pipeline.",
		Long:  fmt.Sprintf("Create a new pipeline from a spec, the spec looks like this\n%s", exampleCreatePipelineRequest),
		Run: func(cmd *cobra.Command, args []string) {
			client, err := client.NewFromAddress(address)
			if err != nil {
				errorAndExit("Error connecting to pps: %s", err.Error())
			}
			var pipelineReader io.Reader
			if pipelinePath == "-" {
				pipelineReader = os.Stdin
				fmt.Print("Reading from stdin.\n")
			} else {
				pipelineFile, err := os.Open(pipelinePath)
				if err != nil {
					errorAndExit("Error opening %s: %s", pipelinePath, err.Error())
				}
				defer func() {
					if err := pipelineFile.Close(); err != nil {
						errorAndExit("Error closing%s: %s", pipelinePath, err.Error())
					}
				}()
				pipelineReader = pipelineFile
			}
			var request ppsclient.CreatePipelineRequest
			decoder := json.NewDecoder(pipelineReader)
			for {
				message := json.RawMessage{}
				if err := decoder.Decode(&message); err != nil {
					if err == io.EOF {
						break
					} else {
						errorAndExit("Error reading from stdin: %s", err.Error())
					}
				}
				if err := jsonpb.UnmarshalString(string(message), &request); err != nil {
					errorAndExit("Error reading from stdin: %s", err.Error())
				}
				if _, err := client.PpsAPIClient.CreatePipeline(
					context.Background(),
					&request,
				); err != nil {
					errorAndExit("Error from CreatePipeline: %s", err.Error())
				}
			}
		},
	}
	createPipeline.Flags().StringVarP(&pipelinePath, "file", "f", "-", "The file containing the pipeline, - reads from stdin.")

	inspectPipeline := &cobra.Command{
		Use:   "inspect-pipeline pipeline-name",
		Short: "Return info about a pipeline.",
		Long:  "Return info about a pipeline.",
		Run: pkgcobra.RunFixedArgs(1, func(args []string) error {
			client, err := client.NewFromAddress(address)
			if err != nil {
				return err
			}
			pipelineInfo, err := client.InspectPipeline(args[0])
			if err != nil {
				errorAndExit("Error from InspectPipeline: %s", err.Error())
			}
			if pipelineInfo == nil {
				errorAndExit("Pipeline %s not found.", args[0])
			}
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			pretty.PrintPipelineHeader(writer)
			pretty.PrintPipelineInfo(writer, pipelineInfo)
			return writer.Flush()
		}),
	}

	listPipeline := &cobra.Command{
		Use:   "list-pipeline",
		Short: "Return info about all pipelines.",
		Long:  "Return info about all pipelines.",
		Run: pkgcobra.RunFixedArgs(0, func(args []string) error {
			client, err := client.NewFromAddress(address)
			if err != nil {
				return err
			}
			pipelineInfos, err := client.ListPipeline()
			if err != nil {
				errorAndExit("Error from ListPipeline: %s", err.Error())
			}
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			pretty.PrintPipelineHeader(writer)
			for _, pipelineInfo := range pipelineInfos {
				pretty.PrintPipelineInfo(writer, pipelineInfo)
			}
			return writer.Flush()
		}),
	}

	deletePipeline := &cobra.Command{
		Use:   "delete-pipeline pipeline-name",
		Short: "Delete a pipeline.",
		Long:  "Delete a pipeline.",
		Run: pkgcobra.RunFixedArgs(1, func(args []string) error {
			client, err := client.NewFromAddress(address)
			if err != nil {
				return err
			}
			if err := client.DeletePipeline(args[0]); err != nil {
				errorAndExit("Error from DeletePipeline: %s", err.Error())
			}
			return nil
		}),
	}

	var specPath string
	runPipeline := &cobra.Command{
		Use:   "run-pipeline pipeline-name [-f job.json]",
		Short: "Run a pipeline once.",
		Long:  fmt.Sprintf("Run a pipeline once, optionally overriding some pipeline options by providing a spec.  The spec looks like this:\n%s", exampleRunPipelineSpec),
		Run: pkgcobra.RunFixedArgs(1, func(args []string) error {
			client, err := client.NewFromAddress(address)
			if err != nil {
				return err
			}

			request := &ppsclient.CreateJobRequest{
				Pipeline: &ppsclient.Pipeline{
					Name: args[0],
				},
			}

			var specReader io.Reader
			if specPath == "-" {
				specReader = os.Stdin
				fmt.Print("Reading from stdin.\n")
			} else if specPath != "" {
				specFile, err := os.Open(specPath)
				if err != nil {
					errorAndExit("Error opening %s: %s", specPath, err.Error())
				}

				defer func() {
					if err := specFile.Close(); err != nil {
						errorAndExit("Error closing%s: %s", specPath, err.Error())
					}
				}()

				specReader = specFile
				if err := jsonpb.Unmarshal(specReader, request); err != nil {
					errorAndExit("Error reading from stdin: %s", err.Error())
				}
			}

			job, err := client.PpsAPIClient.CreateJob(
				context.Background(),
				request,
			)
			if err != nil {
				errorAndExit("Error from RunPipeline: %s", err.Error())
			}
			fmt.Println(job.ID)
			return nil
		}),
	}
	runPipeline.Flags().StringVarP(&specPath, "file", "f", "", "The file containing the run-pipeline spec, - reads from stdin.")

	var result []*cobra.Command
	result = append(result, job)
	result = append(result, createJob)
	result = append(result, inspectJob)
	result = append(result, getLogs)
	result = append(result, listJob)
	result = append(result, pipeline)
	result = append(result, createPipeline)
	result = append(result, inspectPipeline)
	result = append(result, listPipeline)
	result = append(result, deletePipeline)
	result = append(result, runPipeline)
	return result, nil
}

func errorAndExit(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s\n", fmt.Sprintf(format, args...))
	os.Exit(1)
}
