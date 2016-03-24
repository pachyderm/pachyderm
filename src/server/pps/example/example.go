package example

import (
	"github.com/pachyderm/pachyderm/src/client/pfs"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
)

func CreateJobRequest() *ppsclient.CreateJobRequest {
	return &ppsclient.CreateJobRequest{
		Transform: &ppsclient.Transform{
			Cmd: []string{"cmd", "args..."},
		},
		Shards: 1,
		Inputs: []*ppsclient.JobInput{
			{
				Commit: &pfs.Commit{
					Repo: &pfs.Repo{Name: "in_repo"},
					ID:   "10cf676b626044f9a405235bf7660959",
				},
			},
		},
		ParentJob: &ppsclient.Job{
			ID: "a951ca06cfda4377b8ffaa050d1074df",
		},
	}
}

func CreatePipelineRequest() *ppsclient.CreatePipelineRequest {
	return &ppsclient.CreatePipelineRequest{
		Pipeline: &ppsclient.Pipeline{
			Name: "name",
		},
		Transform: &ppsclient.Transform{
			Cmd: []string{"cmd", "args..."},
		},
		Shards: 1,
		Inputs: []*ppsclient.PipelineInput{
			{
				Repo: &pfs.Repo{Name: "in_repo"},
			},
		},
	}
}
