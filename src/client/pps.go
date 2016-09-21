package client

import (
	"io"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"

	protostream "go.pedge.io/proto/stream"
)

// NewJob creates a pps.Job.
func NewJob(jobID string) *pps.Job {
	return &pps.Job{ID: jobID}
}

var (
	// MapMethod defines a pps.Method for mapper pipelines.
	MapMethod = &pps.Method{
		Partition:   pps.Partition_BLOCK,
		Incremental: pps.Incremental_DIFF,
	}
	// ReduceMethod defines a pps.Method for non-incremental reducer pipelines.
	ReduceMethod = &pps.Method{
		Partition:   pps.Partition_FILE,
		Incremental: pps.Incremental_NONE,
	}
	// IncrementalReduceMethod defines a pps.Method for incremental reducer pipelines.
	IncrementalReduceMethod = &pps.Method{
		Partition:   pps.Partition_FILE,
		Incremental: pps.Incremental_DIFF,
	}
	// GlobalMethod defines a pps.Method for non-incremental, non-partitioned pipelines.
	GlobalMethod = &pps.Method{
		Partition:   pps.Partition_REPO,
		Incremental: pps.Incremental_NONE,
	}
	// DefaultMethod defines the default pps.Method for a pipeline.
	DefaultMethod = MapMethod
	// MethodAliasMap maps a string to a pps.Method for JSON decoding.
	MethodAliasMap = map[string]*pps.Method{
		"map":                MapMethod,
		"reduce":             ReduceMethod,
		"incremental_reduce": IncrementalReduceMethod,
		"global":             GlobalMethod,
	}
	// ReservedRepoNames defines a set of reserved repo names for internal use.
	ReservedRepoNames = map[string]bool{
		"out":  true,
		"prev": true,
	}
)

// NewJobInput creates a pps.JobInput.
func NewJobInput(repoName string, commitID string, method *pps.Method) *pps.JobInput {
	return &pps.JobInput{
		Commit: NewCommit(repoName, commitID),
		Method: method,
	}
}

// NewPipeline creates a pps.Pipeline.
func NewPipeline(pipelineName string) *pps.Pipeline {
	return &pps.Pipeline{Name: pipelineName}
}

// NewPipelineInput creates a new pps.PipelineInput
func NewPipelineInput(repoName string, method *pps.Method) *pps.PipelineInput {
	return &pps.PipelineInput{
		Repo:   NewRepo(repoName),
		Method: method,
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
func (c APIClient) CreateJob(
	image string,
	cmd []string,
	stdin []string,
	parallelismSpec *pps.ParallelismSpec,
	inputs []*pps.JobInput,
	parentJobID string,
) (*pps.Job, error) {
	var parentJob *pps.Job
	if parentJobID != "" {
		parentJob = NewJob(parentJobID)
	}
	job, err := c.PpsAPIClient.CreateJob(
		c.ctx(),
		&pps.CreateJobRequest{
			Transform: &pps.Transform{
				Image: image,
				Cmd:   cmd,
				Stdin: stdin,
			},
			ParallelismSpec: parallelismSpec,
			Inputs:          inputs,
			ParentJob:       parentJob,
		},
	)
	return job, sanitizeErr(err)
}

// InspectJob returns info about a specific job.
// blockOutput will cause the call to block until the job has been assigned an output commit.
// blockState will cause the call to block until the job reaches a terminal state (failure or success).
func (c APIClient) InspectJob(jobID string, blockState bool) (*pps.JobInfo, error) {
	jobInfo, err := c.PpsAPIClient.InspectJob(
		c.ctx(),
		&pps.InspectJobRequest{
			Job:        NewJob(jobID),
			BlockState: blockState,
		})
	return jobInfo, sanitizeErr(err)
}

// ListJob returns info about all jobs.
// If pipelineName is non empty then only jobs that were started by the named pipeline will be returned
// If inputCommit is non-nil then only jobs which took the specific commits as inputs will be returned.
// The order of the inputCommits doesn't matter.
func (c APIClient) ListJob(pipelineName string, inputCommit []*pfs.Commit) ([]*pps.JobInfo, error) {
	var pipeline *pps.Pipeline
	if pipelineName != "" {
		pipeline = NewPipeline(pipelineName)
	}
	jobInfos, err := c.PpsAPIClient.ListJob(
		c.ctx(),
		&pps.ListJobRequest{
			Pipeline:    pipeline,
			InputCommit: inputCommit,
		})
	if err != nil {
		return nil, sanitizeErr(err)
	}
	return jobInfos.JobInfo, nil
}

// GetLogs gets logs from a job (logs includes stdout and stderr).
func (c APIClient) GetLogs(
	jobID string,
	writer io.Writer,
) error {
	getLogsClient, err := c.PpsAPIClient.GetLogs(
		c.ctx(),
		&pps.GetLogsRequest{
			Job: NewJob(jobID),
		},
	)
	if err != nil {
		return sanitizeErr(err)
	}
	return sanitizeErr(protostream.WriteFromStreamingBytesClient(getLogsClient, writer))
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
// update indicates that you want to update an existing pipeline
func (c APIClient) CreatePipeline(
	name string,
	image string,
	cmd []string,
	stdin []string,
	parallelismSpec *pps.ParallelismSpec,
	inputs []*pps.PipelineInput,
	update bool,
) error {
	_, err := c.PpsAPIClient.CreatePipeline(
		c.ctx(),
		&pps.CreatePipelineRequest{
			Pipeline: NewPipeline(name),
			Transform: &pps.Transform{
				Image: image,
				Cmd:   cmd,
				Stdin: stdin,
			},
			ParallelismSpec: parallelismSpec,
			Inputs:          inputs,
			Update:          update,
		},
	)
	return sanitizeErr(err)
}

// InspectPipeline returns info about a specific pipeline.
func (c APIClient) InspectPipeline(pipelineName string) (*pps.PipelineInfo, error) {
	pipelineInfo, err := c.PpsAPIClient.InspectPipeline(
		c.ctx(),
		&pps.InspectPipelineRequest{
			Pipeline: NewPipeline(pipelineName),
		},
	)
	return pipelineInfo, sanitizeErr(err)
}

// ListPipeline returns info about all pipelines.
func (c APIClient) ListPipeline() ([]*pps.PipelineInfo, error) {
	pipelineInfos, err := c.PpsAPIClient.ListPipeline(
		c.ctx(),
		&pps.ListPipelineRequest{},
	)
	if err != nil {
		return nil, sanitizeErr(err)
	}
	return pipelineInfos.PipelineInfo, nil
}

// DeletePipeline deletes a pipeline along with its output Repo.
func (c APIClient) DeletePipeline(name string) error {
	_, err := c.PpsAPIClient.DeletePipeline(
		c.ctx(),
		&pps.DeletePipelineRequest{
			Pipeline: NewPipeline(name),
		},
	)
	return sanitizeErr(err)
}

// StartPipeline restarts a stopped pipeline.
func (c APIClient) StartPipeline(name string) error {
	_, err := c.PpsAPIClient.StartPipeline(
		c.ctx(),
		&pps.StartPipelineRequest{
			Pipeline: NewPipeline(name),
		},
	)
	return sanitizeErr(err)
}

// StopPipeline prevents a pipeline from processing things, it can be restarted
// with StartPipeline.
func (c APIClient) StopPipeline(name string) error {
	_, err := c.PpsAPIClient.StopPipeline(
		c.ctx(),
		&pps.StopPipelineRequest{
			Pipeline: NewPipeline(name),
		},
	)
	return sanitizeErr(err)
}
