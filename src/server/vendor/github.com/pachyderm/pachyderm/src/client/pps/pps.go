package pps

import (
	"golang.org/x/net/context"
)

func NewJob(jobID string) *Job {
	return &Job{ID: jobID}
}

func NewPipeline(pipelineName string) *Pipeline {
	return &Pipeline{Name: pipelineName}
}

func CreateJob(
	client APIClient,
	image string,
	cmd []string,
	stdin []string,
	shards uint64,
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
			Shards:    shards,
			Inputs:    inputs,
			ParentJob: parentJob,
		},
	)
}

func CreatePipeline(
	client APIClient,
	name string,
	image string,
	cmd []string,
	stdin []string,
	shards uint64,
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
			Shards: shards,
			Inputs: inputs,
		},
	)
	return err
}
