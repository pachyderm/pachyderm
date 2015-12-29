package ppsutil

import (
	// "io"
	// "io/ioutil"

	// "go.pedge.io/proto/stream"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pps"
	"golang.org/x/net/context"
)

func CreateJob(
	client pps.APIClient,
	image string,
	cmd []string,
	stdin string,
	shards uint64,
	inputCommit []*pfs.Commit,
	parentJob *pps.Job,
) (*pps.Job, error) {
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
	outputRepo *pfs.Repo,
) error {
	_, err := client.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: &pps.Pipeline{
				Name: name,
			},
			Transform: &pps.Transform{
				Image: image,
				Cmd:   cmd,
				Stdin: stdin,
			},
			Shards:     shards,
			InputRepo:  inputRepo,
			OutputRepo: outputRepo,
		},
	)
	return err
}
