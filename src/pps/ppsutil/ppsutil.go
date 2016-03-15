package ppsutil

import (
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	"golang.org/x/net/context"
)

func NewJob(jobID string) *ppsclient.Job {
	return &ppsclient.Job{ID: jobID}
}

func NewPipeline(pipelineName string) *ppsclient.Pipeline {
	return &ppsclient.Pipeline{Name: pipelineName}
}

func CreateJob(
	client ppsclient.APIClient,
	image string,
	cmd []string,
	stdin []string,
	shards uint64,
	inputs []*ppsclient.JobInput,
	parentJobID string,
) (*ppsclient.Job, error) {
	var parentJob *ppsclient.Job
	if parentJobID != "" {
		parentJob = NewJob(parentJobID)
	}
	return client.CreateJob(
		context.Background(),
		&ppsclient.CreateJobRequest{
			Transform: &ppsclient.Transform{
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
	client ppsclient.APIClient,
	name string,
	image string,
	cmd []string,
	stdin []string,
	shards uint64,
	inputs []*ppsclient.PipelineInput,
) error {
	_, err := client.CreatePipeline(
		context.Background(),
		&ppsclient.CreatePipelineRequest{
			Pipeline: NewPipeline(name),
			Transform: &ppsclient.Transform{
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
