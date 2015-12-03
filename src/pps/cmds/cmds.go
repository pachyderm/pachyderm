package cmds

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/pretty"
	"github.com/spf13/cobra"
	"go.pedge.io/pkg/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func Cmds(address string) ([]*cobra.Command, error) {
	var image string
	var outParentCommitID string
	var shards int

	createJob := &cobra.Command{
		Use:   "create-job in-repo-name in-commit-id out-repo-name command [args]",
		Short: "Create a new job. Returns the id of the created job.",
		Long: `Create a new job. With repo-name/commit-id as input and
out-repo-name as output. A commit will be created for the output.
You can find out the name of the commit with inspect-job.`,
		Run: func(cmd *cobra.Command, args []string) {
			apiClient, err := getAPIClient(address)
			if err != nil {
				errorAndExit("Error connecting to pps: %s", err.Error())
			}
			job, err := apiClient.CreateJob(
				context.Background(),
				&pps.CreateJobRequest{
					Spec: &pps.CreateJobRequest_Transform{
						Transform: &pps.Transform{
							Image: image,
							Cmd:   args[3:],
						},
					},
					Shards: uint64(shards),
					InputCommit: []*pfs.Commit{
						{
							Repo: &pfs.Repo{
								Name: args[0],
							},
							Id: args[1],
						},
					},
					OutputParent: &pfs.Commit{
						Repo: &pfs.Repo{
							Name: args[2],
						},
						Id: outParentCommitID,
					},
				})
			if err != nil {
				errorAndExit("Error from CreateJob: %s", err.Error())
			}
			fmt.Println(job.Id)
		},
	}
	createJob.Flags().StringVarP(&image, "image", "i", "", "The image to run the job in.")
	createJob.Flags().StringVarP(&outParentCommitID, "parent", "p", "", "The parent to use for the output commit.")
	createJob.Flags().IntVarP(&shards, "shards", "s", 1, "The sharding factor for the job.")

	inspectJob := &cobra.Command{
		Use:   "inspect-job job-id",
		Short: "Return info about a job.",
		Long:  "Return info about a job.",
		Run: pkgcobra.RunFixedArgs(1, func(args []string) error {
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			jobInfo, err := apiClient.InspectJob(
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
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			var pipeline *pps.Pipeline
			if pipelineName != "" {
				pipeline = &pps.Pipeline{
					Name: pipelineName,
				}
			}
			jobInfos, err := apiClient.ListJob(
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

	createPipeline := &cobra.Command{
		Use:   "create-pipeline pipeline-name input-repo output-repo command [args]",
		Short: "Create a new pipeline.",
		Long:  "Create a new pipeline.",
		Run: func(cmd *cobra.Command, args []string) {
			apiClient, err := getAPIClient(address)
			if err != nil {
				errorAndExit("Error connecting to pps: %s", err.Error())
			}
			if _, err := apiClient.CreatePipeline(
				context.Background(),
				&pps.CreatePipelineRequest{
					Pipeline: &pps.Pipeline{
						Name: args[0],
					},
					Transform: &pps.Transform{
						Image: image,
						Cmd:   args[3:],
					},
					Shards: uint64(shards),
					InputRepo: []*pfs.Repo{
						{
							Name: args[1],
						},
					},
					OutputRepo: &pfs.Repo{
						Name: args[2],
					},
				},
			); err != nil {
				errorAndExit("Error from CreatePipeline: %s", err.Error())
			}
		},
	}
	createPipeline.Flags().StringVarP(&image, "image", "i", "", "The image to run the pipeline's jobs in.")
	createPipeline.Flags().IntVarP(&shards, "shards", "s", 1, "The sharding factor for the pipeline's jobs.")

	inspectPipeline := &cobra.Command{
		Use:   "inspect-pipeline pipeline-name",
		Short: "Return info about a pipeline.",
		Long:  "Return info about a pipeline.",
		Run: pkgcobra.RunFixedArgs(1, func(args []string) error {
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			pipelineInfo, err := apiClient.InspectPipeline(
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
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			pipelineInfos, err := apiClient.ListPipeline(
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
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			if _, err := apiClient.DeletePipeline(
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

	var result []*cobra.Command
	result = append(result, createJob)
	result = append(result, inspectJob)
	result = append(result, listJob)
	result = append(result, createPipeline)
	result = append(result, inspectPipeline)
	result = append(result, listPipeline)
	result = append(result, deletePipeline)
	return result, nil
}

func errorAndExit(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s\n", fmt.Sprintf(format, args...))
	os.Exit(1)
}

func getAPIClient(address string) (pps.APIClient, error) {
	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return pps.NewAPIClient(clientConn), nil
}
