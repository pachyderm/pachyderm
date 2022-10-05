package ppsload

import (
	"fmt"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

func Pipeline(pachClient *client.APIClient, req *pps.RunLoadTestRequest) (*pps.RunLoadTestResponse, error) {
	branch, err := createPipelines(pachClient, req.DagSpec, req.Parallelism, req.PodPatch)
	if err != nil {
		return nil, err
	}
	resp, err := pachClient.PfsAPIClient.RunLoadTest(pachClient.Ctx(), &pfs.RunLoadTestRequest{
		Spec:   req.LoadSpec,
		Branch: branch,
		Seed:   req.Seed,
	})
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &pps.RunLoadTestResponse{
		Error: resp.Error,
	}, nil
}

func createPipelines(pachClient *client.APIClient, spec string, parallelism int64, podPatch string) (*pfs.Branch, error) {
	namespace := "-" + uuid.NewWithoutDashes()[:5]
	var retBranch *pfs.Branch
	for _, pipelineStr := range strings.Split(spec, "\n") {
		if strings.TrimSpace(pipelineStr) == "" {
			continue
		}
		split := strings.Split(pipelineStr, ":")
		repo := strings.TrimSpace(split[0]) + namespace
		// Create source repos.
		if strings.TrimSpace(split[1]) == "" {
			if err := pachClient.CreateProjectRepo(pfs.DefaultProjectName, repo); err != nil {
				return nil, err
			}
			// First source repo will be the target of the PFS load test.
			if retBranch == nil {
				retBranch = client.NewProjectBranch(pfs.DefaultProjectName, repo, "master")
			}
			continue
		}
		// Create pipelines.
		inputs := strings.Split(split[1], ",")
		for i := range inputs {
			inputs[i] = strings.TrimSpace(inputs[i]) + namespace
		}
		project := "" // TODO(CORE-1063): plumb through project somehow
		if err := createPipeline(pachClient, project, repo, inputs, parallelism, podPatch); err != nil {
			return nil, err
		}
	}
	return retBranch, nil
}

func createPipeline(pachClient *client.APIClient, project, repo string, inputRepos []string, parallelism int64, podPatch string) error {
	var inputs []*pps.Input
	for i, inputRepo := range inputRepos {
		inputs = append(inputs, &pps.Input{
			Pfs: &pps.PFSInput{
				Name:      fmt.Sprint("input-", i),
				Repo:      inputRepo,
				Branch:    "master",
				Glob:      "/(*)",
				JoinOn:    "$1",
				OuterJoin: true,
			},
		})
	}
	_, err := pachClient.PpsAPIClient.CreatePipeline(
		pachClient.Ctx(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewProjectPipeline(project, repo),
			Transform: &pps.Transform{
				Cmd:   []string{"bash"},
				Stdin: []string{"cp -r /pfs/input-*/* /pfs/out/"},
			},
			ParallelismSpec: &pps.ParallelismSpec{Constant: uint64(parallelism)},
			Input:           &pps.Input{Join: inputs},
			PodPatch:        podPatch,
		},
	)
	return errors.EnsureStack(err)
}
