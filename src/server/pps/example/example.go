package example

import (
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
)

var (
	// Secret example
	Secret = &ppsclient.Secret{
		Name:      "secret_name",
		MountPath: "/path/in/container",
	}
	// Transform example
	Transform = &ppsclient.Transform{
		Cmd:              []string{"cmd", "args..."},
		AcceptReturnCode: []int64{1},
		Env:              map[string]string{"foo": "bar"},
		Secrets:          []*ppsclient.Secret{Secret},
	}
	// CreateJobRequest example
	CreateJobRequest = &ppsclient.CreateJobRequest{
		Transform:   Transform,
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
	// CreatePipelineRequest example
	CreatePipelineRequest = &ppsclient.CreatePipelineRequest{
		Pipeline: &ppsclient.Pipeline{
			Name: "name",
		},
		Transform:   Transform,
		Parallelism: 1,
		Inputs: []*ppsclient.PipelineInput{
			{
				Repo:   &pfs.Repo{Name: "in_repo"},
				Method: client.ReduceMethod,
			},
		},
	}
	RunPipelineSpec = &ppsclient.CreateJobRequest{
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
)
