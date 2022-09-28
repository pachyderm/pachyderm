package pfsload

import (
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
)

func Commit(pachClient *client.APIClient, taskService task.Service, project, repo, branch string, spec *CommitSpec, seed int64) error {
	env, err := NewEnv(pachClient, taskService, spec, seed)
	if err != nil {
		return err
	}
	for i := 0; i < int(spec.Count); i++ {
		commit, err := pachClient.StartProjectCommit(project, repo, branch)
		if err != nil {
			return err
		}
		for _, mod := range spec.Modifications {
			if err := Modification(env, project, repo, branch, commit.ID, mod); err != nil {
				return err
			}
		}
		if err := pachClient.FinishProjectCommit(project, repo, branch, commit.ID); err != nil {
			return err
		}
		validator := env.Validator()
		if validator != nil {
			if err := validator.Validate(env.Client(), commit); err != nil {
				return err
			}
		}
	}
	return nil
}
