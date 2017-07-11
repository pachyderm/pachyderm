package example

import (
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
		ImagePullSecrets: []string{"my-secret"},
	}
	// CreateJobRequest example
	CreateJobRequest = &ppsclient.CreateJobRequest{
		Transform: Transform,
		ParallelismSpec: &ppsclient.ParallelismSpec{
			Constant: 1,
		},
		Inputs: []*ppsclient.JobInput{
			{
				Commit: &pfs.Commit{
					Repo: &pfs.Repo{Name: "in_repo"},
					ID:   "10cf676b626044f9a405235bf7660959",
				},
				Glob: "*",
				Lazy: true,
			},
		},
	}
	// CreatePipelineRequest example
	CreatePipelineRequest = &ppsclient.CreatePipelineRequest{
		Pipeline: &ppsclient.Pipeline{
			Name: "name",
		},
		Transform: Transform,
		ParallelismSpec: &ppsclient.ParallelismSpec{
			Constant: 1,
		},
		Input: &ppsclient.Input{
			Atom: &ppsclient.AtomInput{
				Repo: "in_repo",
				Glob: "*",
			},
		},
	}
)
