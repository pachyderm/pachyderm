package ppsutil

import (
	"github.com/pachyderm/pachyderm/src/pps"
	"golang.org/x/net/context"
)

func NewJob(jobID string) *pps.Job {
	return &pps.Job{ID: jobID}
}

func NewPipeline(pipelineName string) *pps.Pipeline {
	return &pps.Pipeline{Name: pipelineName}
}

func CreateJob(
	client pps.APIClient,
	image string,
	cmd []string,
	stdin []string,
	shards uint64,
	inputs []*pps.JobInput,
	parentJobID string,
) (*pps.Job, error) {
	var parentJob *pps.Job
	if parentJobID != "" {
		parentJob = NewJob(parentJobID)
	}
	return client.CreateJob(
		context.Background(),
		&pps.CreateJobRequest{
			Transform: &pps.Transform{
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
	client pps.APIClient,
	name string,
	image string,
	cmd []string,
	stdin []string,
	shards uint64,
	inputs []*pps.PipelineInput,
) error {
	_, err := client.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: NewPipeline(name),
			Transform: &pps.Transform{
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
