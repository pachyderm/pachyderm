package ppsload

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

func Pipeline(pachClient *client.APIClient, req *pps.RunLoadTestRequest) (*pps.RunLoadTestResponse, error) {
	var branch *pfs.Branch
	var stateID string
	if req.StateId == "" {
		var err error
		branch, err = createPipelines(pachClient, req.DagSpec, req.Parallelism, req.PodPatch)
		if err != nil {
			return nil, err
		}
	} else {
		state, err := deserializeState(pachClient, req.StateId)
		if err != nil {
			return nil, err
		}
		branch = state.Branch
		stateID = state.PfsStateId
	}
	resp, err := pachClient.PfsAPIClient.RunLoadTest(pachClient.Ctx(), &pfs.RunLoadTestRequest{
		Spec:    req.LoadSpec,
		Branch:  branch,
		Seed:    req.Seed,
		StateId: stateID,
	})
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	stateID, err = serializeState(pachClient, &State{
		Branch:     branch,
		PfsStateId: resp.StateId,
	})
	if err != nil {
		return nil, err
	}
	return &pps.RunLoadTestResponse{
		Error:   resp.Error,
		StateId: stateID,
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
		projectName, repoName := parseRepoName(split[0], namespace)
		// Create source repos.
		if strings.TrimSpace(split[1]) == "" {
			if err := pachClient.CreateProjectRepo(projectName, repoName); err != nil {
				return nil, err
			}
			// First source repo will be the target of the PFS load test.
			if retBranch == nil {
				retBranch = client.NewProjectBranch(projectName, repoName, "master")
			}
		} else {
			// Create pipelines.
			inputs := strings.Split(split[1], ",")
			var repoInputs = make([]*pfs.Repo, 0)
			for _, input := range inputs {
				inputProjectName, inputRepoName := parseRepoName(input, namespace)
				repoInputs = append(repoInputs, &pfs.Repo{Project: &pfs.Project{Name: inputProjectName}, Name: inputRepoName})
			}
			repo := &pfs.Repo{Project: &pfs.Project{Name: projectName}, Name: repoName}
			if err := createPipeline(pachClient, repo, repoInputs, parallelism, podPatch); err != nil {
				return nil, err
			}
		}
	}
	return retBranch, nil
}

func createPipeline(pachClient *client.APIClient, repo *pfs.Repo, inputRepos []*pfs.Repo, parallelism int64, podPatch string) error {
	var inputs []*pps.Input
	for i, inputRepo := range inputRepos {
		inputs = append(inputs, &pps.Input{
			Pfs: &pps.PFSInput{
				Name:      fmt.Sprint("input-", i),
				Project:   inputRepo.GetProject().GetName(),
				Repo:      inputRepo.GetName(),
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
			Pipeline: &pps.Pipeline{Project: repo.Project, Name: repo.Name},
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

func parseRepoName(projectRepoName string, namespace string) (projectName string, repoName string) {
	if strings.Count(projectRepoName, "/") == 1 {
		nameSplit := strings.Split(projectRepoName, "/")
		projectName = strings.TrimSpace(nameSplit[0])
		repoName = strings.TrimSpace(nameSplit[1]) + namespace
	} else {
		projectName = pfs.DefaultProjectName
		repoName = strings.TrimSpace(projectRepoName) + namespace
	}
	return projectName, repoName
}

const stateFileName = "state"

func serializeState(pachClient *client.APIClient, state *State) (string, error) {
	resp, err := pachClient.WithCreateFileSetClient(func(mf client.ModifyFile) error {
		data, err := proto.Marshal(state)
		if err != nil {
			return errors.EnsureStack(err)
		}
		return errors.EnsureStack(mf.PutFile(stateFileName, bytes.NewReader(data)))
	})
	if err != nil {
		return "", err
	}
	return resp.FileSetId, nil
}

func deserializeState(pachClient *client.APIClient, stateID string) (*State, error) {
	commit := client.NewProjectRepo(pfs.DefaultProjectName, client.FileSetsRepoName).NewCommit("", stateID)
	buf := &bytes.Buffer{}
	if err := pachClient.GetFile(commit, stateFileName, buf); err != nil {
		return nil, err
	}
	state := &State{}
	if err := proto.Unmarshal(buf.Bytes(), state); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return state, nil
}
