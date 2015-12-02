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
	c pps.APIClient,
	image string,
	cmd []string,
	shards uint64,
	inputCommit []*pfs.Commit,
	outputParent *pfs.Commit,
) (*pps.Job, error) {
	return c.CreateJob(
		context.Background(),
		&pps.CreateJobRequest{
			Spec: &pps.CreateJobRequest_Transform{
				Transform: &pps.Transform{
					Image: image,
					Cmd:   cmd,
				},
			},
			Shards:       shards,
			InputCommit:  inputCommit,
			OutputParent: outputParent,
		},
	)
}
