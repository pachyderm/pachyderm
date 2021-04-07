package load

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/client"
)

type OperationsSpec struct {
	Count              int                  `yaml:"count,omitempty"`
	FuzzOperationSpecs []*FuzzOperationSpec `yaml:"fuzzOperations,omitempty"`
}

func Operations(env *Env, repo, commit string, spec *OperationsSpec) error {
	for i := 0; i < spec.Count; i++ {
		return FuzzOperation(env, repo, commit, spec.FuzzOperationSpecs)
	}
	return nil
}

// TODO: Add different types of operations.
type OperationSpec struct {
	PutFileSpec *PutFileSpec `yaml:"putFile,omitempty"`
}

func Operation(env *Env, repo, commit string, spec *OperationSpec) error {
	return PutFile(env, repo, commit, spec.PutFileSpec)
}

type PutFileSpec struct {
	FilesSpec *FilesSpec `yaml:"files,omitempty"`
}

func PutFile(env *Env, repo, commit string, spec *PutFileSpec) error {
	c := env.Client()
	files, err := Files(spec.FilesSpec)
	if err != nil {
		return err
	}
	return c.WithModifyFileClient(context.Background(), repo, commit, func(mfc client.ModifyFileClient) error {
		for _, file := range files {
			if err := mfc.PutFile(file.Path(), file.Reader()); err != nil {
				return err
			}
		}
		return nil
	})
}
