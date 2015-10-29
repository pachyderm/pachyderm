package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/pretty"

	"go.pedge.io/env"
	"go.pedge.io/pkg/cobra"
	"go.pedge.io/proto/stream"

	"google.golang.org/grpc"

	"github.com/spf13/cobra"
)

var (
	defaultEnv = map[string]string{
		"PPS_ADDRESS": "0.0.0.0:651",
	}
)

type appEnv struct {
	PachydermPpsd1Port string `env:"PACHYDERM_PPSD_1_PORT"`
	Address            string `env:"PPS_ADDRESS"`
}

func main() {
	env.Main(do, &appEnv{}, defaultEnv)
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	address := appEnv.PachydermPpsd1Port
	if address == "" {
		address = appEnv.Address
	} else {
		address = strings.Replace(address, "tcp://", "", -1)
	}
	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return err
	}
	jobAPIClient := pps.NewJobAPIClient(clientConn)
	pipelineAPIClient := pps.NewPipelineAPIClient(clientConn)
	rootCmd := &cobra.Command{
		Use: "pps",
		Long: `Access the PPS API.

Note that this CLI is experimental and does not even check for common errors.
The environment variable PPS_ADDRESS controls what server the CLI connects to, the default is 0.0.0.0:651.`,
	}

	var image string
	var outParentCommitId string
	createJob := &cobra.Command{
		Use:   "create-job in-repo-name in-commit-id out-repo-name command [args]",
		Short: "Create a new job. Returns the id of the created job.",
		Long: `Create a new job. With repo-name/commit-id as input and
out-repo-name as output. A commit will be created for the output.
You can find out the name of the commit with inspect-job.`,
		Run: func(cmd *cobra.Command, args []string) {
			job, err := jobAPIClient.CreateJob(
				context.Background(),
				&pps.CreateJobRequest{
					Spec: &pps.CreateJobRequest_Transform{
						Transform: &pps.Transform{
							Image: image,
							Cmd:   args[3:],
						},
					},
					Input: &pfs.Commit{
						Repo: &pfs.Repo{
							Name: args[0],
						},
						Id: args[1],
					},
					OutputParent: &pfs.Commit{
						Repo: &pfs.Repo{
							Name: args[2],
						},
						Id: outParentCommitId,
					},
				})
			if err != nil {
				errorAndExit("Error from CreateJob: %s", err.Error())
			}
			fmt.Println(job.Id)
		},
	}
	createJob.Flags().StringVarP(&image, "image", "i", "ubuntu", "The image to run the job in.")
	createJob.Flags().StringVarP(&outParentCommitId, "parent", "p", "", "The parent to use for the output commit.")

	inspectJob := &cobra.Command{
		Use:   "inspect-job job-id",
		Short: "Return info about a job.",
		Long:  "Return info about a job.",
		Run: pkgcobra.RunFixedArgs(1, func(args []string) error {
			jobInfo, err := jobAPIClient.InspectJob(
				context.Background(),
				&pps.InspectJobRequest{
					Job: &pps.Job{
						Id: args[0],
					},
				},
			)
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
			var pipeline *pps.Pipeline
			if pipelineName != "" {
				pipeline = &pps.Pipeline{
					Name: pipelineName,
				}
			}
			jobInfos, err := jobAPIClient.ListJob(
				context.Background(),
				&pps.ListJobRequest{
					Pipeline: pipeline,
				},
			)
			if err != nil {
				errorAndExit("Error from InspectJob: %s", err.Error())
			}
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			pretty.PrintJobHeader(writer)
			for _, jobInfo := range jobInfos.JobInfo {
				pretty.PrintJobInfo(writer, jobInfo)
			}
			return writer.Flush()
		}),
	}
	listJob.Flags().StringVarP(&pipelineName, "pipeline", "p", "", "Limit to jobs made by pipeline.")

	getJobLogs := &cobra.Command{
		Use:   "logs job-id",
		Short: "Return logs from a job.",
		Long:  "Return logs from a job.",
		Run: pkgcobra.RunFixedArgs(1, func(args []string) error {
			logsClient, err := jobAPIClient.GetJobLogs(
				context.Background(),
				&pps.GetJobLogsRequest{
					Job: &pps.Job{
						Id: args[0],
					},
					OutputStream: pps.OutputStream_OUTPUT_STREAM_ALL,
				},
			)
			if err != nil {
				errorAndExit("Error from InspectJob: %s", err.Error())
			}
			if err := protostream.WriteFromStreamingBytesClient(logsClient, os.Stdout); err != nil {
				return err
			}
			return nil
		}),
	}

	createPipeline := &cobra.Command{
		Use:   "create-pipeline pipeline-name input-repo output-repo command [args]",
		Short: "Create a new pipeline.",
		Long:  "Create a new pipeline.",
		Run: func(cmd *cobra.Command, args []string) {
			if _, err := pipelineAPIClient.CreatePipeline(
				context.Background(),
				&pps.CreatePipelineRequest{
					Pipeline: &pps.Pipeline{
						Name: args[0],
					},
					Transform: &pps.Transform{
						Image: image,
						Cmd:   args[3:],
					},
					Input: &pfs.Repo{
						Name: args[1],
					},
					Output: &pfs.Repo{
						Name: args[2],
					},
				},
			); err != nil {
				errorAndExit("Error from CreatePipeline: %s", err.Error())
			}
		},
	}
	createPipeline.Flags().StringVarP(&image, "image", "i", "ubuntu", "The image to run the pipeline's jobs in.")

	inspectPipeline := &cobra.Command{
		Use:   "inspect-pipeline pipeline-name",
		Short: "Return info about a pipeline.",
		Long:  "Return info about a pipeline.",
		Run: pkgcobra.RunFixedArgs(1, func(args []string) error {
			pipelineInfo, err := pipelineAPIClient.InspectPipeline(
				context.Background(),
				&pps.InspectPipelineRequest{
					Pipeline: &pps.Pipeline{
						Name: args[0],
					},
				},
			)
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
			pipelineInfos, err := pipelineAPIClient.ListPipeline(
				context.Background(),
				&pps.ListPipelineRequest{},
			)
			if err != nil {
				errorAndExit("Error from ListPipeline: %s", err.Error())
			}
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			pretty.PrintPipelineHeader(writer)
			for _, pipelineInfo := range pipelineInfos.PipelineInfo {
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
			if _, err := pipelineAPIClient.DeletePipeline(
				context.Background(),
				&pps.DeletePipelineRequest{
					Pipeline: &pps.Pipeline{
						Name: args[0],
					},
				},
			); err != nil {
				errorAndExit("Error from DeletePipeline: %s", err.Error())
			}
			return nil
		}),
	}

	rootCmd.AddCommand(createJob)
	rootCmd.AddCommand(inspectJob)
	rootCmd.AddCommand(listJob)
	rootCmd.AddCommand(getJobLogs)
	rootCmd.AddCommand(createPipeline)
	rootCmd.AddCommand(inspectPipeline)
	rootCmd.AddCommand(listPipeline)
	rootCmd.AddCommand(deletePipeline)
	return rootCmd.Execute()
}

func errorAndExit(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s\n", fmt.Sprintf(format, args...))
	os.Exit(1)
}
