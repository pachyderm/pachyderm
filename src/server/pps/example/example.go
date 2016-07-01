package example

import (
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
)

// CreateJobRequest creates an example job request.
func CreateJobRequest() *ppsclient.CreateJobRequest {
	return &ppsclient.CreateJobRequest{
		Transform: &ppsclient.Transform{
			Cmd:              []string{"cmd", "args..."},
			AcceptReturnCode: []int64{1},
		},
		Parallelism: 1,
		Inputs: []*ppsclient.JobInput{
			{
				Commit: &pfs.Commit{
					Repo: &pfs.Repo{Name: "in_repo"},
					ID:   "10cf676b626044f9a405235bf7660959",
				},
				Method: client.MapMethod,
			},
		},
		ParentJob: &ppsclient.Job{
			ID: "a951ca06cfda4377b8ffaa050d1074df",
		},
	}
}

// CreatePipelineRequest creates an example pipeline request.
func CreatePipelineRequest() *ppsclient.CreatePipelineRequest {
	return &ppsclient.CreatePipelineRequest{
		Pipeline: &ppsclient.Pipeline{
			Name: "name",
		},
		Transform: &ppsclient.Transform{
			Cmd:              []string{"cmd", "args..."},
			AcceptReturnCode: []int64{1},
		},
		Parallelism: 1,
		Inputs: []*ppsclient.PipelineInput{
			{
				Repo:   &pfs.Repo{Name: "in_repo"},
				Method: client.ReduceMethod,
			},
		},
	}
}

// RunPipelineSpec creates an example run pipeline request.
func RunPipelineSpec() *ppsclient.CreateJobRequest {
	return &ppsclient.CreateJobRequest{
		Inputs: []*ppsclient.JobInput{
			{
				Commit: &pfs.Commit{
					Repo: &pfs.Repo{Name: "in_repo"},
					ID:   "10cf676b626044f9a405235bf7660959",
				},
				Method: client.GlobalMethod,
			},
		},
		Parallelism: 3,
	}
}
