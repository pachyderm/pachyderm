package pfsload

import (
	"bytes"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"google.golang.org/protobuf/proto"
)

func Commit(pachClient *client.APIClient, taskService task.Service, branch *pfs.Branch, spec *CommitSpec, seed int64, stateID string) (string, error) {
	env, err := NewEnv(pachClient, taskService, spec, seed)
	if err != nil {
		return "", err
	}
	state := &State{}
	if stateID != "" {
		state, err = deserializeState(pachClient, stateID)
		if err != nil {
			return "", err
		}
		if err := validateState(env, state); err != nil {
			return "", err
		}
	}
	project := branch.Repo.Project.GetName()
	repo := branch.Repo.Name
	for i := 0; i < int(spec.Count); i++ {
		commit, err := pachClient.StartCommit(project, repo, branch.Name)
		if err != nil {
			return "", err
		}
		for _, mod := range spec.Modifications {
			if err := Modification(env, commit, mod); err != nil {
				return "", err
			}
		}
		if err := pachClient.FinishCommit(project, repo, branch.Name, commit.Id); err != nil {
			return "", err
		}
		validator := env.Validator()
		if validator != nil {
			if err := validator.Validate(env.Client(), commit); err != nil {
				return "", err
			}
			state.Commits = append(state.Commits, &State_Commit{
				Commit: commit,
				Hash:   validator.Hash(),
			})
		}
	}
	return serializeState(pachClient, state)
}

func validateState(env *Env, state *State) error {
	validator := env.Validator()
	if validator == nil {
		return nil
	}
	for _, commit := range state.Commits {
		validator.SetHash(commit.Hash)
		if err := validator.Validate(env.Client(), commit.Commit); err != nil {
			return err
		}
	}
	return nil
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
	commit := client.NewRepo(pfs.DefaultProjectName, client.FileSetsRepoName).NewCommit("", stateID)
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
