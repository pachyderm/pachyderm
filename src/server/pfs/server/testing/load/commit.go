package load

import (
	"github.com/pachyderm/pachyderm/v2/src/client"
)

type CommitsSpec struct {
	Count           int               `yaml:"count,omitempty"`
	OperationsSpecs []*OperationsSpec `yaml:"operations,omitempty"`
	ThroughputSpec  *ThroughputSpec   `yaml:"throughput,omitempty"`
	CancelSpec      *CancelSpec       `yaml:"cancel,omitempty"`
	ValidatorSpec   *ValidatorSpec    `yaml:"validator,omitempty"`
}

func Commits(pachClient *client.APIClient, repo, branch string, spec *CommitsSpec) error {
	env := NewEnv(NewPachClient(pachClient), spec)
	for i := 0; i < spec.Count; i++ {
		commit, err := pachClient.StartCommit(repo, branch)
		if err != nil {
			return err
		}
		for _, operationsSpec := range spec.OperationsSpecs {
			if err := Operations(env, repo, commit.ID, operationsSpec); err != nil {
				return err
			}
		}
		if err := pachClient.FinishCommit(repo, commit.ID); err != nil {
			return err
		}
		if err := env.Validate(repo, commit.ID); err != nil {
			return err
		}
	}
	return nil
}
