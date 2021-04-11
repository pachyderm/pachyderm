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
		if err := FuzzOperation(env, repo, commit, spec.FuzzOperationSpecs); err != nil {
			return err
		}
	}
	return nil
}

// TODO: Add different types of operations.
type OperationSpec struct {
	PutFileSpec    *PutFileSpec    `yaml:"putFile,omitempty"`
	DeleteFileSpec *DeleteFileSpec `yaml:"deleteFile,omitempty"`
}

func Operation(env *Env, repo, commit string, spec *OperationSpec) error {
	if spec.PutFileSpec != nil {
		return PutFile(env, repo, commit, spec.PutFileSpec)
	}
	return DeleteFile(env, repo, commit, spec.DeleteFileSpec)
}

type PutFileSpec struct {
	FilesSpec *FilesSpec `yaml:"files,omitempty"`
}

func PutFile(env *Env, repo, commit string, spec *PutFileSpec) error {
	c := env.Client()
	files, err := Files(env, spec.FilesSpec)
	if err != nil {
		return err
	}
	return c.WithModifyFileClient(context.Background(), repo, commit, func(mf client.ModifyFile) error {
		for _, file := range files {
			if err := mf.PutFile(file.Path(), file.Reader()); err != nil {
				return err
			}
		}
		return nil
	})
}

type DeleteFileSpec struct {
	Count int `yaml:"count,omitempty"`
}

func DeleteFile(env *Env, repo, commit string, spec *DeleteFileSpec) error {
	c := env.Client()
	return c.WithModifyFileClient(context.Background(), repo, commit, func(mf client.ModifyFile) error {
		validator := env.Validator()
		for i := 0; i < spec.Count; i++ {
			if err := mf.DeleteFile(validator.RandomFile()); err != nil {
				return err
			}
		}
		return nil
	})
}
