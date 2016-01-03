package ppsutil

import (
	// "io"
	// "io/ioutil"

	// "go.pedge.io/proto/stream"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pps"
	"golang.org/x/net/context"
)

func NewJob(jobID string) *pps.Job {
	return &pps.Job{Id: jobID}
}

func NewPipeline(pipelineName string) *pps.Pipeline {
	return &pps.Pipeline{Name: pipelineName}
}

func CreateJob(
	client pps.APIClient,
	image string,
	cmd []string,
	stdin string,
	shards uint64,
	inputCommit []*pfs.Commit,
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
			Shards:      shards,
			InputCommit: inputCommit,
			ParentJob:   parentJob,
		},
	)
}

func CreatePipeline(
	client pps.APIClient,
	name string,
	image string,
	cmd []string,
	stdin string,
	shards uint64,
	inputRepo []*pfs.Repo,
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
			Shards:    shards,
			InputRepo: inputRepo,
		},
	)
	return err
}
