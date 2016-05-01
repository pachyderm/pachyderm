package pps

import (
	"io"

	"go.pedge.io/proto/stream"
	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/src/client/pfs"
)

func NewJob(jobID string) *Job {
	return &Job{ID: jobID}
}

type InputType int

const (
	MAP InputType = iota
	REDUCE
)

func NewJobInput(repoName string, commitID string, inputType InputType) *JobInput {
	return &JobInput{
		Commit: pfs.NewCommit(repoName, commitID),
		Reduce: inputType == REDUCE,
	}
}

func NewPipeline(pipelineName string) *Pipeline {
	return &Pipeline{Name: pipelineName}
}

func NewPipelineInput(repoName string, inputType InputType) *PipelineInput {
	return &PipelineInput{
		Repo:   pfs.NewRepo(repoName),
		Reduce: inputType == REDUCE,
	}
}

// CreateJob creates and runs a job in PPS.
// image is the Docker image to run the job in.
// cmd is the command passed to the Docker run invocation.
// NOTE as with Docker cmd is not run inside a shell that means that things
// like wildcard globbing (*), pipes (|) and file redirects (> and >>) will not
// work. To get that behavior you should have your command be a shell of your
// choice and pass a shell script to stdin.
// stdin is a slice of lines that are sent to your command on stdin. Lines need
// not end in newline characters.
// parallelism is how many copies of your container should run in parallel. You
// may pass 0 for parallelism in which case PPS will set the parallelism based
// on availabe resources.
// inputs specifies a set of Commits that will be visible to the job during runtime.
// parentJobID specifies the a job to use as a parent, it may be left empty in
// which case there is no parent job. If not left empty your job will use the
// parent Job's output commit as the parent of its output commit.
func CreateJob(
	client APIClient,
	image string,
	cmd []string,
	stdin []string,
	parallelism uint64,
	inputs []*JobInput,
	parentJobID string,
) (*Job, error) {
	var parentJob *Job
	if parentJobID != "" {
		parentJob = NewJob(parentJobID)
	}
	return client.CreateJob(
		context.Background(),
		&CreateJobRequest{
			Transform: &Transform{
				Image: image,
				Cmd:   cmd,
				Stdin: stdin,
			},
			Parallelism: parallelism,
			Inputs:      inputs,
			ParentJob:   parentJob,
		},
	)
}

// InspectJob returns info about a specific job.
// blockOutput will cause the call to block until the job has been assigned an output commit.
// blockState will cause the call to block until the job reaches a terminal state (failure or success).
func InspectJob(client APIClient, jobID string, blockOutput bool, blockState bool) (*JobInfo, error) {
	return client.InspectJob(
		context.Background(),
		&InspectJobRequest{
			Job:         NewJob(jobID),
			BlockOutput: blockOutput,
			BlockState:  blockState,
		})
}

// ListJob returns info about all jobs.
// If pipelineName is non empty then only jobs that were started by the named pipeline will be returned
// If inputCommit is non-nil then only jobs which took the specific commits as inputs will be returned.
// The order of the inputCommits doesn't matter.
func ListJob(client APIClient, pipelineName string, inputCommit []*pfs.Commit) ([]*JobInfo, error) {
	jobInfos, err := client.ListJob(
		context.Background(),
		&ListJobRequest{
			Pipeline:    NewPipeline(pipelineName),
			InputCommit: inputCommit,
		})
	if err != nil {
		return nil, err
	}
	return jobInfos.JobInfo, nil
}

// GetLogs gets logs from a job (logs includes stdout and stderr).
func GetLogs(
	client APIClient,
	jobID string,
	writer io.Writer,
) error {
	getLogsClient, err := client.GetLogs(
		context.Background(),
		&GetLogsRequest{
			Job: NewJob(jobID),
		},
	)
	if err != nil {
		return err
	}
	return protostream.WriteFromStreamingBytesClient(getLogsClient, writer)
}

// CreatePipeline creates a new pipeline, pipelines are the main computation
// object in PPS they create a flow of data from a set of input Repos to an
// output Repo (which has the same name as the pipeline). Whenever new data is
// committed to one of the input repos the pipelines will create jobs to bring
// the output Repo up to data.
// image is the Docker image to run the jobs in.
// cmd is the command passed to the Docker run invocation.
// NOTE as with Docker cmd is not run inside a shell that means that things
// like wildcard globbing (*), pipes (|) and file redirects (> and >>) will not
// work. To get that behavior you should have your command be a shell of your
// choice and pass a shell script to stdin.
// stdin is a slice of lines that are sent to your command on stdin. Lines need
// not end in newline characters.
// parallelism is how many copies of your container should run in parallel. You
// may pass 0 for parallelism in which case PPS will set the parallelism based
// on availabe resources.
// inputs specifies a set of Repos that will be visible to the jobs during runtime.
// commits to these repos will cause the pipeline to create new jobs to process them.
func CreatePipeline(
	client APIClient,
	name string,
	image string,
	cmd []string,
	stdin []string,
	parallelism uint64,
	inputs []*PipelineInput,
) error {
	_, err := client.CreatePipeline(
		context.Background(),
		&CreatePipelineRequest{
			Pipeline: NewPipeline(name),
			Transform: &Transform{
				Image: image,
				Cmd:   cmd,
				Stdin: stdin,
			},
			Parallelism: parallelism,
			Inputs:      inputs,
		},
	)
	return err
}

// InspectPipeline returns info about a specific pipeline.
func InspectPipeline(client APIClient, pipelineName string) (*PipelineInfo, error) {
	return client.InspectPipeline(
		context.Background(),
		&InspectPipelineRequest{
			Pipeline: NewPipeline(pipelineName),
		},
	)
}

// ListPipeline returns info about all pipelines.
func ListPipeline(client APIClient) ([]*PipelineInfo, error) {
	pipelineInfos, err := client.ListPipeline(
		context.Background(),
		&ListPipelineRequest{},
	)
	if err != nil {
		return nil, err
	}
	return pipelineInfos.PipelineInfo, nil
}

// DeletePipeline deletes a pipeline along with its output Repo.
func DeletePipeline(client APIClient, name string) error {
	_, err := client.DeletePipeline(
		context.Background(),
		&DeletePipelineRequest{
			Pipeline: NewPipeline(name),
		},
	)
	return err
}
