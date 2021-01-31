package load

import "github.com/pachyderm/pachyderm/v2/src/client"

type CommitsSpec struct {
	Count           int               `yaml:"count,omitempty"`
	OperationsSpecs []*OperationsSpec `yaml:"operations,omitempty"`
	ValidatorSpec   *ValidatorSpec    `yaml:"validator,omitempty"`
}

func Commits(c *client.APIClient, repo, branch string, spec *CommitsSpec) error {
	var validator *Validator
	if spec.ValidatorSpec != nil {
		validator = NewValidator(spec.ValidatorSpec)
	}
	for i := 0; i < spec.Count; i++ {
		commit, err := c.StartCommit(repo, branch)
		if err != nil {
			return err
		}
		for _, operationsSpec := range spec.OperationsSpecs {
			if err := Operations(c, repo, commit.ID, operationsSpec, validator); err != nil {
				return err
			}
		}
		if err := c.FinishCommit(repo, commit.ID); err != nil {
			return err
		}
		if validator != nil {
			if err := validator.Validate(c, repo, commit.ID); err != nil {
				return err
			}
		}
	}
	return nil
}
