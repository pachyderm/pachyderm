package pfsload

import (
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

func Commit(pachClient *client.APIClient, taskService task.Service, branch *pfs.Branch, spec *CommitSpec, seed int64) error {
	env, err := NewEnv(pachClient, taskService, spec, seed)
	if err != nil {
		return err
	}
	project := branch.Repo.Project.GetName()
	repo := branch.Repo.Name
	for i := 0; i < int(spec.Count); i++ {
		commit, err := pachClient.StartProjectCommit(project, repo, branch.Name)
		if err != nil {
			return err
		}
		for _, mod := range spec.Modifications {
			if err := Modification(env, commit, mod); err != nil {
				return err
			}
		}
		if err := pachClient.FinishProjectCommit(project, repo, branch.Name, commit.ID); err != nil {
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
