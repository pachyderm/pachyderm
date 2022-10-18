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
		repo := strings.TrimSpace(split[0]) + namespace
		// Create source repos.
		if strings.TrimSpace(split[1]) == "" {
			if err := pachClient.CreateRepo(repo); err != nil {
				return nil, err
			}
			// First source repo will be the target of the PFS load test.
			if retBranch == nil {
				retBranch = client.NewBranch(repo, "master")
			}
			continue
		}
		// Create pipelines.
		inputs := strings.Split(split[1], ",")
		for i := range inputs {
			inputs[i] = strings.TrimSpace(inputs[i]) + namespace
		}
		if err := createPipeline(pachClient, repo, inputs, parallelism, podPatch); err != nil {
			return nil, err
		}
	}
	return retBranch, nil
}

func createPipeline(pachClient *client.APIClient, repo string, inputRepos []string, parallelism int64, podPatch string) error {
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
			Pipeline: client.NewPipeline(repo),
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
	commit := client.NewRepo(client.FileSetsRepoName).NewCommit("", stateID)
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
